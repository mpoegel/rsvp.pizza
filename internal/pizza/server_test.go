package pizza_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mpoegel/rsvp.pizza/internal/pizza"
)

func TestHandleIndex(t *testing.T) {
	// GIVEN
	pizza.StaticDir = "../../static"
	ts := httptest.NewServer(http.HandlerFunc(pizza.HandleIndex))
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
	pizza.StaticDir = "../../static"
	ts := httptest.NewServer(http.HandlerFunc(pizza.HandleSubmit))
	defer ts.Close()
	url := fmt.Sprintf("%s?date=1672060005&date=1672040005&email=popfizz@foo.com", ts.URL)

	// WHEN
	res, err := http.Post(url, "", nil)

	// THEN
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.NotNil(t, res)
}

func TestHandleConfirmation(t *testing.T) {
	// GIVEN

	// WHEN

	// THEN
}
