package pizza

import (
	"log/slog"
	"net/http"
	"slices"
	"strings"
)

func (s *Server) HandleFriday(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticateRequest(r)
	if !ok {
		s.executeTemplate(w, "RSVPFail", nil)
		return
	}

	friday, err := s.loadFriday(r.PathValue("ID"), claims)
	if err != nil {
		s.executeTemplate(w, "RSVPFail", nil)
		return
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
	slog.Debug("rsvp request", "email", email, "dates", dates)

	for _, d := range dates {
		friday, err := s.loadFriday(d, claims)
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

	slog.Debug("incoming decline request", "url", r.URL, "email", claims.Email, "dates", dates)

	for _, d := range dates {
		friday, err := s.loadFriday(d, claims)
		if err != nil {
			s.executeTemplate(w, "RSVPFail", nil)
			return
		}

		if slices.Contains(friday.Guests, claims.Email) {
			if err = s.store.RemoveFriendFromFriday(claims.Email, friday.Date); err != nil {
				slog.Error("failed to remove friend from friday", "err", err, "email", claims.Email, "friday", d)
				s.executeTemplate(w, "RSVPFail", nil)
				return
			}
			if s.config.Calendar.Enabled {
				if err = s.calendar.DeclineEvent(d, claims.Email); err != nil {
					slog.Error("failed to decline calendar invite", "err", err, "email", claims.Email, "friday", d)
					s.executeTemplate(w, "RSVPFail", nil)
					return
				}
			}
		}
	}

	s.executeTemplate(w, "DeclineSuccess", nil)
}
