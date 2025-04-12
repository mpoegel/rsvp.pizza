package pizza_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pizza "github.com/mpoegel/rsvp.pizza/pkg/pizza"
	assert "github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	require "github.com/stretchr/testify/require"
)

func TestHandleIndex(t *testing.T) {
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

	claims := &pizza.TokenClaims{
		Email: "foo@bar.com",
		Name:  "test",
		Exp:   time.Now().Add(1 * time.Hour).Unix(),
	}
	authenticator.On("IsValidSession", mock.Anything).Return(claims, true)
	authenticator.On("GetAuthURL").Return("/auth")
	accessor.On("AddFriend", claims.Email, claims.Name).Return(nil)
	accessor.On("GetPreferences", claims.Email).Return(pizza.Preferences{}, nil)
	accessor.On("GetUpcomingFridays", 30).Return([]pizza.Friday{}, nil)

	server, err := pizza.NewServer(config, accessor, calendar, authenticator, metrics)
	require.Nil(t, err)
	mux := http.NewServeMux()
	server.LoadRoutes(mux)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// WHEN
	req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	require.Nil(t, err)
	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: "foobar",
	})
	res, err := http.DefaultClient.Do(req)

	// THEN
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	accessor.AssertExpectations(t)
}

func TestHandleSubmit(t *testing.T) {
	// GIVEN
	config := pizza.LoadConfigEnv()
	config.StaticDir = "../../static"
	accessor := &pizza.MockAccessor{}
	calendar := &pizza.MockCalendar{}
	authenticator := &pizza.MockAuthenticator{}
	metrics := &pizza.MockMetricsRegistry{}
	counter := &pizza.MockCounterMetric{}
	groupName := "all"
	estZone, _ := time.LoadLocation("America/New_York")

	metrics.On("NewCounterMetric", mock.Anything, mock.Anything).Return(counter)
	counter.On("Increment").Return()

	claims := &pizza.TokenClaims{
		GivenName: "Foo",
		Email:     "foo@bar.com",
		Name:      "test",
		Exp:       time.Now().Add(1 * time.Hour).Unix(),
		Groups:    []string{groupName},
	}
	authenticator.On("IsValidSession", mock.Anything).Return(claims, true)
	friday1 := pizza.Friday{
		Date:      time.Unix(1672060005, 0).In(estZone),
		Group:     &groupName,
		Enabled:   true,
		MaxGuests: 5,
	}
	friday2 := pizza.Friday{
		Date:      time.Unix(1672040005, 0).In(estZone),
		Group:     &groupName,
		Enabled:   true,
		MaxGuests: 5,
	}
	accessor.On("GetFriday", friday1.Date).Return(friday1, nil)
	accessor.On("GetFriday", friday2.Date).Return(friday2, nil)
	accessor.On("AddFriendToFriday", claims.Email, friday1).Return(nil)
	accessor.On("AddFriendToFriday", claims.Email, friday2).Return(nil)
	calendar.On("InviteToEvent", "1672060005", claims.Email, claims.GivenName).Return(nil)
	calendar.On("InviteToEvent", "1672040005", claims.Email, claims.GivenName).Return(nil)

	server, err := pizza.NewServer(config, accessor, calendar, authenticator, metrics)
	require.Nil(t, err)
	mux := http.NewServeMux()
	server.LoadRoutes(mux)
	ts := httptest.NewServer(mux)
	defer ts.Close()
	url := fmt.Sprintf("%s/x/rsvp?date=1672060005&date=1672040005", ts.URL)

	// WHEN
	req, err := http.NewRequest(http.MethodPost, url, nil)
	require.Nil(t, err)
	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: "foobar",
	})
	res, err := http.DefaultClient.Do(req)

	// THEN
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	accessor.AssertExpectations(t)
	calendar.AssertExpectations(t)
}
