package pizza

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"

	oidc "github.com/coreos/go-oidc"
	mux "github.com/gorilla/mux"
	zap "go.uber.org/zap"
	oauth2 "golang.org/x/oauth2"
)

var EventDuration = time.Hour * 4

type WrappedData struct {
	Friends      map[string][]time.Time `json:"friends"`
	TotalFridays int                    `json:"totalFridays"`
}

type Server struct {
	s        http.Server
	store    Accessor
	calendar *Calendar
	config   Config

	oauth2Provider *oidc.Provider
	oauth2Conf     oauth2.Config
	verifier       *oidc.IDTokenVerifier
	keycloak       *Keycloak

	indexGetMetric      CounterMetric
	submitPostMetric    CounterMetric
	wrappedGetMetric    CounterMetric
	requestErrorMetric  CounterMetric
	internalErrorMetric CounterMetric

	wrapped  map[int]WrappedData
	sessions map[string]*TokenClaims
}

func NewServer(config Config, metricsReg MetricsRegistry) (*Server, error) {
	r := mux.NewRouter()

	var accessor Accessor
	var err error
	Log.Info("using the sqlite accessor")
	accessor, err = NewSQLAccessor(config.DBFile, false)
	if err != nil {
		return nil, err
	}

	googleCal, err := NewGoogleCalendar(config.Calendar.CredentialFile, config.Calendar.TokenFile, config.Calendar.ID, context.Background())
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, config.OAuth2.KeycloakURL+"/realms/"+config.OAuth2.Realm)
	if err != nil {
		return nil, err
	}
	k, err := NewKeycloak(config.OAuth2)
	if err != nil {
		Log.Error("keycloak failure", zap.Error(err))
		return nil, err
	}

	s := Server{
		s: http.Server{
			Addr:         fmt.Sprintf("0.0.0.0:%d", config.Port),
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			Handler:      r,
		},
		store:    accessor,
		calendar: NewCalendar(googleCal),
		config:   config,

		oauth2Provider: provider,
		oauth2Conf: oauth2.Config{
			ClientID:     config.OAuth2.ClientID,
			ClientSecret: config.OAuth2.ClientSecret,
			RedirectURL:  config.OAuth2.RedirectURL + "/login/callback",
			Endpoint:     provider.Endpoint(),
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		},
		verifier: provider.Verifier(&oidc.Config{
			ClientID: config.OAuth2.ClientID,
		}),
		keycloak: k,

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

		wrapped:  map[int]WrappedData{},
		sessions: map[string]*TokenClaims{},
	}

	r.HandleFunc("/", s.HandleIndex)
	r.HandleFunc("/rsvp", s.HandleRSVP)
	r.HandleFunc("/wrapped", s.HandledWrapped)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(config.StaticDir))))
	r.HandleFunc("/login", s.HandleLogin)
	r.HandleFunc("/login/callback", s.HandleLoginCallback)
	r.HandleFunc("/logout", s.HandleLogout)
	r.HandleFunc("/admin", s.HandleAdmin)
	r.HandleFunc("/admin/edit", s.HandleAdminEdit)

	r.HandleFunc("/profile", s.HandleGetProfile)
	r.HandleFunc("/profile/edit", s.HandleUpdateProfile)

	r.HandleFunc("/api/token", s.HandleAPIAuth)
	r.HandleFunc("/api/friday", s.HandleAPIFriday)
	r.HandleFunc("/api/friday/{ID}", s.HandleAPIFriday)

	return &s, nil
}

func (s *Server) Start() error {
	// watch the calendar to keep credentials renewed and learn when they have expired
	go s.WatchCalendar(1 * time.Hour)
	// start the HTTP server
	if err := s.s.ListenAndServe(); err != http.ErrServerClosed {
		Log.Error("http listen error", zap.Error(err))
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
	timer := time.NewTimer(period)
	for {
		if _, err := s.calendar.ListEvents(1); err != nil {
			Log.Warn("failed to list calendar events", zap.Error(err))
		} else {
			Log.Debug("calendar credentials are valid")
		}
		<-timer.C
		timer.Reset(period)
	}
}

type IndexFridayData struct {
	Date      string
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
}

func (s *Server) HandleIndex(w http.ResponseWriter, r *http.Request) {
	s.indexGetMetric.Increment()

	plate, err := template.ParseFiles(path.Join(s.config.StaticDir, "html/index.html"))
	if err != nil {
		Log.Error("template index failure", zap.Error(err))
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

		Log.Info("welcome", zap.String("name", claims.Name))

		if err = s.store.AddFriend(claims.Email, claims.Name); err != nil {
			Log.Warn("failed to add friend", zap.Error(err))
		}

		data.Name = claims.GivenName
		data.LogoutURL = fmt.Sprintf("%s/%s?post_logout_redirect_uri=%s/logout&client_id=%s", s.oauth2Conf.Endpoint.AuthURL, "../logout", s.config.OAuth2.RedirectURL, "pizza")

		fridays, err := s.store.GetUpcomingFridays(30)
		if err != nil {
			Log.Error("failed to get fridays", zap.Error(err))
			s.Handle500(w, r)
			return
		}

		estZone, _ := time.LoadLocation("America/New_York")
		data.FridayTimes = make([]IndexFridayData, 0)
		for _, friday := range fridays {
			if friday.Group != nil && !claims.InGroup(*friday.Group) {
				// skip friday when the user is not in the invited group
				continue
			}

			fData := IndexFridayData{
				MaxGuests: friday.MaxGuests,
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

			eventID := strconv.FormatInt(fData.ID, 10)
			// get the calendar event to see who has already RSVP'ed
			// TODO switch to using the local guest list instead of the calendar
			if event, err := s.calendar.GetEvent(eventID); err != nil && err != ErrEventNotFound {
				Log.Warn("failed to get calendar event", zap.Error(err), zap.String("eventID", eventID))
				fData.Guests = make([]string, 0)
			} else if err != nil {
				fData.Guests = make([]string, 0)
			} else {
				fData.Guests = make([]string, len(event.Attendees))
				for k, email := range event.Attendees {
					if name, err := s.store.GetFriendName(email); err != nil {
						fData.Guests[k] = email
					} else {
						fData.Guests[k] = name
					}
				}
			}
			data.FridayTimes = append(data.FridayTimes, fData)
		}
	}

	if err = plate.Execute(w, data); err != nil {
		Log.Error("template execution failure", zap.Error(err))
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

	Log.Debug("incoming submit request", zap.Stringer("url", r.URL))

	form := r.URL.Query()
	dates, ok := form["date"]
	if !ok {
		template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_fail.html"))).Execute(w, nil)
		return
	}
	email := strings.ToLower(claims.Email)
	Log.Debug("rsvp request", zap.String("email", email), zap.Strings("dates", dates))

	for _, d := range dates {
		num, err := strconv.ParseInt(d, 10, 64)
		if err != nil {
			Log.Error("failed parsing date int from rsvp form", zap.String("date", d))
			template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_fail.html"))).Execute(w, nil)
			return
		}
		if err = s.CreateAndInvite(d, time.Unix(num, 0), email, claims.GivenName); err != nil {
			template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_error.html"))).Execute(w, nil)
			return
		}
	}

	template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_success.html"))).Execute(w, nil)
}

func (s *Server) CreateAndInvite(ID string, startTime time.Time, email, name string) error {
	newEvent := CalendarEvent{
		AnyoneCanAddSelf:      false,
		Description:           "Welcome to Pizza Friday!",
		StartTime:             startTime,
		GuestsCanInviteOthers: false,
		GuestsCanModify:       false,
		Id:                    ID,
		Locked:                true,
		EndTime:               startTime.Add(time.Hour + 5),
		Status:                "confirmed",
		Summary:               "Pizza Friday",
		Visibility:            "private",
	}

	// update local table with new guest list
	estZone, _ := time.LoadLocation("America/New_York")
	friday := Friday{
		Date: startTime.In(estZone),
	}
	if err := s.store.AddFriendToFriday(email, friday); err != nil {
		Log.Error("update to local invite list failed", zap.Error(err))
		return err
	}

	err := s.calendar.InviteToEvent(ID, email, name)
	if err != nil && err == ErrEventNotFound {
		if err = s.calendar.CreateEvent(newEvent); err != nil {
			Log.Error("could not create event", zap.String("eventID", ID), zap.String("email", email), zap.Error(err))
			return err
		}
		err = s.calendar.InviteToEvent(ID, email, name)
	}
	if err != nil {
		Log.Error("invite failed", zap.String("eventID", ID), zap.String("email", email), zap.Error(err))
		return err
	}

	Log.Debug("event updated", zap.String("eventID", ID), zap.String("email", email), zap.String("name", name))
	return nil
}

func (s *Server) Handle4xx(w http.ResponseWriter, r *http.Request) {
	s.requestErrorMetric.Increment()
	plate, err := template.ParseFiles(path.Join(s.config.StaticDir, "html/4xx.html"))
	if err != nil {
		Log.Error("template 4xx failure", zap.Error(err))
		s.Handle500(w, r)
		return
	}
	data := PageData{}
	if err = plate.Execute(w, data); err != nil {
		Log.Error("template execution failure", zap.Error(err))
		s.Handle500(w, r)
		return
	}
}

func (s *Server) Handle500(w http.ResponseWriter, r *http.Request) {
	s.internalErrorMetric.Increment()
	plate, err := template.ParseFiles(path.Join(s.config.StaticDir, "html/500.html"))
	if err != nil {
		Log.Error("template 500 failure", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data := PageData{}
	if err = plate.Execute(w, data); err != nil {
		Log.Error("template execution failure", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
