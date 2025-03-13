package pizza_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	jsonapi "github.com/hashicorp/jsonapi"
	api "github.com/mpoegel/rsvp.pizza/pkg/api"
	pizza "github.com/mpoegel/rsvp.pizza/pkg/pizza"
	assert "github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	require "github.com/stretchr/testify/require"
)

func TestHandleApiToken(t *testing.T) {
	// GIVEN
	config := pizza.LoadConfigEnv()
	config.StaticDir = "../../static"
	accessor := &pizza.MockAccessor{}
	calendar := &pizza.MockCalendar{}
	authenticator := &pizza.MockAuthenticator{}
	metrics := &pizza.MockMetricsRegistry{}
	counter := &pizza.MockCounterMetric{}

	metrics.On("NewCounterMetric", mock.Anything, mock.Anything).Return(counter)
	counter.On("Increment").Return()

	jwt := &pizza.JWT{
		AccessToken:      "this_is_a_token",
		IDToken:          "id",
		ExpiresIn:        100,
		RefreshExpiresIn: 50,
		RefreshToken:     "refreshing",
		TokenType:        "arcade",
		NotBeforePolicy:  1,
		SessionState:     "good",
		Scope:            "everything",
	}
	tokenOpts := pizza.AuthTokenOptions{
		Username:  "foo",
		Password:  "secret",
		GrantType: "password",
	}
	authenticator.On("GetToken", mock.Anything, tokenOpts).Return(jwt, nil)

	server, err := pizza.NewServer(config, accessor, calendar, authenticator, metrics)
	require.Nil(t, err)
	mux := http.NewServeMux()
	server.LoadRoutes(mux)
	ts := httptest.NewServer(mux)
	defer ts.Close()
	url := fmt.Sprintf("%s/api/token?username=%s&password=%s&grant_type=%s", ts.URL, tokenOpts.Username, tokenOpts.Password, tokenOpts.GrantType)

	// WHEN
	req, err := http.NewRequest(http.MethodPost, url, nil)
	require.Nil(t, err)
	res, err := http.DefaultClient.Do(req)

	// THEN
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	b, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	assert.JSONEq(t, `{"access_token":"this_is_a_token","id_token":"id","expires_in":100,"refresh_expires_in":50,"refresh_token":"refreshing",
	"token_type":"arcade","not-before-policy":1,"session_state":"good","scope":"everything"}`, string(b))

	authenticator.AssertExpectations(t)
}

func TestHandleApiGetFriday(t *testing.T) {
	// GIVEN
	config := pizza.LoadConfigEnv()
	config.StaticDir = "../../static"
	accessor := &pizza.MockAccessor{}
	calendar := &pizza.MockCalendar{}
	authenticator := &pizza.MockAuthenticator{}
	metrics := &pizza.MockMetricsRegistry{}
	counter := &pizza.MockCounterMetric{}
	details := "details"

	metrics.On("NewCounterMetric", mock.Anything, mock.Anything).Return(counter)
	counter.On("Increment").Return()

	token := &pizza.AccessToken{
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	authenticator.On("DecodeAccessToken", mock.Anything, "token").Return(token, nil)
	friday1 := pizza.Friday{
		Date:      time.Now(),
		Details:   &details,
		Guests:    []string{},
		MaxGuests: 10,
		Enabled:   true,
	}
	friday1ID := strconv.FormatInt(friday1.Date.Unix(), 10)
	accessor.On("GetUpcomingFridays", 30).Return([]pizza.Friday{friday1}, nil)
	accessor.On("GetFriendByEmail", "kirk").Return(pizza.Friend{ID: "1", Name: "Captain Kirk"}, nil)
	event := pizza.CalendarEvent{
		Attendees: []string{"kirk"},
	}
	calendar.On("GetEvent", friday1ID).Return(event, nil)

	server, err := pizza.NewServer(config, accessor, calendar, authenticator, metrics)
	require.Nil(t, err)
	mux := http.NewServeMux()
	server.LoadRoutes(mux)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// WHEN
	req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/friday", nil)
	require.Nil(t, err)
	req.Header.Add("Authorization", "Bearer token")
	req.Header.Add("Accept", "application/vnd.api+json")
	res, err := http.DefaultClient.Do(req)

	// THEN
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	b, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	expected := fmt.Sprintf(`{
	"data":[{
		"type":"friday",
		"id":"%s",
		"attributes":{
			"details":"details",
			"start_time":%s
		},
		"relationships":{
			"guests":{
				"data":[
					{"id":"1","type":"guest"}
				]
			}
		},
		"links":{
			"self":"/api/friday/%s"
		}
	}],
	"included":[
		{
			"id":"1",
			"type":"guest",
			"attributes": {
				"name":"Captain Kirk"
			},
			"links": {
				"self": "/api/guest/1",
				"profile": "/api/guest/1/profile"
			}
		}
	]
}`,
		friday1ID, friday1ID, friday1ID)
	assert.JSONEq(t, expected, string(b))

	authenticator.AssertExpectations(t)
	accessor.AssertExpectations(t)
	calendar.AssertExpectations(t)
}

func TestHandleApiPatchFriday(t *testing.T) {
	// GIVEN
	config := pizza.LoadConfigEnv()
	config.StaticDir = "../../static"
	accessor := &pizza.MockAccessor{}
	calendar := &pizza.MockCalendar{}
	authenticator := &pizza.MockAuthenticator{}
	metrics := &pizza.MockMetricsRegistry{}
	counter := &pizza.MockCounterMetric{}
	groupName := "all"
	details := "details"
	estZone, _ := time.LoadLocation("America/New_York")
	fTime := time.Unix(time.Now().Add(time.Hour*36).Unix(), 0)
	reqFriday := &api.Friday{
		ID: strconv.FormatInt(fTime.Unix(), 10),
	}

	metrics.On("NewCounterMetric", mock.Anything, mock.Anything).Return(counter)
	counter.On("Increment").Return()

	token := &pizza.AccessToken{
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Claims: pizza.TokenClaims{
			Groups: []string{groupName},
			Email:  "foo@bar.com",
		},
	}
	authenticator.On("DecodeAccessToken", mock.Anything, "token").Return(token, nil)
	friday := pizza.Friday{
		Date:    fTime.In(estZone),
		Group:   &groupName,
		Enabled: true,
		Details: &details,
	}
	accessor.On("GetFriday", friday.Date).Return(friday, nil)
	accessor.On("AddFriendToFriday", token.Claims.Email, friday).Return(nil)
	accessor.On("GetFriendByEmail", mock.Anything).Return(pizza.Friend{ID: "2", Name: "Spock"}, nil)
	calendar.On("InviteToEvent", reqFriday.ID, token.Claims.Email, token.Claims.GivenName).Return(nil)
	event := pizza.CalendarEvent{
		Attendees: []string{"spock"},
	}
	calendar.On("GetEvent", reqFriday.ID).Return(event, nil)

	server, err := pizza.NewServer(config, accessor, calendar, authenticator, metrics)
	require.Nil(t, err)
	mux := http.NewServeMux()
	server.LoadRoutes(mux)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	reqBody := &bytes.Buffer{}
	require.Nil(t, jsonapi.MarshalPayload(reqBody, reqFriday))

	// WHEN
	req, err := http.NewRequest(http.MethodPatch, ts.URL+"/api/friday/"+reqFriday.ID, reqBody)
	require.Nil(t, err)
	req.Header.Add("Authorization", "Bearer token")
	req.Header.Add("Accept", "application/vnd.api+json")
	req.Header.Add("Content-Type", "application/vnd.api+json")
	res, err := http.DefaultClient.Do(req)

	// THEN
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	b, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	expected := fmt.Sprintf(`{
	"data":{
		"type":"friday",
		"id":"%s",
		"attributes":{
			"details":"details",
			"start_time":%s
		},
		"relationships":{
			"guests":{
				"data":[
					{"id":"2","type":"guest"}
				]
			}
		},
		"links":{
			"self":"/api/friday/%s"
		}
	},
	"included":[
		{
			"id":"2",
			"type":"guest",
			"attributes": {
				"name": "Spock"
			},
			"links": {
				"self": "/api/guest/2",
				"profile": "/api/guest/2/profile"
			}
		}
	]
}`,
		reqFriday.ID, reqFriday.ID, reqFriday.ID)
	assert.JSONEq(t, expected, string(b))

	authenticator.AssertExpectations(t)
	accessor.AssertExpectations(t)
	calendar.AssertExpectations(t)
}
