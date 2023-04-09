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

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var StaticDir = "static"
var EventDuration = time.Hour * 4

type Server struct {
	s      http.Server
	config Config
}

func NewServer(config Config) (Server, error) {
	r := mux.NewRouter()
	r.HandleFunc("/", HandleIndex)
	r.HandleFunc("/submit", HandleSubmit)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(StaticDir))))

	return Server{
		s: http.Server{
			Addr:         fmt.Sprintf("0.0.0.0:%d", config.Port),
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			Handler:      r,
		},
		config: config,
	}, nil
}

func (s *Server) Start() error {
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

type IndexFridayData struct {
	Date string
	ID   int64
}

type PageData struct {
	FridayTimes []IndexFridayData
}

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	plate, err := template.ParseFiles(path.Join(StaticDir, "html/index.html"))
	if err != nil {
		Log.Error("template index failure", zap.Error(err))
		Handle500(w, r)
		return
	}
	data := PageData{}

	fridays, err := GetCachedFridays(30)
	if err != nil {
		Log.Error("failed to get fridays", zap.Error(err))
		Handle500(w, r)
		return
	}

	estZone, _ := time.LoadLocation("America/New_York")
	data.FridayTimes = make([]IndexFridayData, len(fridays))
	for i, t := range fridays {
		t = t.In(estZone)
		data.FridayTimes[i].Date = t.Format(time.RFC822)
		data.FridayTimes[i].ID = t.Unix()
	}

	if err = plate.Execute(w, data); err != nil {
		Log.Error("template execution failure", zap.Error(err))
		Handle500(w, r)
		return
	}
}

func HandleSubmit(w http.ResponseWriter, r *http.Request) {
	plate, err := template.ParseFiles(path.Join(StaticDir, "html/submit.html"))
	if err != nil {
		Log.Error("template submit failure", zap.Error(err))
		Handle500(w, r)
		return
	}
	data := PageData{}

	Log.Debug("incoming submit request", zap.Stringer("url", r.URL))

	form := r.URL.Query()
	dates, ok := form["date"]
	if !ok {
		Handle4xx(w, r)
		return
	}
	email := form.Get("email")
	if len(email) == 0 {
		Handle4xx(w, r)
		return
	}
	email = strings.ToLower(email)
	Log.Debug("rsvp request", zap.String("email", email), zap.Strings("dates", dates))

	if ok, err := IsFriendAllowed(email); !ok {
		if err != nil {
			Log.Error("error checking email for rsvp request", zap.Error(err))
			Handle500(w, r)
		} else {
			Handle4xx(w, r)
		}
		return
	}

	pendingDates := make([]time.Time, len(dates))
	for i, d := range dates {
		num, err := strconv.ParseInt(d, 10, 64)
		if err != nil {
			Log.Error("error parsing date int from rsvp form", zap.String("date", d))
			Handle500(w, r)
			return
		}
		pendingDates[i] = time.Unix(num, 0)

		friendName, err := GetFriendName(email)
		if err != nil {
			Log.Error("could not get friend name", zap.Error(err), zap.String("email", email))
			Handle500(w, r)
			return
		}

		event, err := InviteToCalendarEvent(d, pendingDates[i], pendingDates[i].Add(time.Hour+5), friendName, email)
		if err != nil {
			Log.Error("invite failed", zap.String("eventID", d), zap.String("email", email))
			Handle500(w, r)
			return
		}
		Log.Debug("event updated", zap.Any("event", event))
	}

	if err = plate.Execute(w, data); err != nil {
		Log.Error("template execution failure", zap.Error(err))
		Handle500(w, r)
		return
	}
}

func Handle4xx(w http.ResponseWriter, r *http.Request) {
	plate, err := template.ParseFiles(path.Join(StaticDir, "html/4xx.html"))
	if err != nil {
		Log.Error("template 4xx failure", zap.Error(err))
		Handle500(w, r)
		return
	}
	data := PageData{}
	if err = plate.Execute(w, data); err != nil {
		Log.Error("template execution failure", zap.Error(err))
		Handle500(w, r)
		return
	}
}

func Handle500(w http.ResponseWriter, r *http.Request) {
	plate, err := template.ParseFiles(path.Join(StaticDir, "html/500.html"))
	if err != nil {
		Log.Error("template 400 failure", zap.Error(err))
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
