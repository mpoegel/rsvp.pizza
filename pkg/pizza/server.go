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
	accessor, err = NewSQLAccessor(config.DBFile)
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
	r.HandleFunc("/admin/submit", s.HandleAdminSubmit)

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
	Date   string
	ID     int64
	Guests []string
	Active bool
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
		for i, friday := range fridays {
			if friday.Group != nil && !claims.InGroup(*friday.Group) {
				// skip friday when the user is not in the invited group
				continue
			}
			data.FridayTimes = append(data.FridayTimes, IndexFridayData{})
			t := friday.Date
			t = t.In(estZone)
			data.FridayTimes[i].Date = t.Format(time.RFC822)
			data.FridayTimes[i].ID = t.Unix()

			eventID := strconv.FormatInt(data.FridayTimes[i].ID, 10)
			if event, err := s.calendar.GetEvent(eventID); err != nil && err != ErrEventNotFound {
				Log.Warn("failed to get calendar event", zap.Error(err), zap.String("eventID", eventID))
				data.FridayTimes[i].Guests = make([]string, 0)
			} else if err != nil {
				data.FridayTimes[i].Guests = make([]string, 0)
			} else {
				data.FridayTimes[i].Guests = make([]string, len(event.Attendees))
				for k, email := range event.Attendees {
					if name, err := s.store.GetFriendName(email); err != nil {
						data.FridayTimes[i].Guests[k] = email
					} else {
						data.FridayTimes[i].Guests[k] = name
					}
				}
			}
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

	friendName := claims.GivenName

	newEvent := CalendarEvent{
		AnyoneCanAddSelf:      false,
		Description:           "Welcome to Pizza Friday!",
		EndTime:               time.Now(),
		GuestsCanInviteOthers: false,
		GuestsCanModify:       false,
		Id:                    "",
		Locked:                true,
		StartTime:             time.Now(),
		Status:                "confirmed",
		Summary:               "Pizza Friday",
		Visibility:            "private",
	}

	pendingDates := make([]time.Time, len(dates))
	for i, d := range dates {
		num, err := strconv.ParseInt(d, 10, 64)
		if err != nil {
			Log.Error("failed parsing date int from rsvp form", zap.String("date", d))
			template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_fail.html"))).Execute(w, nil)
			return
		}
		pendingDates[i] = time.Unix(num, 0)
		newEvent.StartTime = pendingDates[i]
		newEvent.EndTime = pendingDates[i].Add(time.Hour + 5)
		newEvent.Id = d

		err = s.calendar.InviteToEvent(d, email, friendName)
		if err != nil && err == ErrEventNotFound {
			if err = s.calendar.CreateEvent(newEvent); err != nil {
				Log.Error("could not create event", zap.String("eventID", d), zap.String("email", email))
				template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_error.html"))).Execute(w, nil)
				return
			}
			err = s.calendar.InviteToEvent(d, email, friendName)
		}
		if err != nil {
			Log.Error("invite failed", zap.String("eventID", d), zap.String("email", email))
			template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_error.html"))).Execute(w, nil)
			return
		}
		Log.Debug("event updated", zap.String("eventID", d), zap.String("email", email), zap.String("name", friendName))
	}

	template.Must(template.ParseFiles(path.Join(s.config.StaticDir, "html/snippets/rsvp_success.html"))).Execute(w, nil)
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
