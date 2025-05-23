package pizza

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	jsonapi "github.com/hashicorp/jsonapi"
	api "github.com/mpoegel/rsvp.pizza/pkg/api"
)

func WriteAPIError(err error, status int, w http.ResponseWriter) {
	errObj := &jsonapi.ErrorObject{
		Title:  http.StatusText(status),
		Detail: err.Error(),
		Status: strconv.FormatInt(int64(status), 10),
	}
	allErrs := []*jsonapi.ErrorObject{errObj}
	// ignore marshal errors
	jsonapi.MarshalErrors(w, allErrs)
}

func (s *Server) HandleAPIAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		slog.Error("form parse failure on admin edit", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	opt := AuthTokenOptions{}
	if len(r.Form["username"]) > 0 {
		opt.Username = r.Form["username"][0]
	}
	if len(r.Form["password"]) > 0 {
		opt.Password = r.Form["password"][0]
	}
	if len(r.Form["grant_type"]) > 0 {
		opt.GrantType = r.Form["grant_type"][0]
	}
	if len(r.Form["refresh_token"]) > 0 {
		opt.RefreshToken = r.Form["refresh_token"][0]
	}

	jwt, err := s.authenticator.GetToken(context.Background(), opt)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	encoder := json.NewEncoder(w)
	if err = encoder.Encode(jwt); err != nil {
		slog.Error("json encoding failure", "error", err)
	}
}

func (s *Server) HandleAPIFriday(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.CheckAuthorization(r)
	if !ok {
		WriteAPIError(errors.New("not authorized"), http.StatusUnauthorized, w)
		return
	}

	if r.Header.Get("Accept") != jsonapi.MediaType {
		WriteAPIError(fmt.Errorf("must accept %s", jsonapi.MediaType), http.StatusNotAcceptable, w)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.HandleAPIGetFriday(claims, w, r)
	case http.MethodPatch:
		s.HandleAPIPatchFriday(claims, w, r)
	default:
		WriteAPIError(errors.New("only GET and PATCH allowed"), http.StatusMethodNotAllowed, w)
	}
}

func (s *Server) HandleAPIGetFriday(accessToken *AccessToken, w http.ResponseWriter, r *http.Request) {
	fridays := make([]Friday, 0)
	var err error
	estZone, _ := time.LoadLocation("America/New_York")

	fridayID := r.PathValue("ID")
	isDirectReq := len(fridayID) > 0
	if isDirectReq {
		rawTime, err := strconv.ParseInt(fridayID, 10, 64)
		if err != nil {
			WriteAPIError(err, http.StatusBadRequest, w)
			return
		}

		f, err := s.store.GetFriday(time.Unix(rawTime, 0).In(estZone))
		if err != nil {
			WriteAPIError(fmt.Errorf("no matching friday found with ID '%s'", fridayID), http.StatusNotFound, w)
			return
		}

		fridays = append(fridays, f)
	} else {
		fridays, err = s.store.GetUpcomingFridays(30)
		if err != nil {
			slog.Error("failed to get fridays", "error", err)
			WriteAPIError(errors.New("database error"), http.StatusInternalServerError, w)
			return
		}
	}

	res := make([]*api.Friday, 0)
	for _, f := range fridays {
		id := strconv.FormatInt(f.Date.Unix(), 10)

		friday := &api.Friday{
			ID:        id,
			StartTime: f.Date,
			Guests:    nil,
		}

		if f.Details != nil {
			friday.Details = *f.Details
		}

		// not part of invited group OR friday is disabled
		if (f.Group != nil && !accessToken.Claims.InGroup(*f.Group)) || !f.Enabled {
			// if this friday was specifically requested, the response needs to be 404
			if isDirectReq {
				WriteAPIError(fmt.Errorf("no matching friday found with ID '%s'", fridayID), http.StatusNotFound, w)
				return
			}
			continue
		}

		friday.Guests = make([]*api.Guest, len(f.Guests))
		for i, attendee := range f.Guests {
			if friend, err := s.store.GetFriendByEmail(attendee); err == nil {
				g := &api.Guest{
					ID:   friend.ID,
					Name: friend.Name,
				}
				friday.Guests[i] = g
			}
		}

		res = append(res, friday)
	}

	w.Header().Set("Content-Type", jsonapi.MediaType)
	w.WriteHeader(http.StatusOK)

	// if a specific friday was requested, return an object not an array as per json api spec
	if isDirectReq {
		err = jsonapi.MarshalPayload(w, res[0])
	} else {
		err = jsonapi.MarshalPayload(w, res)
	}

	if err != nil {
		slog.Warn("api marshal payload", "error", err)
		WriteAPIError(errors.New("failed to compose response data"), http.StatusInternalServerError, w)
	}
}

func (s *Server) HandleAPIPatchFriday(accessToken *AccessToken, w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != jsonapi.MediaType {
		WriteAPIError(fmt.Errorf("unsupported media type '%s'", r.Header.Get("Content-Type")), http.StatusUnsupportedMediaType, w)
		return
	}

	friday := &api.Friday{}
	if err := jsonapi.UnmarshalPayload(r.Body, friday); err != nil {
		WriteAPIError(err, http.StatusBadRequest, w)
		return
	}

	estZone, _ := time.LoadLocation("America/New_York")
	rawTime, err := strconv.ParseInt(friday.ID, 10, 64)
	if err != nil {
		WriteAPIError(err, http.StatusBadRequest, w)
		return
	}

	f, err := s.store.GetFriday(time.Unix(rawTime, 0).In(estZone))
	if err != nil {
		WriteAPIError(fmt.Errorf("no matching friday found with ID '%s'", friday.ID), http.StatusNotFound, w)
		return
	}
	if f.Details != nil {
		friday.Details = *f.Details
	}
	friday.StartTime = f.Date

	// not part of invited group OR friday not enabled
	if (f.Group != nil && !accessToken.Claims.InGroup(*f.Group)) || !f.Enabled {
		WriteAPIError(fmt.Errorf("no matching friday found with ID '%s'", friday.ID), http.StatusNotFound, w)
		return
	}

	// check the requested guests
	for _, g := range friday.Guests {
		// backwards compatibility for RSVPing using email instead of ID
		if g.ID != accessToken.Claims.Email {
			self, err := s.store.GetFriendByID(g.ID)
			if err != nil || self.Email != accessToken.Claims.Email {
				WriteAPIError(fmt.Errorf("not allowed to invite guest '%s'", g.ID), http.StatusUnauthorized, w)
				return
			}
		}
	}

	if friday.StartTime.Before(time.Now()) {
		// TODO inappropriate usage of StatusNotModified
		WriteAPIError(errors.New("friday is in the past"), http.StatusNotModified, w)
		return
	}

	if friday.StartTime.After(time.Now().AddDate(0, 1, 0)) {
		WriteAPIError(errors.New("friday is more than one month away"), http.StatusTooEarly, w)
		return
	}

	if friday.StartTime.Before(time.Now().Add(time.Hour * 24)) {
		// TODO inappropriate usage of StatusNotModified
		WriteAPIError(errors.New("modifications cannot be made less than 24 hours before the friday"), http.StatusNotModified, w)
		return
	}

	// all good to update invite
	slog.Info("rsvp request", "email", accessToken.Claims.Email)

	if err = s.CreateAndInvite(friday.ID, f, accessToken.Claims.Email, accessToken.Claims.Name); err != nil {
		WriteAPIError(errors.New("calendar failure"), http.StatusInternalServerError, w)
		return
	}

	// TODO replace GetEvent with local guest list
	if event, err := s.calendar.GetEvent(friday.ID); err != nil && err != ErrEventNotFound {
		slog.Warn("failed to get calendar event", "error", err, "eventID", friday.ID)
	} else {
		friday.Guests = make([]*api.Guest, 0)
		for _, attendee := range event.Attendees {
			if friend, err := s.store.GetFriendByEmail(attendee.Email); err == nil {
				g := &api.Guest{
					ID:   friend.ID,
					Name: friend.Name,
				}
				friday.Guests = append(friday.Guests, g)
			}
		}
	}

	w.Header().Set("Content-Type", jsonapi.MediaType)
	w.WriteHeader(http.StatusOK)

	if err = jsonapi.MarshalPayload(w, friday); err != nil {
		slog.Warn("api marshal payload", "error", err)
		WriteAPIError(errors.New("failed to compose response data"), http.StatusInternalServerError, w)
	}
}

func (s *Server) HandleAPIGuest(w http.ResponseWriter, r *http.Request) {
	_, ok := s.CheckAuthorization(r)
	if !ok {
		WriteAPIError(errors.New("not authorized"), http.StatusUnauthorized, w)
		return
	}

	if r.Header.Get("Accept") != jsonapi.MediaType {
		WriteAPIError(fmt.Errorf("must accept %s", jsonapi.MediaType), http.StatusNotAcceptable, w)
		return
	}

	guestID := r.PathValue("ID")
	if guestID == "" {
		WriteAPIError(errors.New("missing guest ID"), http.StatusBadRequest, w)
		return
	}

	friend, err := s.store.GetFriendByID(guestID)
	if err != nil {
		WriteAPIError(errors.New("guest does not exist"), http.StatusNotFound, w)
		return
	}

	guest := &api.Guest{
		ID:   guestID,
		Name: friend.Name,
	}

	w.Header().Set("Content-Type", jsonapi.MediaType)
	w.WriteHeader(http.StatusOK)

	if err = jsonapi.MarshalPayload(w, guest); err != nil {
		slog.Warn("api marshal payload", "error", err)
		WriteAPIError(errors.New("failed to compose response data"), http.StatusInternalServerError, w)
	}
}

func (s *Server) HandleAPIGuestProfile(w http.ResponseWriter, r *http.Request) {
	token, ok := s.CheckAuthorization(r)
	if !ok {
		WriteAPIError(errors.New("not authorized"), http.StatusUnauthorized, w)
		return
	}

	if r.Header.Get("Accept") != jsonapi.MediaType {
		WriteAPIError(fmt.Errorf("must accept %s", jsonapi.MediaType), http.StatusNotAcceptable, w)
		return
	}

	guestID := r.PathValue("ID")
	if guestID == "" {
		WriteAPIError(errors.New("missing guest ID"), http.StatusBadRequest, w)
		return
	}

	if self, err := s.store.GetFriendByEmail(token.Claims.Email); err != nil {
		// user does not exist in our db, but is authorized, so add them
		if err = s.store.AddFriend(token.Claims.Email, token.Claims.Name); err != nil {
			slog.Warn("failed to add new friend", "err", err, "email", token.Claims.Email)
			WriteAPIError(errors.New("failed to create new friend"), http.StatusInternalServerError, w)
			return
		}
	} else {
		// only the pizza hosts can view other guest profiles
		if self.ID != guestID && !token.Claims.HasRole("pizza_host") {
			WriteAPIError(errors.New("not authorized to view guest profile"), http.StatusForbidden, w)
			return
		}
	}

	friend, err := s.store.GetFriendByID(guestID)
	if err != nil {
		WriteAPIError(errors.New("guest does not exist"), http.StatusNotFound, w)
		return
	}

	prefs, err := s.store.GetPreferences(friend.Email)
	if err != nil {
		slog.Warn("failed to get preferences for friend that should exist", "err", err, "email", friend.Email)
		WriteAPIError(errors.New("could not get guest preferences"), http.StatusInternalServerError, w)
		return
	}

	guestProfile := api.GuestProfile{
		ID:       guestID,
		Email:    friend.Email,
		Toppings: make([]string, len(prefs.Toppings)),
		Cheese:   make([]string, len(prefs.Cheese)),
		Sauce:    make([]string, len(prefs.Sauce)),
		Doneness: prefs.Doneness.String(),
	}

	for i, topping := range prefs.Toppings {
		guestProfile.Toppings[i] = topping.String()
	}
	for i, cheese := range prefs.Cheese {
		guestProfile.Cheese[i] = cheese.String()
	}
	for i, sauce := range prefs.Sauce {
		guestProfile.Sauce[i] = sauce.String()
	}

	w.Header().Set("Content-Type", jsonapi.MediaType)
	w.WriteHeader(http.StatusOK)

	if err = jsonapi.MarshalPayload(w, &guestProfile); err != nil {
		slog.Warn("api marshal payload", "error", err)
		WriteAPIError(errors.New("failed to compose response data"), http.StatusInternalServerError, w)
	}

}
