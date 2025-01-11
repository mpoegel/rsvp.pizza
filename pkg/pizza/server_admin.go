package pizza

import (
	"fmt"
	"net/http"
	"path"
	"strconv"
	"text/template"
	"time"

	zap "go.uber.org/zap"
)

const futureFridayLimit = 30

func getFutureFridays() []time.Time {
	dates := make([]time.Time, 0)
	loc, _ := time.LoadLocation("America/New_York")
	start := time.Now()
	friday := time.Date(start.Year(), start.Month(), start.Day(), 17, 30, 0, 0, loc)
	for friday.Weekday() != time.Friday {
		friday = friday.AddDate(0, 0, 1)
	}
	endDate := time.Now().AddDate(0, 0, futureFridayLimit)
	for friday.Before(endDate) {
		dates = append(dates, friday)
		friday = friday.AddDate(0, 0, 7)
	}
	return dates
}

func (s *Server) HandleAdmin(w http.ResponseWriter, r *http.Request) {

	claims, ok := s.authenticateRequest(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if !claims.HasRole("pizza_host") {
		s.Handle4xx(w, r)
		return
	}

	plate, err := template.ParseFiles(path.Join(s.config.StaticDir, "html/admin.html"))
	if err != nil {
		Log.Error("template submit failure", zap.Error(err))
		s.Handle500(w, r)
		return
	}
	if _, err = plate.ParseGlob(path.Join(s.config.StaticDir, "html/snippets/*.html")); err != nil {
		Log.Error("template snippets submit failure", zap.Error(err))
		s.Handle500(w, r)
		return
	}

	data := PageData{
		Name: claims.GivenName,
	}

	allFridays := getFutureFridays()
	setFridays, err := s.store.GetUpcomingFridays(futureFridayLimit)
	Log.Info("loaded data", zap.Any("fridays", setFridays))
	if err != nil {
		Log.Error("failed to get fridays", zap.Error(err))
		s.Handle500(w, r)
		return
	}
	fridayIndex := 0
	data.FridayTimes = make([]IndexFridayData, 0)
	for _, friday := range allFridays {
		f := IndexFridayData{
			Date:      friday.Format(time.RFC822),
			ID:        friday.Unix(),
			Guests:    nil,
			Active:    false,
			Group:     "",
			Details:   "",
			MaxGuests: 10,
		}
		if fridayIndex < len(setFridays) && friday.Equal(setFridays[fridayIndex].Date) {
			f.Active = true
			if setFridays[fridayIndex].Group != nil {
				f.Group = *setFridays[fridayIndex].Group
			}
			if setFridays[fridayIndex].Details != nil {
				f.Details = *setFridays[fridayIndex].Details
			}
			f.MaxGuests = setFridays[fridayIndex].MaxGuests
			f.Active = setFridays[fridayIndex].Enabled
			fridayIndex++
		}

		eventID := strconv.FormatInt(f.ID, 10)
		if event, err := s.calendar.GetEvent(eventID); err != nil && err != ErrEventNotFound {
			Log.Warn("failed to get calendar event", zap.Error(err), zap.String("eventID", eventID))
			f.Guests = make([]string, 0)
		} else if err != nil {
			f.Guests = make([]string, 0)
		} else {
			f.Guests = make([]string, len(event.Attendees))
			for k, email := range event.Attendees {
				if name, err := s.store.GetFriendName(email); err != nil {
					f.Guests[k] = email
				} else {
					f.Guests[k] = name
				}
			}
		}

		data.FridayTimes = append(data.FridayTimes, f)
	}

	if err = plate.ExecuteTemplate(w, "Admin", data); err != nil {
		Log.Error("template execution failure", zap.Error(err))
		s.Handle500(w, r)
		return
	}
}

func getToast(msg string) []byte {
	return []byte(fmt.Sprintf(`<span class="toast">%s</span>`, msg))
}

func (s *Server) HandleAdminEdit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		Log.Error("form parse failure on admin edit", zap.Error(err))
		w.Write(getToast("bad request"))
		return
	}
	form := r.URL.Query()
	isActive := form["active"]
	group := r.Form["group"]
	details := r.Form["details"]
	maxGuestsStr := r.Form["maxGuests"]
	dates := r.Form["date"]
	needsActivation := len(r.Form["activate"]) > 0

	Log.Info("admin edit", zap.Strings("dates", dates),
		zap.Strings("group", group),
		zap.Strings("details", details),
		zap.Strings("maxGuests", maxGuestsStr),
		zap.Strings("isActive", isActive),
		zap.Bool("needsActivation", needsActivation))

	loc, _ := time.LoadLocation("America/New_York")
	num, err := strconv.ParseInt(dates[0], 10, 64)
	if err != nil {
		Log.Error("failed parsing date int from rsvp form", zap.String("date", dates[0]))
		w.Write(getToast("parse error"))
		return
	}

	maxGuests, err := strconv.ParseInt(maxGuestsStr[0], 10, 64)
	if err != nil {
		w.Write(getToast("max guests must be an integer"))
		return
	}

	friday := Friday{
		Date:      time.Unix(num, 0).In(loc),
		Group:     nil,
		Details:   nil,
		MaxGuests: int(maxGuests),
		Enabled:   needsActivation,
	}
	if len(group) > 0 && group[0] != "" {
		friday.Group = &group[0]
	}
	if len(details) > 0 && details[0] != "" {
		friday.Details = &details[0]
	}

	if exists, err := s.store.DoesFridayExist(friday.Date); err != nil {
		Log.Error("failed check friday", zap.Error(err))
	} else if needsActivation && !exists {
		err := s.store.AddFriday(friday.Date)
		if err != nil {
			Log.Error("failed to add friday", zap.Error(err))
			w.Write(getToast("failed to add friday"))
		} else if err = s.store.UpdateFriday(friday); err != nil {
			Log.Error("failed to update friday", zap.Error(err))
			w.Write(getToast("failed to update friday"))
		} else {
			Log.Info("added friday", zap.Any("friday", friday))
			w.Write(getToast("added friday"))
		}
	} else if !needsActivation && exists {
		if err := s.store.UpdateFriday(friday); err != nil {
			Log.Error("failed to disable friday", zap.Error(err))
		} else {
			Log.Info("disabled friday", zap.Time("date", friday.Date))
			w.Write(getToast("disabled friday"))
		}
	} else if err = s.store.UpdateFriday(friday); err != nil {
		Log.Error("failed to update friday", zap.Error(err))
		w.Write(getToast("failed to update friday"))
	} else {
		Log.Info("updated", zap.Any("friday", friday))
		w.Write(getToast("updated friday"))
	}
}
