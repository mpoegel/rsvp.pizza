package pizza

import (
	"time"
)

type Accessor interface {
	CreateTables() error
	IsFriendAllowed(email string) (bool, error)
	GetFriendName(email string) (string, error)
	GetUpcomingFridays(daysAhead int) ([]Friday, error)
	GetUpcomingFridaysAfter(after time.Time, daysAhead int) ([]Friday, error)
	DoesFridayExist(date time.Time) (bool, error)
	AddFriday(date time.Time) error
	AddFriend(email, name string) error
	ListFriends() ([]Friend, error)
	ListFridays() ([]Friday, error)
	RemoveFriend(email string) error
	RemoveFriday(date time.Time) error
	// UpdateFriday(friday Friday) error
}

type Friend struct {
	Email string
	Name  string
}

type Friday struct {
	Date    time.Time
	Group   *string
	Details *string
}
