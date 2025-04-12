package pizza

import (
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

func (s *Server) HandleFriday(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticateRequest(r)
	if !ok {
		s.executeTemplate(w, "RSVPFail", nil)
		return
	}
	fridayTime, err := parseFridayTime(r.PathValue("ID"))
	if err != nil {
		s.executeTemplate(w, "RSVPFail", nil)
		return
	}

	friday, err := s.loadFriday(fridayTime, claims)
	if err != nil {
		if claims.HasRole("pizza_host") {
			friday = &Friday{
				Date:      fridayTime,
				Guests:    []string{},
				MaxGuests: 10,
			}
		} else {
			s.executeTemplate(w, "RSVPFail", nil)
			return
		}
	}

	fData := s.newIndexFridayData(friday, claims)
	s.executeTemplate(w, "SelectedFriday", fData)
}

func (s *Server) HandleRSVP(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticateRequest(r)
	if !ok {
		s.executeTemplate(w, "RSVPFail", nil)
		return
	}

	slog.Debug("incoming submit request", "url", r.URL)

	form := r.URL.Query()
	dates, ok := form["date"]
	if !ok {
		s.executeTemplate(w, "RSVPFail", nil)
		return
	}
	email := strings.ToLower(claims.Email)

	if err := r.ParseForm(); err != nil {
		slog.Error("form parse failure on rsvp", "error", err)
		w.Write(getToast("bad request"))
		return
	}
	if len(r.Form["plus-one"]) > 0 {
		if !claims.HasRole("plusOne") && !claims.HasRole("pizza_host") {
			s.executeTemplate(w, "RSVPError", nil)
			return
		}
		email = r.Form["plus-one"][0]
		friend, err := s.store.GetFriendByEmail(email)
		if err != nil {
			w.Write(getToast("friend not found"))
			return
		}
		slog.Info("rsvp request", "name", friend.Name, "dates", dates, "by", claims.Email)
	} else {
		slog.Debug("rsvp request", "email", email, "dates", dates)
	}

	for _, d := range dates {
		fridayTime, err := parseFridayTime(d)
		if err != nil {
			s.executeTemplate(w, "RSVPFail", nil)
			return
		}
		friday, err := s.loadFriday(fridayTime, claims)
		if err != nil {
			s.executeTemplate(w, "RSVPFail", nil)
			return
		}

		if err = s.CreateAndInvite(d, *friday, email, claims.GivenName); err != nil {
			s.executeTemplate(w, "RSVPError", nil)
			return
		}
	}

	s.executeTemplate(w, "RSVPSuccess", nil)
}

func (s *Server) HandleDeleteRSVP(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticateRequest(r)
	if !ok {
		s.executeTemplate(w, "RSVPFail", nil)
		return
	}

	form := r.URL.Query()
	dates, ok := form["date"]
	if !ok {
		s.executeTemplate(w, "RSVPFail", nil)
		return
	}

	guestEmail := claims.Email
	if guestEmails, ok := form["guest"]; ok {
		if !claims.HasRole("pizza_host") {
			s.executeTemplate(w, "RSVPFail", nil)
			return
		}
		guestEmail = guestEmails[0]
	}

	slog.Debug("incoming decline request", "url", r.URL, "email", guestEmail, "dates", dates)

	for _, d := range dates {
		fridayTime, err := parseFridayTime(d)
		if err != nil {
			s.executeTemplate(w, "RSVPFail", nil)
			return
		}
		friday, err := s.loadFriday(fridayTime, claims)
		if err != nil {
			s.executeTemplate(w, "RSVPFail", nil)
			return
		}

		if slices.Contains(friday.Guests, guestEmail) {
			if err = s.store.RemoveFriendFromFriday(guestEmail, friday.Date); err != nil {
				slog.Error("failed to remove friend from friday", "err", err, "email", guestEmail, "friday", d)
				s.executeTemplate(w, "RSVPFail", nil)
				return
			}
			if s.config.Calendar.Enabled {
				if err = s.calendar.DeclineEvent(d, guestEmail); err != nil {
					slog.Error("failed to decline calendar invite", "err", err, "email", guestEmail, "friday", d)
					s.executeTemplate(w, "RSVPFail", nil)
					return
				}
			}
		}
	}

	s.executeTemplate(w, "DeclineSuccess", nil)
}

func (s *Server) HandleFridayGetEdit(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticateRequest(r)
	if !ok || !claims.HasRole("pizza_host") {
		s.executeTemplate(w, "RSVPFail", nil)
		return
	}

	fridayTime, err := parseFridayTime(r.PathValue("ID"))
	if err != nil {
		s.executeTemplate(w, "RSVPFail", nil)
		return
	}
	friday, err := s.loadFriday(fridayTime, claims)
	if err != nil {
		s.executeTemplate(w, "RSVPFail", nil)
		return
	}

	fData := s.newIndexFridayData(friday, claims)
	s.executeTemplate(w, "SelectedFridayEdit", fData)
}

func (s *Server) HandleFridaySaveEdit(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticateRequest(r)
	if !ok || !claims.HasRole("pizza_host") {
		s.executeTemplate(w, "RSVPFail", nil)
		return
	}

	fridayTime, err := parseFridayTime(r.PathValue("ID"))
	if err != nil {
		s.executeTemplate(w, "RSVPFail", nil)
		return
	}
	friday, err := s.loadFriday(fridayTime, claims)
	if err != nil {
		s.executeTemplate(w, "RSVPFail", nil)
		return
	}

	if err := r.ParseForm(); err != nil {
		slog.Error("form parse failure on admin edit", "error", err)
		w.Write(getToast("bad request"))
		return
	}
	group := r.Form["group"]
	details := r.Form["details"]
	maxGuestsStr := r.Form["maxGuests"]

	slog.Info("admin edit", "group", group, "details", details, "maxGuests", maxGuestsStr)

	if len(group) > 0 {
		friday.Group = &group[0]
	}
	if len(details) > 0 {
		friday.Details = &details[0]
	}
	maxGuests, err := strconv.ParseInt(maxGuestsStr[0], 10, 64)
	if err != nil {
		w.Write(getToast("max guests must be an integer"))
		return
	}
	friday.MaxGuests = int(maxGuests)
	if err = s.store.UpdateFriday(*friday); err != nil {
		s.executeTemplate(w, "RSVPFail", nil)
		return
	}

	fData := s.newIndexFridayData(friday, claims)
	s.executeTemplate(w, "SelectedFriday", fData)
}

func (s *Server) HandleFridayEnable(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticateRequest(r)
	if !ok || !claims.HasRole("pizza_host") {
		s.executeTemplate(w, "RSVPFail", nil)
		return
	}

	fridayTime, err := parseFridayTime(r.PathValue("ID"))
	if err != nil {
		s.executeTemplate(w, "RSVPFail", nil)
		return
	}
	slog.Info("enable friday", "time", fridayTime)
	friday, err := s.loadFriday(fridayTime, claims)
	if err != nil {
		err = s.store.AddFriday(fridayTime)
		if err != nil {
			s.executeTemplate(w, "RSVPFail", nil)
			return
		}
		friday, err = s.loadFriday(fridayTime, claims)
		if err != nil {
			s.executeTemplate(w, "RSVPFail", nil)
			return
		}
	}
	friday.Enabled = true

	if err = s.store.UpdateFriday(*friday); err != nil {
		slog.Info("update friday failed", "err", err)
		s.executeTemplate(w, "RSVPFail", nil)
		return
	}

	fData := s.newIndexFridayData(friday, claims)
	s.executeTemplate(w, "SelectedFridayEdit", fData)
}

func (s *Server) HandleFridayDisable(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticateRequest(r)
	if !ok || !claims.HasRole("pizza_host") {
		s.executeTemplate(w, "RSVPFail", nil)
		return
	}

	fridayTime, err := parseFridayTime(r.PathValue("ID"))
	if err != nil {
		s.executeTemplate(w, "RSVPFail", nil)
		return
	}
	friday, err := s.loadFriday(fridayTime, claims)
	if err != nil {
		s.executeTemplate(w, "RSVPFail", nil)
		return
	}

	friday.Enabled = false
	if err = s.store.UpdateFriday(*friday); err != nil {
		s.executeTemplate(w, "RSVPFail", nil)
		return
	}

	fData := s.newIndexFridayData(friday, claims)
	s.executeTemplate(w, "SelectedFriday", fData)
}
