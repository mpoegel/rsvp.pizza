package pizza_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mpoegel/rsvp.pizza/pkg/pizza"
)

func TestHandleIndex(t *testing.T) {
	// GIVEN
	config, err := pizza.LoadConfig("../../configs/pizza.yaml")
	require.Nil(t, err)
	server, err := pizza.NewServer(config, nil)
	require.Nil(t, err)
	ts := httptest.NewServer(http.HandlerFunc(server.HandleIndex))
	defer ts.Close()

	// WHEN
	res, err := http.Get(ts.URL)

	// THEN
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.NotNil(t, res)
}

func TestHandleSubmit(t *testing.T) {
	// GIVEN
	config, err := pizza.LoadConfig("../../configs/pizza.yaml")
	require.Nil(t, err)
	server, err := pizza.NewServer(config, nil)
	require.Nil(t, err)
	ts := httptest.NewServer(http.HandlerFunc(server.HandleSubmit))
	defer ts.Close()
	url := fmt.Sprintf("%s?date=1672060005&date=1672040005&email=popfizz@foo.com", ts.URL)

	// WHEN
	res, err := http.Post(url, "", nil)

	// THEN
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.NotNil(t, res)
}
