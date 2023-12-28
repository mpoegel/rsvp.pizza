package pizza_test

import (
	"os"
	"testing"
	"time"

	"github.com/mpoegel/rsvp.pizza/pkg/pizza"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSqlAccessor_IsFriendAllowed(t *testing.T) {
	// GIVEN
	sqlfile := "test.db"
	os.Remove(sqlfile)
	accessor, err := pizza.NewSQLAccessor(sqlfile)
	require.Nil(t, err)
	defer accessor.Close()
	require.Nil(t, accessor.CreateTables())
	require.Nil(t, accessor.AddFriend("foo@bar.com", "test"))

	// WHEN
	ok, err := accessor.IsFriendAllowed("foo@bar.com")

	// THEN
	assert.Nil(t, err)
	assert.True(t, ok)

	// WHEN
	ok, err = accessor.IsFriendAllowed("bar@bar.com")

	// THEN
	assert.Nil(t, err)
	assert.False(t, ok)
}

func TestSqlAccessor_GetFriendName(t *testing.T) {
	// GIVEN
	sqlfile := "test.db"
	os.Remove(sqlfile)
	accessor, err := pizza.NewSQLAccessor(sqlfile)
	require.Nil(t, err)
	defer accessor.Close()
	require.Nil(t, accessor.CreateTables())
	require.Nil(t, accessor.AddFriend("foo@bar.com", "test"))

	// WHEN
	name, err := accessor.GetFriendName("foo@bar.com")

	// THEN
	assert.Nil(t, err)
	assert.Equal(t, "test", name)

	// WHEN
	name, err = accessor.GetFriendName("bar@bar.com")

	// THEN
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(name))
}

func TestSqlAccessor_GetUpcomingFridays(t *testing.T) {
	// GIVEN
	sqlfile := "test.db"
	os.Remove(sqlfile)
	accessor, err := pizza.NewSQLAccessor(sqlfile)
	require.Nil(t, err)
	defer accessor.Close()
	require.Nil(t, accessor.CreateTables())
	friday1 := time.Now().AddDate(0, 0, 7).UTC()
	require.Nil(t, accessor.AddFriday(friday1))
	friday2 := time.Now().AddDate(0, 0, 14).UTC()
	require.Nil(t, accessor.AddFriday(friday2))
	require.Nil(t, accessor.AddFriday(time.Now().AddDate(0, 0, -2)))
	require.Nil(t, accessor.AddFriday(time.Now().Add(23*time.Hour)))

	// WHEN
	fridays, err := accessor.GetUpcomingFridaysAfter(time.Now().UTC().AddDate(0, 0, 1), 30)
	assert.Nil(t, err)
	assert.NotNil(t, fridays)
	require.Equal(t, 2, len(fridays))
	assert.Equal(t, friday1, fridays[0])
	assert.Equal(t, friday2, fridays[1])
}

func TestSqlAccessor_ListFriends(t *testing.T) {
	// GIVEN
	sqlfile := "test.db"
	os.Remove(sqlfile)
	accessor, err := pizza.NewSQLAccessor(sqlfile)
	require.Nil(t, err)
	defer accessor.Close()
	require.Nil(t, accessor.CreateTables())
	require.Nil(t, accessor.AddFriend("foo@bar.com", "test1"))
	require.Nil(t, accessor.AddFriend("another@better.net", "test2"))

	// WHEN
	friends, err := accessor.ListFriends()

	// THEN
	assert.Nil(t, err)
	require.NotNil(t, friends)
	require.Equal(t, 2, len(friends))
	assert.Equal(t, "foo@bar.com", friends[0].Email)
	assert.Equal(t, "test1", friends[0].Name)
	assert.Equal(t, "another@better.net", friends[1].Email)
	assert.Equal(t, "test2", friends[1].Name)
}
