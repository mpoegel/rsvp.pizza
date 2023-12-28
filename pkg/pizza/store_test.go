package pizza_test

import (
	"errors"
	"testing"
	"time"

	"github.com/mpoegel/rsvp.pizza/pkg/pizza"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAccessor struct {
	mock.Mock
}

func (m *MockAccessor) IsFriendAllowed(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

func (m *MockAccessor) GetFriendName(email string) (string, error) {
	args := m.Called(email)
	return args.String(0), args.Error(1)
}

func (m *MockAccessor) GetUpcomingFridays(daysAhead int) ([]time.Time, error) {
	args := m.Called(daysAhead)
	return args.Get(0).([]time.Time), args.Error(1)
}

func (m *MockAccessor) GetUpcomingFridaysAfter(after time.Time, daysAhead int) ([]time.Time, error) {
	args := m.Called(after, daysAhead)
	return args.Get(0).([]time.Time), args.Error(1)
}

func (m *MockAccessor) CreateTables() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAccessor) AddFriday(date time.Time) error {
	args := m.Called(date)
	return args.Error(0)
}

func TestStore_IsFriendAllowed(t *testing.T) {
	// GIVEN
	friend := "ted@tedlasso.com"
	notFriend := "rupert@tedlasso.com"
	accessor := MockAccessor{}
	accessor.On("IsFriendAllowed", friend).Once().Return(true, nil)
	accessor.On("IsFriendAllowed", notFriend).Once().Return(false, nil)
	store := pizza.NewStore(&accessor)

	// WHEN & THEN
	isAllowed, err := store.IsFriendAllowed(friend)
	assert.Nil(t, err)
	assert.True(t, isAllowed)

	isAllowed, err = store.IsFriendAllowed(notFriend)
	assert.Nil(t, err)
	assert.False(t, isAllowed)

	// FINALLY
	assert.True(t, accessor.AssertExpectations(t))
}

func TestStore_IsFriendAllowedWithCache(t *testing.T) {
	// GIVEN
	friend := "ted@tedlasso.com"
	notFriend := "rupert@tedlasso.com"
	errFriend := "jade@tedlasso.com"
	accessor := MockAccessor{}
	accessor.On("IsFriendAllowed", friend).Once().Return(true, nil)
	accessor.On("IsFriendAllowed", notFriend).Times(2).Return(false, nil)
	accessor.On("IsFriendAllowed", errFriend).Once().Return(false, errors.New("err"))
	accessor.On("IsFriendAllowed", errFriend).Once().Return(false, nil)
	store := pizza.NewStore(&accessor)
	store.SetCacheTTL(time.Minute, 50*time.Millisecond, 50*time.Millisecond)

	allowedTest := func(email string, result bool) {
		// WHEN
		isAllowed, err := store.IsFriendAllowed(email)

		// THEN
		assert.Nil(t, err)
		if result {
			assert.True(t, isAllowed)
		} else {
			assert.False(t, isAllowed)
		}
	}

	t.Run("emptyCacheAllowed", func(t *testing.T) { allowedTest(friend, true) })
	t.Run("emptyCacheNotAllowed", func(t *testing.T) { allowedTest(notFriend, false) })
	t.Run("cachedNotAllowed", func(t *testing.T) { allowedTest(notFriend, false) })

	time.Sleep(51 * time.Millisecond)
	t.Run("refillCacheNotAllowed", func(t *testing.T) { allowedTest(notFriend, false) })

	isAllowed, err := store.IsFriendAllowed(errFriend)
	assert.NotNil(t, err)
	assert.False(t, isAllowed)

	t.Run("emptyCacheAfterError", func(t *testing.T) { allowedTest(errFriend, false) })

	// FINALLY
	assert.True(t, accessor.AssertExpectations(t))
}

func TestStore_GetFriendName(t *testing.T) {
	// GIVEN
	friend := "ted@tedlasso.com"
	friendName := "ted"
	accessor := MockAccessor{}
	accessor.On("GetFriendName", friend).Times(2).Return(friendName, nil)
	store := pizza.NewStore(&accessor)

	// WHEN & THEN
	name, err := store.GetFriendName(friend)
	assert.Nil(t, err)
	assert.Equal(t, friendName, name)

	name, err = store.GetFriendName(friend)
	assert.Nil(t, err)
	assert.Equal(t, friendName, name)

	// FINALLY
	assert.True(t, accessor.AssertExpectations(t))
}

func TestStore_GetFriendNameWithCache(t *testing.T) {
	// GIVEN
	friend := "ted@tedlasso.com"
	friendName := "ted"
	accessor := MockAccessor{}
	accessor.On("GetFriendName", friend).Times(2).Return(friendName, nil)
	store := pizza.NewStore(&accessor)
	store.SetCacheTTL(time.Minute, 50*time.Millisecond, 50*time.Millisecond)

	// WHEN & THEN
	name, err := store.GetFriendName(friend)
	assert.Nil(t, err)
	assert.Equal(t, friendName, name)

	isAllowed, err := store.IsFriendAllowed(friend)
	assert.Nil(t, err)
	assert.True(t, isAllowed)

	name, err = store.GetFriendName(friend)
	assert.Nil(t, err)
	assert.Equal(t, friendName, name)

	time.Sleep(51 * time.Millisecond)
	name, err = store.GetFriendName(friend)
	assert.Nil(t, err)
	assert.Equal(t, friendName, name)

	// FINALLY
	assert.True(t, accessor.AssertExpectations(t))
}

func TestStore_GetUpcomingFridays(t *testing.T) {
	// GIVEN
	accessor := MockAccessor{}
	fridays := []time.Time{time.Now()}
	daysAhead := 30
	accessor.On("GetUpcomingFridays", daysAhead).Times(2).Return(fridays, nil)
	store := pizza.NewStore(&accessor)

	// WHEN & THEN
	f, err := store.GetUpcomingFridays(daysAhead)
	assert.Nil(t, err)
	assert.Equal(t, fridays, f)

	f, err = store.GetUpcomingFridays(daysAhead)
	assert.Nil(t, err)
	assert.Equal(t, fridays, f)

	// FINALLY
	assert.True(t, accessor.AssertExpectations(t))
}

func TestStore_GetUpcomingFridaysWithCache(t *testing.T) {
	// GIVEN
	accessor := MockAccessor{}
	fridays := []time.Time{time.Now()}
	daysAhead := 30
	accessor.On("GetUpcomingFridays", daysAhead).Times(2).Return(fridays, nil)
	store := pizza.NewStore(&accessor)
	store.SetCacheTTL(50*time.Millisecond, 50*time.Millisecond, 50*time.Millisecond)

	getFridays := func(t *testing.T) {
		// WHEN & THEN
		f, err := store.GetUpcomingFridays(daysAhead)
		assert.Nil(t, err)
		assert.Equal(t, fridays, f)
	}

	t.Run("emptyCache", getFridays)
	t.Run("fullCache", getFridays)

	time.Sleep(51 * time.Millisecond)
	t.Run("refillCache", getFridays)

	// FINALLY
	assert.True(t, accessor.AssertExpectations(t))
}
