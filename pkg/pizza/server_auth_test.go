package pizza_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	pizza "github.com/mpoegel/rsvp.pizza/pkg/pizza"
	assert "github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	require "github.com/stretchr/testify/require"
)

func TestHandleLogin(t *testing.T) {
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

	authenticator.On("IsValidSession", "foo").Return(nil, true)
	authenticator.On("VerifyToken", mock.Anything, "faketoken").Return(&pizza.IDToken{}, nil)
	authenticator.On("VerifyToken", mock.Anything, "badtoken").Return(nil, errors.New("bad token"))
	authenticator.On("GetAuthCodeURL", mock.Anything, mock.Anything).Return("/auth?state=foo")

	client := http.Client{
		// Disable redirects
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	server, err := pizza.NewServer(config, accessor, calendar, authenticator, metrics)
	require.Nil(t, err)
	mux := http.NewServeMux()
	server.LoadRoutes(mux)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// ---
	// Request has valid token
	// ---

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/login", nil)
	require.Nil(t, err)
	req.Header.Add("Authorization", "Bearer faketoken")

	// WHEN
	res, err := client.Do(req)

	// THEN
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// ---
	// Request is missing token
	// ---

	req, err = http.NewRequest(http.MethodGet, ts.URL+"/login", nil)
	require.Nil(t, err)

	// WHEN
	res, err = client.Do(req)

	// THEN
	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, res.StatusCode)
	loc, err := res.Location()
	assert.Nil(t, err)
	assert.Equal(t, ts.URL+"/auth?state=foo", loc.String())

	// ---
	// Request has invalid token
	// ---

	req, err = http.NewRequest(http.MethodGet, ts.URL+"/login", nil)
	require.Nil(t, err)
	req.Header.Add("Authorization", "Bearer badtoken")

	// WHEN
	res, err = client.Do(req)

	// THEN
	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, res.StatusCode)
	loc, err = res.Location()
	assert.Nil(t, err)
	assert.Equal(t, ts.URL+"/auth?state=foo", loc.String())
}

func TestHandleLoginCallback(t *testing.T) {
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

	state := ""
	authenticator.On("GetAuthCodeURL", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			state = args[1].(string)
		}).
		Return("/auth?state=foo")
	authenticator.On("ExchangeCodeForToken", mock.Anything, mock.Anything, "bar").Return(&pizza.IDToken{}, nil)
	authenticator.On("IsValidSession", "foobar").Times(1).Return(nil, false)
	authenticator.On("IsValidSession", mock.Anything).Return(nil, true)

	client := http.Client{
		// Disable redirects
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	server, err := pizza.NewServer(config, accessor, calendar, authenticator, metrics)
	require.Nil(t, err)

	mux := http.NewServeMux()
	server.LoadRoutes(mux)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// ---
	// bad state input
	// ---

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/login/callback?state=foobar", ts.URL), nil)
	require.Nil(t, err)

	// WHEN
	res, err := client.Do(req)

	// THEN
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	// ---
	// valid state
	// ---

	req, err = http.NewRequest(http.MethodGet, ts.URL+"/login", nil)
	require.Nil(t, err)

	// WHEN
	res, err = client.Do(req)

	// THEN
	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, res.StatusCode)
	loc, err := res.Location()
	assert.Nil(t, err)
	fmt.Println(loc)

	// simulate login callback after login redirect
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/login/callback?state=%s&code=bar", ts.URL, state), nil)
	require.Nil(t, err)

	// WHEN
	res, err = client.Do(req)

	// THEN
	assert.Nil(t, err)
	assert.Equal(t, http.StatusSeeOther, res.StatusCode)
}
