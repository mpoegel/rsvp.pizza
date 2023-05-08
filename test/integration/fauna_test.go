package pizza_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/mpoegel/rsvp.pizza/internal/pizza"

	"github.com/stretchr/testify/assert"
)

func TestIsFriendAllowed(t *testing.T) {
	// GIVEN
	client := pizza.NewFaunaClient(os.Getenv("FAUNADB_SECRET"))

	// WHEN & THEN
	client.IsFriendAllowed("fake.account@gmail.com")
}

func TestGetAllFridays(t *testing.T) {
	// GIVEN
	client := pizza.NewFaunaClient(os.Getenv("FAUNADB_SECRET"))

	// WHEN & THEN
	client.GetUpcomingFridays(14)
}

func TestGetFriendName(t *testing.T) {
	// GIVEN
	client := pizza.NewFaunaClient(os.Getenv("FAUNADB_SECRET"))

	// WHEN & THEN
	name, err := client.GetFriendName(os.Getenv("TEST_EMAIL"))
	assert.Nil(t, err)
	fmt.Println(name)
}
