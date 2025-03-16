package pizza_test

import (
	"os"
	"testing"
	"time"

	"github.com/mpoegel/rsvp.pizza/pkg/pizza"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSqlAccessor_GetFriendByEmail(t *testing.T) {
	// GIVEN
	sqlfile := "test.db"
	os.Remove(sqlfile)
	defer os.Remove(sqlfile)
	accessor, err := pizza.NewSQLAccessor(sqlfile, true)
	require.Nil(t, err)
	defer accessor.Close()
	require.Nil(t, accessor.CreateTables())
	require.Nil(t, accessor.AddFriend("foo@bar.com", "test"))

	// WHEN
	friend, err := accessor.GetFriendByEmail("foo@bar.com")

	// THEN
	assert.Nil(t, err)
	assert.Equal(t, "1", friend.ID)
	assert.Equal(t, "test", friend.Name)
	assert.Equal(t, "foo@bar.com", friend.Email)

	// WHEN
	friend, err = accessor.GetFriendByID("1")
	assert.Nil(t, err)
	assert.Equal(t, "1", friend.ID)
	assert.Equal(t, "test", friend.Name)
	assert.Equal(t, "foo@bar.com", friend.Email)

	// WHEN
	friend, err = accessor.GetFriendByEmail("bar@bar.com")

	// THEN
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(friend.Name))
}

func TestSqlAccessor_GetUpcomingFridays(t *testing.T) {
	// GIVEN
	sqlfile := "test.db"
	os.Remove(sqlfile)
	defer os.Remove(sqlfile)
	accessor, err := pizza.NewSQLAccessor(sqlfile, true)
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
	assert.Equal(t, friday1, fridays[0].Date)
	assert.Equal(t, friday2, fridays[1].Date)
}

func TestSqlAccessor_ListFridays(t *testing.T) {
	// GIVEN
	sqlfile := "test.db"
	os.Remove(sqlfile)
	defer os.Remove(sqlfile)
	accessor, err := pizza.NewSQLAccessor(sqlfile, true)
	require.Nil(t, err)
	defer accessor.Close()
	require.Nil(t, accessor.CreateTables())
	loc, _ := time.LoadLocation("America/New_York")
	f1 := time.Date(2023, 12, 22, 17, 30, 0, 0, loc)
	f2 := time.Date(2023, 12, 29, 17, 30, 0, 0, loc)
	require.Nil(t, accessor.AddFriday(f1))
	require.Nil(t, accessor.AddFriday(f2))

	// WHEN
	fridays, err := accessor.ListFridays()

	// THEN
	assert.Nil(t, err)
	require.NotNil(t, fridays)
	require.Equal(t, 2, len(fridays))
	assert.Equal(t, f1, fridays[0].Date)
	assert.Equal(t, f2, fridays[1].Date)
}

func TestSqlAccessor_RemoveFriday(t *testing.T) {
	// GIVEN
	sqlfile := "test.db"
	os.Remove(sqlfile)
	defer os.Remove(sqlfile)
	accessor, err := pizza.NewSQLAccessor(sqlfile, true)
	require.Nil(t, err)
	defer accessor.Close()
	require.Nil(t, accessor.CreateTables())
	loc, _ := time.LoadLocation("America/New_York")
	f1 := time.Date(2023, 12, 22, 17, 30, 0, 0, loc)
	f2 := time.Date(2023, 12, 29, 17, 30, 0, 0, loc)
	require.Nil(t, accessor.AddFriday(f1))
	require.Nil(t, accessor.AddFriday(f2))

	// WHEN
	err = accessor.RemoveFriday(f1)
	fridays, err2 := accessor.ListFridays()

	// THEN
	assert.Nil(t, err)
	assert.Nil(t, err2)
	require.NotNil(t, fridays)
	require.Equal(t, 1, len(fridays))
	assert.Equal(t, f2, fridays[0].Date)
}

func TestSqlAccessor_AddAndRemoveFriendFriday(t *testing.T) {
	// GIVEN
	sqlfile := "test.db"
	os.Remove(sqlfile)
	defer os.Remove(sqlfile)
	accessor, err := pizza.NewSQLAccessor(sqlfile, true)
	require.Nil(t, err)
	defer accessor.Close()
	require.Nil(t, accessor.CreateTables())
	loc, _ := time.LoadLocation("America/New_York")
	f1 := time.Date(2023, 12, 22, 17, 30, 0, 0, loc)
	require.Nil(t, accessor.AddFriday(f1))

	// WHEN
	assert.Nil(t, accessor.AddFriendToFriday("foo", pizza.Friday{Date: f1}))
	assert.Nil(t, accessor.AddFriendToFriday("bar", pizza.Friday{Date: f1}))
	friday, err := accessor.GetFriday(f1)

	// THEN
	assert.Nil(t, err)
	assert.Equal(t, 2, len(friday.Guests))
	assert.Equal(t, "foo", friday.Guests[0])
	assert.Equal(t, "bar", friday.Guests[1])

	// WHEN
	assert.Nil(t, accessor.RemoveFriendFromFriday("foo", f1))
	friday, err = accessor.GetFriday(f1)

	// THEN
	assert.Nil(t, err)
	assert.Equal(t, 1, len(friday.Guests))
	assert.Equal(t, "bar", friday.Guests[0])
}
