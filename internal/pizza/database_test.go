package pizza_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/mpoegel/rsvp.pizza/internal/pizza"

	"github.com/stretchr/testify/assert"
)

func TestIsFriendAllowed(t *testing.T) {
	pizza.IsFriendAllowed("fake.account@gmail.com")
}

func TestGetAllFridays(t *testing.T) {
	pizza.GetUpcomingFridays(14)
}

func TestGetFriendName(t *testing.T) {
	name, err := pizza.GetFriendName(os.Getenv("TEST_EMAIL"))
	assert.Nil(t, err)
	fmt.Println(name)
}
