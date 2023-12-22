package pizza

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var StaticDir = "static"
var EventDuration = time.Hour * 4

type Server struct {
	s        http.Server
	store    *Store
	calendar *Calendar
	config   Config

	indexGetMetric      CounterMetric
	submitPostMetric    CounterMetric
	requestErrorMetric  CounterMetric
	internalErrorMetric CounterMetric
}

func NewServer(config Config, metricsReg MetricsRegistry) (Server, error) {
	r := mux.NewRouter()

	var accessor Accessor
	if faunaSecret := os.Getenv("FAUNADB_SECRET"); len(faunaSecret) > 0 {
		accessor = NewFaunaClient(faunaSecret)
	} else {
		panic("no FAUNADB_SECRET found")
	}

	googleCal, err := NewGoogleCalendar(config.Calendar.CredentialFile, config.Calendar.TokenFile, config.Calendar.ID, context.Background())
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
		indexGetMetric: metricsReg.NewCounterMetric("pizza_requests",
			map[string]string{"method": "get", "path": "/"}),
		submitPostMetric: metricsReg.NewCounterMetric("pizza_requests",
			map[string]string{"method": "post", "path": "/submit"}),
		requestErrorMetric: metricsReg.NewCounterMetric("pizza_errors",
			map[string]string{"statusCode": "4xx"}),
		internalErrorMetric: metricsReg.NewCounterMetric("pizza_errors",
			map[string]string{"statusCode": "500"}),
	}

	r.HandleFunc("/", s.HandleIndex)
	r.HandleFunc("/submit", s.HandleSubmit)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(StaticDir))))

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

type IndexFridayData struct {
	Date   string
	ID     int64
	Guests []int
}

type PageData struct {
	FridayTimes []IndexFridayData
}

func (s *Server) HandleIndex(w http.ResponseWriter, r *http.Request) {
	s.indexGetMetric.Increment()
	plate, err := template.ParseFiles(path.Join(StaticDir, "html/index.html"))
	if err != nil {
		Log.Error("template index failure", zap.Error(err))
		s.Handle500(w, r)
		return
	}
	data := PageData{}

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
	plate, err := template.ParseFiles(path.Join(StaticDir, "html/submit.html"))
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

func (s *Server) Handle4xx(w http.ResponseWriter, r *http.Request) {
	s.requestErrorMetric.Increment()
	plate, err := template.ParseFiles(path.Join(StaticDir, "html/4xx.html"))
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
	plate, err := template.ParseFiles(path.Join(StaticDir, "html/500.html"))
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
