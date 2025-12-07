package pizza

import (
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"slices"
	"text/template"
	"time"
)

type WrappedPageData struct {
	LoggedIn     bool
	IsAvailable  bool
	Name         string
	Year         int
	Attendance   []string
	TotalFridays int
	TotalPunches int
}

func (s *Server) HandledWrapped(w http.ResponseWriter, r *http.Request) {
	s.wrappedGetMetric.Increment()
	plate, err := template.ParseFiles(path.Join(s.config.StaticDir, "html/wrapped.html"))
	if err != nil {
		slog.Error("template wrapped failure", "error", err)
		s.Handle500(w, r)
		return
	}
	if _, err = plate.ParseGlob(path.Join(s.config.StaticDir, "html/snippets/*.html")); err != nil {
		slog.Error("template snippets parse failure", "error", err)
		s.Handle500(w, r)
		return
	}

	now := time.Now()
	data := WrappedPageData{
		LoggedIn:   false,
		Attendance: make([]string, 0),
		Year:       now.Year(),
	}

	claims, ok := s.authenticateRequest(r)
	if !ok {
		s.executeTemplate(w, "Wrapped", data)
		return
	}
	data.LoggedIn = true
	data.Name = claims.GivenName

	if !IsTimeForWrapped() {
		data.IsAvailable = false
		s.executeTemplate(w, "Wrapped", data)
		return
	}

	fridays, err := s.store.GetFridaysInYear(data.Year)
	if err != nil {
		slog.Error("failed to get fridays in year", "err", err)
		data.IsAvailable = false
		s.executeTemplate(w, "Wrapped", data)
		return
	}

	// TODO sort fridays
	for _, friday := range fridays {
		if slices.Contains(friday.Guests, claims.Email) {
			data.Attendance = append(data.Attendance, fmt.Sprintf("%s %d", getShortMonth(friday.Date.Month()), friday.Date.Day()))
		}
	}

	data.IsAvailable = true
	data.TotalFridays = len(fridays)
	data.TotalPunches = len(data.Attendance) + 1

	s.executeTemplate(w, "Wrapped", data)

}

func IsTimeForWrapped() bool {
	// available only after the second week of December
	now := time.Now()
	return now.Month() == time.December && now.Day() >= 14
}
