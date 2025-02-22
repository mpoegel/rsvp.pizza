package pizza

import (
	"errors"
	"log/slog"
	"net/http"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"
)

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
		slog.Error("template wrapped failure", "error", err)
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
		allowed := false
		// TODO update with auth
		// allowed, err := s.store.IsFriendAllowed(email)
		// if err != nil {
		// 	slog.Error("is friend allowed check failed", "error", err)
		// 	s.Handle500(w, r)
		// 	return
		// }
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
			slog.Error("could not get friend name", "error", err)
			return
		}
		// only use the first name
		nameParts := strings.Split(data.Name, " ")
		data.Name = nameParts[0]
	}
	if err = plate.Execute(w, data); err != nil {
		slog.Error("template execution failure", "error", err)
		s.Handle500(w, r)
		return
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
	slog.Info("wrapped cache update", "year", year, "data", data)
	// update cache then return
	s.wrapped[year] = data
	return data, nil
}
