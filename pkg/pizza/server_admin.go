package pizza

import (
	"net/http"
	"path"
	"strconv"
	"text/template"
	"time"

	zap "go.uber.org/zap"
)

const futureFridayLimit = 90

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
	plate, err := template.ParseFiles(path.Join(s.config.StaticDir, "html/admin.html"))
	if err != nil {
		Log.Error("template submit failure", zap.Error(err))
		s.Handle500(w, r)
		return
	}

	claims, ok := s.authenticateRequest(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if !claims.HasRole("pizza_host") {
		s.Handle4xx(w, r)
		return
	}

	data := PageData{
		Name: claims.GivenName,
	}

	allFridays := getFutureFridays()
	setFridays, err := s.store.GetUpcomingFridays(futureFridayLimit)
	if err != nil {
		Log.Error("failed to get fridays", zap.Error(err))
		s.Handle500(w, r)
		return
	}
	fridayIndex := 0
	data.FridayTimes = make([]IndexFridayData, 0)
	for _, friday := range allFridays {
		f := IndexFridayData{
			Date:   friday.Format(time.RFC822),
			ID:     friday.Unix(),
			Guests: nil,
			Active: false,
		}
		if fridayIndex < len(setFridays) && friday.Equal(setFridays[fridayIndex]) {
			f.Active = true
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
				if name, err := s.store.getFriendName(email); err != nil {
					f.Guests[k] = email
				} else {
					f.Guests[k] = name
				}
			}
		}

		data.FridayTimes = append(data.FridayTimes, f)
	}

	if err = plate.Execute(w, data); err != nil {
		Log.Error("template execution failure", zap.Error(err))
		s.Handle500(w, r)
		return
	}
}

func (s *Server) HandleAdminSubmit(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticateRequest(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if !claims.HasRole("pizza_host") {
		s.Handle4xx(w, r)
		return
	}

	form := r.URL.Query()
	dates := form["date"]
	loc, _ := time.LoadLocation("America/New_York")
	dateIndex := 0
	allFridays := getFutureFridays()
	for _, d := range allFridays {
		f := time.Now().AddDate(1, 0, 0).In(loc)
		if dateIndex < len(dates) {
			num, err := strconv.ParseInt(dates[dateIndex], 10, 64)
			if err != nil {
				Log.Error("failed parsing date int from rsvp form", zap.String("date", dates[dateIndex]))
				s.Handle500(w, r)
				return
			}
			f = time.Unix(num, 0).In(loc)
		}
		if d.Equal(f) {
			// friday selected, so add it
			if exists, err := s.store.accessor.DoesFridayExist(f); err != nil {
				Log.Error("failed check friday", zap.Error(err))
				continue
			} else if !exists {
				err := s.store.accessor.AddFriday(f)
				if err != nil {
					Log.Error("failed to add friday", zap.Error(err))
				} else {
					Log.Info("added friday", zap.Time("date", f))
				}
			}
			dateIndex++
		} else if f.After(d) {
			// friday is not selected, so remove it
			// TODO warn if users have already RSVP'ed
			if exists, err := s.store.accessor.DoesFridayExist(d); err != nil {
				Log.Error("failed to check friday", zap.Error(err))
				continue
			} else if exists {
				err := s.store.accessor.RemoveFriday(d)
				if err != nil {
					Log.Error("failed to remove friday", zap.Error(err))
				} else {
					Log.Info("removed friday", zap.Time("date", d))
				}
				// NOTE the calendar event must be deleted manually
			}
		}
		// else {
		// f.Before(d) == true
		// do nothing
		// }
	}

	http.Redirect(w, r, "/admin", http.StatusFound)
}
