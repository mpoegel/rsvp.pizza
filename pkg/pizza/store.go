package pizza

import (
	"strconv"
	"time"
)

type Accessor interface {
	CreateTables() error
	IsFriendAllowed(email string) (bool, error)
	GetFriendName(email string) (string, error)
	GetUpcomingFridays(daysAhead int) ([]time.Time, error)
	GetUpcomingFridaysAfter(after time.Time, daysAhead int) ([]time.Time, error)
	DoesFridayExist(date time.Time) (bool, error)
	AddFriday(date time.Time) error
	AddFriend(email, name string) error
	ListFriends() ([]Friend, error)
	ListFridays() ([]Friday, error)
	RemoveFriend(email string) error
	RemoveFriday(date time.Time) error
}

type Friend struct {
	Email string
	Name  string
}

type Friday struct {
	Date time.Time
}

const (
	DefaultFridayCacheTTL         = 1 * time.Hour
	DefaultFriendNameCacheTTL     = 24 * time.Hour
	DefaultNegativeFriendCacheTTL = 5 * time.Minute
)

type Store struct {
	accessor            Accessor
	fridayCache         *Cache[[]time.Time]
	friendNameCache     *Cache[string]
	negativeFriendCache *Cache[bool]
}

func NewStore(accessor Accessor) *Store {
	return &Store{
		accessor:            accessor,
		fridayCache:         nil,
		friendNameCache:     nil,
		negativeFriendCache: nil,
	}
}

func (s *Store) SetCacheTTL(fridayCacheTTL, friendNameCacheTTL, negativeFriendCacheTTL time.Duration) {
	s.fridayCache = NewCache2(fridayCacheTTL, s.getUpcomingFridaysStr)
	s.friendNameCache = NewCache2(friendNameCacheTTL, s.getFriendName)
	s.negativeFriendCache = NewCache2[bool](negativeFriendCacheTTL, nil)
}

func (s *Store) IsFriendAllowed(email string) (bool, error) {
	if s.negativeFriendCache != nil && s.negativeFriendCache.Has(email) {
		return false, nil
	}
	if s.friendNameCache != nil && s.friendNameCache.Has(email) {
		return true, nil
	}
	exists, err := s.accessor.IsFriendAllowed(email)
	if err != nil {
		return false, err
	}
	if !exists && s.negativeFriendCache != nil {
		s.negativeFriendCache.Store(email, false)
	}
	return exists, nil
}

func (s *Store) getFriendName(email string) (string, error) {
	return s.accessor.GetFriendName(email)
}

func (s *Store) GetFriendName(email string) (string, error) {
	if s.friendNameCache != nil {
		return s.friendNameCache.Get(email)
	} else {
		return s.getFriendName(email)
	}
}

func (s *Store) getUpcomingFridaysStr(daysAhead string) ([]time.Time, error) {
	days, err := strconv.ParseInt(daysAhead, 10, 32)
	if err != nil {
		return nil, err
	}
	return s.accessor.GetUpcomingFridays(int(days))
}

func (s *Store) GetUpcomingFridays(daysAhead int) ([]time.Time, error) {
	if s.fridayCache != nil {
		return s.fridayCache.Get(strconv.Itoa(daysAhead))
	} else {
		return s.getUpcomingFridaysStr(strconv.Itoa(daysAhead))
	}
}
