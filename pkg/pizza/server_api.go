package pizza

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	gocloak "github.com/Nerzal/gocloak/v13"
	jwt "github.com/golang-jwt/jwt/v5"
	jsonapi "github.com/google/jsonapi"
	mux "github.com/gorilla/mux"
	api "github.com/mpoegel/rsvp.pizza/pkg/api"
	zap "go.uber.org/zap"
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
		Log.Error("form parse failure on admin edit", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	opt := gocloak.TokenOptions{}
	if len(r.Form["username"]) > 0 {
		opt.Username = &r.Form["username"][0]
	}
	if len(r.Form["password"]) > 0 {
		opt.Password = &r.Form["password"][0]
	}
	if len(r.Form["grant_type"]) > 0 {
		opt.GrantType = &r.Form["grant_type"][0]
	}
	if len(r.Form["refresh_token"]) > 0 {
		opt.RefreshToken = &r.Form["refresh_token"][0]
	}

	jwt, err := s.keycloak.GetToken(context.Background(), opt)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	encoder := json.NewEncoder(w)
	if err = encoder.Encode(jwt); err != nil {
		Log.Error("json encoding failure", zap.Error(err))
	}
}

func (s *Server) HandleAPIFriday(w http.ResponseWriter, r *http.Request) {
	token, claims, ok := s.CheckAuthorization(r)
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
		s.HandleAPIGetFriday(token, claims, w, r)
	case http.MethodPatch:
		s.HandleAPIPatchFriday(token, claims, w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *Server) HandleAPIGetFriday(token *jwt.Token, claims *TokenClaims, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fridays := make([]Friday, 0)
	var err error
	estZone, _ := time.LoadLocation("America/New_York")

	fridayID, ok := vars["ID"]
	if ok {
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
			Log.Error("failed to get fridays", zap.Error(err))
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
			Details:   *f.Details,
			Guests:    nil,
		}

		if event, err := s.calendar.GetEvent(id); err != nil && err != ErrEventNotFound {
			Log.Warn("failed to get calendar event", zap.Error(err), zap.String("eventID", id))
		} else {
			friday.Guests = make([]*api.Guest, len(event.Attendees))
			for k, email := range event.Attendees {
				g := &api.Guest{
					ID:    email,
					Email: email,
				}
				if name, err := s.store.GetFriendName(email); err == nil {
					g.Name = name
				}
				friday.Guests[k] = g
			}
		}

		res = append(res, friday)
	}

	w.Header().Set("Content-Type", jsonapi.MediaType)
	w.WriteHeader(http.StatusOK)

	if err = jsonapi.MarshalPayload(w, res); err != nil {
		Log.Warn("api marshal payload", zap.Error(err))
		WriteAPIError(errors.New("failed to compose response data"), http.StatusInternalServerError, w)
	}
}

func (s *Server) HandleAPIPatchFriday(token *jwt.Token, claims *TokenClaims, w http.ResponseWriter, r *http.Request) {
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
	friday.Details = *f.Details
	friday.StartTime = f.Date

	// not part of invited group
	if f.Group != nil && !claims.InGroup(*f.Group) {
		WriteAPIError(fmt.Errorf("no matching friday found with ID '%s'", friday.ID), http.StatusNotFound, w)
		return
	}

	// check the requested guests
	for _, g := range friday.Guests {
		if g.ID != claims.Email {
			WriteAPIError(fmt.Errorf("not allowed to invite guest '%s'", g.ID), http.StatusUnauthorized, w)
			return
		}
	}

	// all good to update invite
	Log.Info("rsvp request", zap.String("email", claims.Email))

	if err = s.CreateAndInvite(friday.ID, friday.StartTime, claims.Email, claims.Name); err != nil {
		WriteAPIError(errors.New("calendar failure"), http.StatusInternalServerError, w)
		return
	}

	if event, err := s.calendar.GetEvent(friday.ID); err != nil && err != ErrEventNotFound {
		Log.Warn("failed to get calendar event", zap.Error(err), zap.String("eventID", friday.ID))
	} else {
		friday.Guests = make([]*api.Guest, len(event.Attendees))
		for k, email := range event.Attendees {
			g := &api.Guest{
				ID:    email,
				Email: email,
			}
			if name, err := s.store.GetFriendName(email); err == nil {
				g.Name = name
			}
			friday.Guests[k] = g
		}
	}

	w.Header().Set("Content-Type", jsonapi.MediaType)
	w.WriteHeader(http.StatusOK)

	if err = jsonapi.MarshalPayload(w, friday); err != nil {
		Log.Warn("api marshal payload", zap.Error(err))
		WriteAPIError(errors.New("failed to compose response data"), http.StatusInternalServerError, w)
	}
}
