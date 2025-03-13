package pizza

import (
	"time"

	"github.com/mpoegel/rsvp.pizza/pkg/types"
)

type Accessor interface {
	CreateTables() error

	GetFriendByEmail(email string) (Friend, error)
	AddFriend(email, name string) error

	GetUpcomingFridays(daysAhead int) ([]Friday, error)
	GetUpcomingFridaysAfter(after time.Time, daysAhead int) ([]Friday, error)

	DoesFridayExist(date time.Time) (bool, error)
	ListFridays() ([]Friday, error)
	AddFriday(date time.Time) error
	GetFriday(date time.Time) (Friday, error)
	RemoveFriday(date time.Time) error
	UpdateFriday(friday Friday) error
	AddFriendToFriday(email string, friday Friday) error

	GetPreferences(email string) (Preferences, error)
	SetPreferences(email string, preferences Preferences) error
}

type Friend struct {
	ID    string
	Email string
	Name  string
}

type Friday struct {
	Date      time.Time
	Group     *string
	Details   *string
	Guests    []string
	MaxGuests int
	Enabled   bool
}

type Preferences struct {
	Toppings []types.Topping
	Cheese   []types.Cheese
	Sauce    []types.Sauce
	Doneness types.Doneness
}
