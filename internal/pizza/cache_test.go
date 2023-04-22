package pizza_test

import (
	"testing"
	"time"

	"github.com/mpoegel/rsvp.pizza/internal/pizza"

	"github.com/stretchr/testify/assert"
)

func TestCacheGet(t *testing.T) {
	// GIVEN
	data := []int{1, 2, 3}
	refresh := func(key string) ([]int, error) {
		return data, nil
	}
	cache := pizza.NewCache(100*time.Millisecond, refresh)

	// WHEN
	vals, err := cache.Get("foo")

	// THEN
	assert.Nil(t, err)
	assert.Equal(t, data, vals)

	// WHEN
	data = []int{4, 5, 6}
	time.Sleep(200 * time.Millisecond)
	vals, err = cache.Get("foo")

	// THEN
	assert.Nil(t, err)
	assert.Equal(t, data, vals)
}

func TestCacheStore(t *testing.T) {
	// GIVEN
	data := 42
	cache := pizza.NewCache[int](100*time.Millisecond, nil)

	// WHEN
	_, err := cache.Get("foo")

	// THEN
	assert.NotNil(t, err)

	// WHEN
	cache.Store("foo", data)
	val, err := cache.Get("foo")
	assert.Nil(t, err)
	assert.Equal(t, data, val)
}
