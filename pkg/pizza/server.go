package pizza

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"

	oidc "github.com/coreos/go-oidc"
	uuid "github.com/google/uuid"
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
	store    *Store
	calendar *Calendar
	config   Config

	oauth2Provider *oidc.Provider
	oauth2Conf     oauth2.Config
	verifier       *oidc.IDTokenVerifier

	indexGetMetric      CounterMetric
	submitPostMetric    CounterMetric
	wrappedGetMetric    CounterMetric
	requestErrorMetric  CounterMetric
	internalErrorMetric CounterMetric

	wrapped  map[int]WrappedData
	sessions map[string]*TokenClaims

	// activeState string
}

func NewServer(config Config, metricsReg MetricsRegistry) (Server, error) {
	r := mux.NewRouter()

	var accessor Accessor
	var err error
	if config.UseSQLite {
		Log.Info("using the sqlite accessor")
		accessor, err = NewSQLAccessor(config.DBFile)
		if err != nil {
			return Server{}, err
		}
	} else if len(config.FaunaSecret) > 0 {
		Log.Info("using the faunadb accessor")
		accessor = NewFaunaClient(config.FaunaSecret)
	} else {
		panic("no FAUNADB_SECRET found")
	}

	googleCal, err := NewGoogleCalendar(config.Calendar.CredentialFile, config.Calendar.TokenFile, config.Calendar.ID, context.Background())
	if err != nil {
		return Server{}, err
	}

	// ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// defer cancel()
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, config.OAuth2.RealmsURL)
	if err != nil {
		return Server{}, err
	}

	s := Server{
		s: http.Server{
			Addr:         fmt.Sprintf("0.0.0.0:%d", config.Port),
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			Handler:      r,
		},
		store:    NewStore(accessor),
		calendar: NewCalendar(googleCal),
		config:   config,

		oauth2Provider: provider,
		oauth2Conf: oauth2.Config{
			ClientID:     config.OAuth2.ClientID,
			ClientSecret: config.OAuth2.ClientSecret,
			RedirectURL:  config.OAuth2.RedirectURL,
			Endpoint:     provider.Endpoint(),
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		},
		verifier: provider.Verifier(&oidc.Config{
			ClientID: config.OAuth2.ClientID,
		}),

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
	r.HandleFunc("/submit", s.HandleSubmit)
	r.HandleFunc("/wrapped", s.HandledWrapped)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(config.StaticDir))))
	r.HandleFunc("/login", s.HandleLogin)
	r.HandleFunc("/login/callback", s.HandleLoginCallback)

	return s, nil
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

func (s *Server) GetWrapped(year int) (WrappedData, error) {
	// restrict range
	if year != 2023 {
		return WrappedData{}, errors.New("no wrapped for year")
	}

	// check cache
	if d, ok := s.wrapped[year]; ok {
		return d, nil
	}

	// fetch from source
	start := time.Time{}.AddDate(year, 1, 1)
	end := time.Time{}.AddDate(year, 12, 31)
	events, err := s.calendar.ListEventsBetween(start, end, 100)
	if err != nil {
		return WrappedData{}, err
	}

	data := WrappedData{
		Friends:      map[string][]time.Time{},
		TotalFridays: 0,
	}
	for _, event := range events {
		for _, attendee := range event.Attendees {
			if _, ok := data.Friends[attendee]; !ok {
				data.Friends[attendee] = []time.Time{event.StartTime}
			} else {
				data.Friends[attendee] = append(data.Friends[attendee], event.StartTime)
			}
		}
		data.TotalFridays++
	}
	Log.Info("wrapped cache update", zap.Int("year", year), zap.Any("data", data))
	// update cache then return
	s.wrapped[year] = data
	return data, nil
}

type IndexFridayData struct {
	Date   string
	ID     int64
	Guests []int
}

type PageData struct {
	FridayTimes []IndexFridayData
	Name        string
}

func (s *Server) HandleIndex(w http.ResponseWriter, r *http.Request) {
	s.indexGetMetric.Increment()

	var claims *TokenClaims
	for _, cookie := range r.Cookies() {
		if cookie.Name == "session" {
			var ok bool
			claims, ok = s.sessions[cookie.Value]
			// either bad session or auth has expired
			if !ok || time.Now().After(time.Unix(claims.Exp, 0)) {
				delete(s.sessions, cookie.Value)
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
		}
	}
	if claims == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	Log.Info("welcome", zap.String("name", claims.Name))

	plate, err := template.ParseFiles(path.Join(s.config.StaticDir, "html/index.html"))
	if err != nil {
		Log.Error("template index failure", zap.Error(err))
		s.Handle500(w, r)
		return
	}
	data := PageData{
		Name: claims.GivenName,
	}

	fridays, err := s.store.GetUpcomingFridays(30)
	if err != nil {
		Log.Error("failed to get fridays", zap.Error(err))
		s.Handle500(w, r)
		return
	}

	estZone, _ := time.LoadLocation("America/New_York")
	data.FridayTimes = make([]IndexFridayData, len(fridays))
	for i, t := range fridays {
		t = t.In(estZone)
		data.FridayTimes[i].Date = t.Format(time.RFC822)
		data.FridayTimes[i].ID = t.Unix()

		eventID := strconv.FormatInt(data.FridayTimes[i].ID, 10)
		if event, err := s.calendar.GetEvent(eventID); err != nil && err != ErrEventNotFound {
			Log.Warn("failed to get calendar event", zap.Error(err), zap.String("eventID", eventID))
			data.FridayTimes[i].Guests = make([]int, 0)
		} else if err != nil {
			data.FridayTimes[i].Guests = make([]int, 0)
		} else {
			data.FridayTimes[i].Guests = make([]int, len(event.Attendees))
		}
	}

	if err = plate.Execute(w, data); err != nil {
		Log.Error("template execution failure", zap.Error(err))
		s.Handle500(w, r)
		return
	}
}

func (s *Server) HandleSubmit(w http.ResponseWriter, r *http.Request) {
	s.submitPostMetric.Increment()
	plate, err := template.ParseFiles(path.Join(s.config.StaticDir, "html/submit.html"))
	if err != nil {
		Log.Error("template submit failure", zap.Error(err))
		s.Handle500(w, r)
		return
	}
	data := PageData{}

	Log.Debug("incoming submit request", zap.Stringer("url", r.URL))

	form := r.URL.Query()
	dates, ok := form["date"]
	if !ok {
		s.Handle4xx(w, r)
		return
	}
	email := form.Get("email")
	if len(email) == 0 {
		s.Handle4xx(w, r)
		return
	}
	email = strings.ToLower(email)
	Log.Debug("rsvp request", zap.String("email", email), zap.Strings("dates", dates))

	if ok, err := s.store.IsFriendAllowed(email); !ok {
		if err != nil {
			Log.Error("error checking email for rsvp request", zap.Error(err))
			s.Handle500(w, r)
		} else {
			s.Handle4xx(w, r)
		}
		return
	}

	friendName, err := s.store.GetFriendName(email)
	if err != nil {
		Log.Error("could not get friend name", zap.Error(err), zap.String("email", email))
		s.Handle500(w, r)
		return
	}

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
			s.Handle500(w, r)
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
				s.Handle500(w, r)
				return
			}
			err = s.calendar.InviteToEvent(d, email, friendName)
		}
		if err != nil {
			Log.Error("invite failed", zap.String("eventID", d), zap.String("email", email))
			s.Handle500(w, r)
			return
		}
		Log.Debug("event updated", zap.String("eventID", d), zap.String("email", email), zap.String("name", friendName))
	}

	if err = plate.Execute(w, data); err != nil {
		Log.Error("template execution failure", zap.Error(err))
		s.Handle500(w, r)
		return
	}
}

func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
	state := uuid.New()
	rawAccessToken := r.Header.Get("Authorization")
	if rawAccessToken == "" {
		s.sessions[state.String()] = nil
		http.Redirect(w, r, s.oauth2Conf.AuthCodeURL(state.String()), http.StatusFound)
		return
	}

	authParts := strings.Split(rawAccessToken, " ")
	if len(authParts) != 2 {
		w.WriteHeader(400)
		return
	}

	ctx := context.Background()
	_, err := s.verifier.Verify(ctx, authParts[1])
	if err != nil {
		s.sessions[state.String()] = nil
		http.Redirect(w, r, s.oauth2Conf.AuthCodeURL(state.String()), http.StatusFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

type TokenClaims struct {
	Exp               int64  `json:"exp"`
	Iat               int64  `json:"iat"`
	AuthTime          int64  `json:"auth_time"`
	Jti               string `json:"jti"`
	Iss               string `json:"iss"`
	Aud               string `json:"aud"`
	Sub               string `json:"sub"`
	Typ               string `json:"typ"`
	Azp               string `json:"azp"`
	SessionState      string `json:"session_state"`
	At_hash           string `json:"at_hash"`
	Acr               string `json:"acr"`
	Sid               string `json:"sid"`
	EmailVerified     bool   `json:"email_verified"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	Email             string `json:"email"`
}

func (s *Server) HandleLoginCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if _, ok := s.sessions[state]; !ok {
		Log.Warn("state did not match")
		http.Error(w, "state did not match", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	oauth2Token, err := s.oauth2Conf.Exchange(ctx, r.URL.Query().Get("code"))
	if err != nil {
		Log.Warn("failed to exchange code for token", zap.Error(err))
		http.Error(w, "auth error", http.StatusInternalServerError)
		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		Log.Warn("no id_token field in oauth2 token")
		http.Error(w, "auth error", http.StatusInternalServerError)
		return
	}

	idToken, err := s.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		Log.Warn("failed to verify ID token", zap.Error(err))
		http.Error(w, "auth error", http.StatusInternalServerError)
		return
	}

	var claims TokenClaims
	if err := idToken.Claims(&claims); err != nil {
		Log.Warn("failed to get claims", zap.Error(err))
		http.Error(w, "auth error", http.StatusInternalServerError)
		return
	}

	Log.Info("login success", zap.Any("claims", claims))
	cookie := &http.Cookie{
		Name:     "session",
		Value:    state,
		Path:     "/",
		Expires:  time.Now().AddDate(0, 0, 10),
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	}
	if err := cookie.Valid(); err != nil {
		Log.Warn("bad cookie", zap.Error(err))
	}
	http.SetCookie(w, cookie)
	r.AddCookie(cookie)

	s.sessions[state] = &claims
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type WrappedPageData struct {
	Email        string
	Name         string
	Attendance   []string
	TotalFridays int
}

func (s *Server) HandledWrapped(w http.ResponseWriter, r *http.Request) {
	s.wrappedGetMetric.Increment()
	plate, err := template.ParseFiles(path.Join(s.config.StaticDir, "html/wrapped.html"))
	if err != nil {
		Log.Error("template wrapped failure", zap.Error(err))
		s.Handle500(w, r)
		return
	}
	data := WrappedPageData{}
	form := r.URL.Query()
	email := form.Get("email")
	yearStr := form.Get("year")
	year := 2023
	if len(yearStr) != 0 {
		year, err = strconv.Atoi(yearStr)
		if err != nil {
			s.Handle4xx(w, r)
			return
		}
	}
	if len(email) > 0 {
		allowed, err := s.store.IsFriendAllowed(email)
		if err != nil {
			Log.Error("is friend allowed check failed", zap.Error(err))
			s.Handle500(w, r)
			return
		}
		if !allowed {
			s.Handle4xx(w, r)
			return
		}
		wrapped, err := s.GetWrapped(year)
		if err != nil {
			// TODO possible 500 here too
			s.Handle4xx(w, r)
			return
		}
		data = WrappedPageData{
			Email:        email,
			Name:         "",
			Attendance:   make([]string, len(wrapped.Friends[email])),
			TotalFridays: wrapped.TotalFridays,
		}
		for i, t := range wrapped.Friends[email] {
			data.Attendance[i] = t.Format(time.DateOnly)
		}
		data.Name, err = s.store.GetFriendName(email)
		if err != nil {
			Log.Error("could not get friend name", zap.Error(err))
			return
		}
		// only use the first name
		nameParts := strings.Split(data.Name, " ")
		data.Name = nameParts[0]
	}
	if err = plate.Execute(w, data); err != nil {
		Log.Error("template execution failure", zap.Error(err))
		s.Handle500(w, r)
		return
	}
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
