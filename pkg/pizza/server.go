package pizza

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"slices"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var EventDuration = time.Hour * 4

type WrappedData struct {
	Friends      map[string][]time.Time `json:"friends"`
	TotalFridays int                    `json:"totalFridays"`
}

type Server struct {
	s             http.Server
	config        Config
	store         Accessor
	calendar      Calendar
	authenticator Authenticator

	indexGetMetric      CounterMetric
	submitPostMetric    CounterMetric
	wrappedGetMetric    CounterMetric
	requestErrorMetric  CounterMetric
	internalErrorMetric CounterMetric

	wrapped map[int]WrappedData
}

func NewServer(config Config, accessor Accessor, calendar Calendar, auth Authenticator, metricsReg MetricsRegistry) (*Server, error) {
	mux := http.NewServeMux()

	s := Server{
		s: http.Server{
			Addr:         fmt.Sprintf("0.0.0.0:%d", config.Port),
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			Handler:      mux,
		},
		config:        config,
		store:         accessor,
		calendar:      calendar,
		authenticator: auth,

		indexGetMetric: metricsReg.NewCounterMetric("pizza_requests",
			map[string]string{"method": "get", "path": "/"}),
		submitPostMetric: metricsReg.NewCounterMetric("pizza_requests",
			map[string]string{"method": "post", "path": "/submit"}),
		wrappedGetMetric: metricsReg.NewCounterMetric("pizza_requests",
			map[string]string{"method": "get", "path": "/wrapped"}),
		requestErrorMetric: metricsReg.NewCounterMetric("pizza_errors",
			map[string]string{"statusCode": "4xx"}),
		internalErrorMetric: metricsReg.NewCounterMetric("pizza_errors",
			map[string]string{"statusCode": "500"}),

		wrapped: map[int]WrappedData{},
	}

	s.LoadRoutes(mux)

	return &s, nil
}

func (s *Server) LoadRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", s.HandleIndex)
	mux.HandleFunc("GET /friday/{ID}", s.HandleFridayPartial)
	mux.HandleFunc("POST /rsvp", s.HandleRSVP)
	mux.HandleFunc("DELETE /rsvp", s.HandleDeleteRSVP)
	mux.HandleFunc("GET /wrapped", s.HandledWrapped)
	mux.HandleFunc("GET /login", s.HandleLogin)
	mux.HandleFunc("GET /login/callback", s.HandleLoginCallback)
	mux.HandleFunc("GET /logout", s.HandleLogout)
	mux.HandleFunc("GET /admin", s.HandleAdmin)
	mux.HandleFunc("POST /admin/edit", s.HandleAdminEdit)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(s.config.StaticDir))))

	mux.HandleFunc("GET /profile", s.HandleGetProfile)
	mux.HandleFunc("POST /profile/edit", s.HandleUpdateProfile)

	mux.HandleFunc("POST /api/token", s.HandleAPIAuth)
	mux.HandleFunc("GET /api/friday", s.HandleAPIFriday)
	mux.HandleFunc("GET /api/friday/{ID}", s.HandleAPIFriday)
	mux.HandleFunc("PATCH /api/friday/{ID}", s.HandleAPIFriday)
	mux.HandleFunc("GET /api/guest/{ID}", s.HandleAPIGuest)
	mux.HandleFunc("GET /api/guest/{ID}/profile", s.HandleAPIGuestProfile)

	mux.HandleFunc("GET /p/{ID}", s.HandlePizza)
}

func (s *Server) Start() error {
	// watch the calendar to keep credentials renewed and learn when they have expired
	go s.WatchCalendar(1 * time.Hour)
	// start the HTTP server
	if err := s.s.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("http listen error", "error", err)
		return err
	}
	return nil
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()
	s.s.Shutdown(ctx)
}

func (s *Server) WatchCalendar(period time.Duration) {
	timer := time.NewTimer(0)
	estZone, _ := time.LoadLocation("America/New_York")
	for {
		<-timer.C
		fridays, err := s.store.GetUpcomingFridays(30)
		if err != nil {
			slog.Error("[sync] failed to get upcoming fridays", "err", err)
			timer.Reset(1 * time.Minute)
		}
		for _, friday := range fridays {
			t := friday.Date.In(estZone)
			eventID := strconv.FormatInt(t.Unix(), 10)
			event, err := s.calendar.GetEvent(eventID)
			if err != nil {
				slog.Warn("[sync] failed to get calendar event", "err", err, "eventID", eventID)
			} else {
				for _, attendee := range event.Attendees {
					if attendee.ResponseStatus == "declined" {
						if err = s.store.RemoveFriendFromFriday(attendee.Email, t); err != nil {
							slog.Error("[sync] failed to remove friend from friday after calendar decline", "err", err, "email", attendee.Email, "eventID", eventID)
						}
					}
				}
			}
		}

		slog.Info("[sync] calendar sync complete")
		timer.Reset(period)
	}
}

type IndexFridayData struct {
	Date      string
	ShortDate string
	ID        int64
	Guests    []string
	Active    bool
	Group     string
	Details   string
	IsInvited bool
	MaxGuests int
}

type PageData struct {
	FridayTimes []IndexFridayData
	Name        string
	LoggedIn    bool
	LogoutURL   string
	IsAdmin     bool
	PixelPizza  PixelPizzaPageData
}

func (s *Server) HandleIndex(w http.ResponseWriter, r *http.Request) {
	s.indexGetMetric.Increment()

	plate, err := template.ParseFiles(path.Join(s.config.StaticDir, "html/index.html"))
	if err != nil {
		slog.Error("template index failure", "error", err)
		s.Handle500(w, r)
		return
	}
	if _, err = plate.ParseGlob(path.Join(s.config.StaticDir, "html/snippets/*.html")); err != nil {
		slog.Error("template snippets parse failure", "error", err)
		s.Handle500(w, r)
		return
	}

	data := PageData{
		LoggedIn: false,
	}

	claims, ok := s.authenticateRequest(r)
	if ok {
		data.LoggedIn = true
		data.IsAdmin = claims.HasRole("pizza_host")

		slog.Info("welcome", "name", claims.Name)

		if err = s.store.AddFriend(claims.Email, claims.Name); err != nil {
			slog.Warn("failed to add friend", "error", err)
		}

		data.Name = claims.GivenName
		data.LogoutURL = fmt.Sprintf("%s/%s?post_logout_redirect_uri=%s/logout&client_id=%s", s.authenticator.GetAuthURL(), "../logout", s.config.OAuth2.RedirectURL, "pizza")

		if prefs, err := s.store.GetPreferences(claims.Email); err != nil {
			slog.Error("failed to get user preferences", "email", claims.Email, "err", err)
		} else {
			data.PixelPizza.Pizza = NewPixelPizzaFromPreferences(prefs).Render("darkblue")
			data.PixelPizza.Size = "12px"
		}

		fridays, err := s.store.GetUpcomingFridays(30)
		if err != nil {
			slog.Error("failed to get fridays", "error", err)
			s.Handle500(w, r)
			return
		}

		estZone, _ := time.LoadLocation("America/New_York")
		data.FridayTimes = make([]IndexFridayData, 0)
		for _, friday := range fridays {
			if (friday.Group != nil && !claims.InGroup(*friday.Group)) || !friday.Enabled {
				// skip friday when the user is not in the invited group
				// also skip fridays that are disabled
				continue
			}

			fData := IndexFridayData{
				MaxGuests: friday.MaxGuests,
				ShortDate: fmt.Sprintf("%s %d", friday.Date.Month().String(), friday.Date.Day()),
			}
			t := friday.Date
			t = t.In(estZone)
			fData.Date = t.Format(time.RFC822)
			fData.ID = t.Unix()
			if friday.Details != nil {
				fData.Details = *friday.Details
			}
			// add indicator if guest has already RSVP'ed for this friday
			fData.IsInvited = false
			for _, guest := range friday.Guests {
				if guest == claims.Email {
					fData.IsInvited = true
				}
			}

			// get the calendar event to see who has already RSVP'ed
			fData.Guests = make([]string, len(friday.Guests))
			for k, attendee := range friday.Guests {
				if friend, err := s.store.GetFriendByEmail(attendee); err != nil {
					fData.Guests[k] = attendee
				} else {
					fData.Guests[k] = friend.Name
				}
			}
			data.FridayTimes = append(data.FridayTimes, fData)
		}
	}

	if err = plate.Execute(w, data); err != nil {
		slog.Error("template execution failure", "error", err)
		s.Handle500(w, r)
		return
	}
}

func (s *Server) HandleFridayPartial(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticateRequest(r)
	if !ok {
		template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_fail.html"))).Execute(w, nil)
		return
	}
	plate, err := template.ParseGlob(path.Join(s.config.StaticDir, "html/snippets/*.html"))
	if err != nil {
		slog.Error("template snippets parse failure", "error", err)
		s.Handle500(w, r)
		return
	}
	d := r.PathValue("ID")
	num, err := strconv.ParseInt(d, 10, 64)
	if err != nil {
		slog.Error("failed parsing date int from rsvp form", "date", d)
		template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_fail.html"))).Execute(w, nil)
		return
	}
	estZone, _ := time.LoadLocation("America/New_York")
	friday, err := s.store.GetFriday(time.Unix(num, 0).In(estZone))
	if err != nil {
		// friday does not exist
		slog.Info("friday does not exist", "error", err)
		template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_fail.html"))).Execute(w, nil)
		return
	}

	if (friday.Group != nil && !claims.InGroup(*friday.Group)) || !friday.Enabled {
		// not part of invited group OR friday not enabled
		slog.Info("friday not enabled or claims check")
		template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_fail.html"))).Execute(w, nil)
		return
	}

	fData := IndexFridayData{
		MaxGuests: friday.MaxGuests,
		ShortDate: fmt.Sprintf("%s %d", friday.Date.Month().String(), friday.Date.Day()),
	}
	t := friday.Date
	t = t.In(estZone)
	fData.Date = t.Format(time.RFC822)
	fData.ID = t.Unix()
	if friday.Details != nil {
		fData.Details = *friday.Details
	}
	// add indicator if guest has already RSVP'ed for this friday
	fData.IsInvited = false
	for _, guest := range friday.Guests {
		if guest == claims.Email {
			fData.IsInvited = true
		}
	}

	// get the calendar event to see who has already RSVP'ed
	fData.Guests = make([]string, len(friday.Guests))
	for k, attendee := range friday.Guests {
		if friend, err := s.store.GetFriendByEmail(attendee); err != nil {
			fData.Guests[k] = attendee
		} else {
			fData.Guests[k] = friend.Name
		}
	}

	if err = plate.ExecuteTemplate(w, "SelectedFriday", fData); err != nil {
		slog.Error("template execution failure", "error", err)
		s.Handle500(w, r)
		return
	}
}

func (s *Server) HandleRSVP(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticateRequest(r)
	if !ok {
		template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_fail.html"))).Execute(w, nil)
		return
	}

	slog.Debug("incoming submit request", "url", r.URL)

	form := r.URL.Query()
	dates, ok := form["date"]
	if !ok {
		template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_fail.html"))).Execute(w, nil)
		return
	}
	email := strings.ToLower(claims.Email)
	slog.Debug("rsvp request", "email", email, "dates", dates)

	for _, d := range dates {
		num, err := strconv.ParseInt(d, 10, 64)
		if err != nil {
			slog.Error("failed parsing date int from rsvp form", "date", d)
			template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_fail.html"))).Execute(w, nil)
			return
		}
		estZone, _ := time.LoadLocation("America/New_York")
		friday, err := s.store.GetFriday(time.Unix(num, 0).In(estZone))
		if err != nil {
			// friday does not exist
			slog.Info("friday does not exist", "error", err)
			template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_fail.html"))).Execute(w, nil)
			return
		}

		if (friday.Group != nil && !claims.InGroup(*friday.Group)) || !friday.Enabled {
			// not part of invited group OR friday not enabled
			slog.Info("friday not enabled or claims check")
			template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_fail.html"))).Execute(w, nil)
			return
		}

		if err = s.CreateAndInvite(d, friday, email, claims.GivenName); err != nil {
			template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_error.html"))).Execute(w, nil)
			return
		}
	}

	template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_success.html"))).Execute(w, nil)
}

func (s *Server) CreateAndInvite(ID string, friday Friday, email, name string) error {
	newEvent := CalendarEvent{
		AnyoneCanAddSelf:      false,
		Description:           "Welcome to Pizza Friday!",
		StartTime:             friday.Date,
		GuestsCanInviteOthers: false,
		GuestsCanModify:       false,
		Id:                    ID,
		Locked:                true,
		EndTime:               friday.Date.Add(time.Hour + 5),
		Status:                "confirmed",
		Summary:               "Pizza Friday",
		Visibility:            "private",
	}

	// update local table with new guest list
	if err := s.store.AddFriendToFriday(email, friday); err != nil {
		slog.Error("update to local invite list failed", "error", err)
		return err
	}

	if !s.config.Calendar.Enabled {
		return nil
	}

	err := s.calendar.InviteToEvent(ID, email, name)
	if err != nil && err == ErrEventNotFound {
		if err = s.calendar.CreateEvent(newEvent); err != nil {
			slog.Error("could not create event", "eventID", ID, "email", email, "error", err)
			return err
		}
		err = s.calendar.InviteToEvent(ID, email, name)
	}
	if err != nil {
		slog.Error("invite failed", "eventID", ID, "email", email, "error", err)
		return err
	}

	slog.Debug("event updated", "eventID", ID, "email", email, "name", name)
	return nil
}

func (s *Server) HandleDeleteRSVP(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticateRequest(r)
	if !ok {
		template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_fail.html"))).Execute(w, nil)
		return
	}

	form := r.URL.Query()
	dates, ok := form["date"]
	if !ok {
		template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_fail.html"))).Execute(w, nil)
		return
	}

	slog.Debug("incoming decline request", "url", r.URL, "email", claims.Email, "dates", dates)
	estZone, _ := time.LoadLocation("America/New_York")

	for _, d := range dates {
		fridayID, err := strconv.ParseInt(d, 10, 64)
		if err != nil {
			template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_fail.html"))).Execute(w, nil)
			return
		}

		friday, err := s.store.GetFriday(time.Unix(fridayID, 0).In(estZone))
		if err != nil {
			template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_fail.html"))).Execute(w, nil)
			return
		}

		if (friday.Group != nil && !claims.InGroup(*friday.Group)) || !friday.Enabled {
			// not part of invited group OR friday not enabled
			template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_fail.html"))).Execute(w, nil)
			return
		}

		if slices.Contains(friday.Guests, claims.Email) {
			if err = s.store.RemoveFriendFromFriday(claims.Email, friday.Date); err != nil {
				slog.Error("failed to remove friend from friday", "err", err, "email", claims.Email, "friday", d)
				template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_fail.html"))).Execute(w, nil)
				return
			}
			if s.config.Calendar.Enabled {
				if err = s.calendar.DeclineEvent(d, claims.Email); err != nil {
					slog.Error("failed to decline calendar invite", "err", err, "email", claims.Email, "friday", d)
					template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_fail.html"))).Execute(w, nil)
					return
				}
			}
		}
	}

	template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/decline_success.html"))).Execute(w, nil)
}

type PixelPizzaPageData struct {
	Size  string
	Pizza [][]string
}

func (s *Server) HandlePizza(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("ID")
	pixelPizza, err := NewPixelPizzaFromID(id)
	if err != nil {
		slog.Warn("failed to parse pizza ID", "id", id, "err", err)
		s.Handle4xx(w, r)
		return
	}
	slog.Info("generated pizza", "id", pixelPizza.ID(), "pizza", pixelPizza.String())
	w.Write([]byte(pixelPizza.HTML("#fff;")))
}

func (s *Server) Handle4xx(w http.ResponseWriter, r *http.Request) {
	s.requestErrorMetric.Increment()
	plate, err := template.ParseFiles(path.Join(s.config.StaticDir, "html/4xx.html"))
	if err != nil {
		slog.Error("template 4xx failure", "error", err)
		s.Handle500(w, r)
		return
	}
	data := PageData{}
	if err = plate.Execute(w, data); err != nil {
		slog.Error("template execution failure", "error", err)
		s.Handle500(w, r)
		return
	}
}

func (s *Server) Handle500(w http.ResponseWriter, r *http.Request) {
	s.internalErrorMetric.Increment()
	plate, err := template.ParseFiles(path.Join(s.config.StaticDir, "html/500.html"))
	if err != nil {
		slog.Error("template 500 failure", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data := PageData{}
	if err = plate.Execute(w, data); err != nil {
		slog.Error("template execution failure", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
