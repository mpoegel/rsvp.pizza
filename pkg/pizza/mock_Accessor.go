// Code generated by mockery v2.52.3. DO NOT EDIT.

package pizza

import (
	time "time"

	mock "github.com/stretchr/testify/mock"
)

// MockAccessor is an autogenerated mock type for the Accessor type
type MockAccessor struct {
	mock.Mock
}

type MockAccessor_Expecter struct {
	mock *mock.Mock
}

func (_m *MockAccessor) EXPECT() *MockAccessor_Expecter {
	return &MockAccessor_Expecter{mock: &_m.Mock}
}

// AddFriday provides a mock function with given fields: date
func (_m *MockAccessor) AddFriday(date time.Time) error {
	ret := _m.Called(date)

	if len(ret) == 0 {
		panic("no return value specified for AddFriday")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(time.Time) error); ok {
		r0 = rf(date)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockAccessor_AddFriday_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddFriday'
type MockAccessor_AddFriday_Call struct {
	*mock.Call
}

// AddFriday is a helper method to define mock.On call
//   - date time.Time
func (_e *MockAccessor_Expecter) AddFriday(date interface{}) *MockAccessor_AddFriday_Call {
	return &MockAccessor_AddFriday_Call{Call: _e.mock.On("AddFriday", date)}
}

func (_c *MockAccessor_AddFriday_Call) Run(run func(date time.Time)) *MockAccessor_AddFriday_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(time.Time))
	})
	return _c
}

func (_c *MockAccessor_AddFriday_Call) Return(_a0 error) *MockAccessor_AddFriday_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockAccessor_AddFriday_Call) RunAndReturn(run func(time.Time) error) *MockAccessor_AddFriday_Call {
	_c.Call.Return(run)
	return _c
}

// AddFriend provides a mock function with given fields: email, name
func (_m *MockAccessor) AddFriend(email string, name string) error {
	ret := _m.Called(email, name)

	if len(ret) == 0 {
		panic("no return value specified for AddFriend")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(email, name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockAccessor_AddFriend_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddFriend'
type MockAccessor_AddFriend_Call struct {
	*mock.Call
}

// AddFriend is a helper method to define mock.On call
//   - email string
//   - name string
func (_e *MockAccessor_Expecter) AddFriend(email interface{}, name interface{}) *MockAccessor_AddFriend_Call {
	return &MockAccessor_AddFriend_Call{Call: _e.mock.On("AddFriend", email, name)}
}

func (_c *MockAccessor_AddFriend_Call) Run(run func(email string, name string)) *MockAccessor_AddFriend_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *MockAccessor_AddFriend_Call) Return(_a0 error) *MockAccessor_AddFriend_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockAccessor_AddFriend_Call) RunAndReturn(run func(string, string) error) *MockAccessor_AddFriend_Call {
	_c.Call.Return(run)
	return _c
}

// AddFriendToFriday provides a mock function with given fields: email, friday
func (_m *MockAccessor) AddFriendToFriday(email string, friday Friday) error {
	ret := _m.Called(email, friday)

	if len(ret) == 0 {
		panic("no return value specified for AddFriendToFriday")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, Friday) error); ok {
		r0 = rf(email, friday)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockAccessor_AddFriendToFriday_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddFriendToFriday'
type MockAccessor_AddFriendToFriday_Call struct {
	*mock.Call
}

// AddFriendToFriday is a helper method to define mock.On call
//   - email string
//   - friday Friday
func (_e *MockAccessor_Expecter) AddFriendToFriday(email interface{}, friday interface{}) *MockAccessor_AddFriendToFriday_Call {
	return &MockAccessor_AddFriendToFriday_Call{Call: _e.mock.On("AddFriendToFriday", email, friday)}
}

func (_c *MockAccessor_AddFriendToFriday_Call) Run(run func(email string, friday Friday)) *MockAccessor_AddFriendToFriday_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(Friday))
	})
	return _c
}

func (_c *MockAccessor_AddFriendToFriday_Call) Return(_a0 error) *MockAccessor_AddFriendToFriday_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockAccessor_AddFriendToFriday_Call) RunAndReturn(run func(string, Friday) error) *MockAccessor_AddFriendToFriday_Call {
	_c.Call.Return(run)
	return _c
}

// CreateTables provides a mock function with no fields
func (_m *MockAccessor) CreateTables() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for CreateTables")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockAccessor_CreateTables_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateTables'
type MockAccessor_CreateTables_Call struct {
	*mock.Call
}

// CreateTables is a helper method to define mock.On call
func (_e *MockAccessor_Expecter) CreateTables() *MockAccessor_CreateTables_Call {
	return &MockAccessor_CreateTables_Call{Call: _e.mock.On("CreateTables")}
}

func (_c *MockAccessor_CreateTables_Call) Run(run func()) *MockAccessor_CreateTables_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockAccessor_CreateTables_Call) Return(_a0 error) *MockAccessor_CreateTables_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockAccessor_CreateTables_Call) RunAndReturn(run func() error) *MockAccessor_CreateTables_Call {
	_c.Call.Return(run)
	return _c
}

// DoesFridayExist provides a mock function with given fields: date
func (_m *MockAccessor) DoesFridayExist(date time.Time) (bool, error) {
	ret := _m.Called(date)

	if len(ret) == 0 {
		panic("no return value specified for DoesFridayExist")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(time.Time) (bool, error)); ok {
		return rf(date)
	}
	if rf, ok := ret.Get(0).(func(time.Time) bool); ok {
		r0 = rf(date)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(time.Time) error); ok {
		r1 = rf(date)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockAccessor_DoesFridayExist_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DoesFridayExist'
type MockAccessor_DoesFridayExist_Call struct {
	*mock.Call
}

// DoesFridayExist is a helper method to define mock.On call
//   - date time.Time
func (_e *MockAccessor_Expecter) DoesFridayExist(date interface{}) *MockAccessor_DoesFridayExist_Call {
	return &MockAccessor_DoesFridayExist_Call{Call: _e.mock.On("DoesFridayExist", date)}
}

func (_c *MockAccessor_DoesFridayExist_Call) Run(run func(date time.Time)) *MockAccessor_DoesFridayExist_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(time.Time))
	})
	return _c
}

func (_c *MockAccessor_DoesFridayExist_Call) Return(_a0 bool, _a1 error) *MockAccessor_DoesFridayExist_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockAccessor_DoesFridayExist_Call) RunAndReturn(run func(time.Time) (bool, error)) *MockAccessor_DoesFridayExist_Call {
	_c.Call.Return(run)
	return _c
}

// GetFriday provides a mock function with given fields: date
func (_m *MockAccessor) GetFriday(date time.Time) (Friday, error) {
	ret := _m.Called(date)

	if len(ret) == 0 {
		panic("no return value specified for GetFriday")
	}

	var r0 Friday
	var r1 error
	if rf, ok := ret.Get(0).(func(time.Time) (Friday, error)); ok {
		return rf(date)
	}
	if rf, ok := ret.Get(0).(func(time.Time) Friday); ok {
		r0 = rf(date)
	} else {
		r0 = ret.Get(0).(Friday)
	}

	if rf, ok := ret.Get(1).(func(time.Time) error); ok {
		r1 = rf(date)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockAccessor_GetFriday_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetFriday'
type MockAccessor_GetFriday_Call struct {
	*mock.Call
}

// GetFriday is a helper method to define mock.On call
//   - date time.Time
func (_e *MockAccessor_Expecter) GetFriday(date interface{}) *MockAccessor_GetFriday_Call {
	return &MockAccessor_GetFriday_Call{Call: _e.mock.On("GetFriday", date)}
}

func (_c *MockAccessor_GetFriday_Call) Run(run func(date time.Time)) *MockAccessor_GetFriday_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(time.Time))
	})
	return _c
}

func (_c *MockAccessor_GetFriday_Call) Return(_a0 Friday, _a1 error) *MockAccessor_GetFriday_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockAccessor_GetFriday_Call) RunAndReturn(run func(time.Time) (Friday, error)) *MockAccessor_GetFriday_Call {
	_c.Call.Return(run)
	return _c
}

// GetFriendByEmail provides a mock function with given fields: email
func (_m *MockAccessor) GetFriendByEmail(email string) (Friend, error) {
	ret := _m.Called(email)

	if len(ret) == 0 {
		panic("no return value specified for GetFriendByEmail")
	}

	var r0 Friend
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (Friend, error)); ok {
		return rf(email)
	}
	if rf, ok := ret.Get(0).(func(string) Friend); ok {
		r0 = rf(email)
	} else {
		r0 = ret.Get(0).(Friend)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(email)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockAccessor_GetFriendByEmail_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetFriendByEmail'
type MockAccessor_GetFriendByEmail_Call struct {
	*mock.Call
}

// GetFriendByEmail is a helper method to define mock.On call
//   - email string
func (_e *MockAccessor_Expecter) GetFriendByEmail(email interface{}) *MockAccessor_GetFriendByEmail_Call {
	return &MockAccessor_GetFriendByEmail_Call{Call: _e.mock.On("GetFriendByEmail", email)}
}

func (_c *MockAccessor_GetFriendByEmail_Call) Run(run func(email string)) *MockAccessor_GetFriendByEmail_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockAccessor_GetFriendByEmail_Call) Return(_a0 Friend, _a1 error) *MockAccessor_GetFriendByEmail_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockAccessor_GetFriendByEmail_Call) RunAndReturn(run func(string) (Friend, error)) *MockAccessor_GetFriendByEmail_Call {
	_c.Call.Return(run)
	return _c
}

// GetPreferences provides a mock function with given fields: email
func (_m *MockAccessor) GetPreferences(email string) (Preferences, error) {
	ret := _m.Called(email)

	if len(ret) == 0 {
		panic("no return value specified for GetPreferences")
	}

	var r0 Preferences
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (Preferences, error)); ok {
		return rf(email)
	}
	if rf, ok := ret.Get(0).(func(string) Preferences); ok {
		r0 = rf(email)
	} else {
		r0 = ret.Get(0).(Preferences)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(email)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockAccessor_GetPreferences_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPreferences'
type MockAccessor_GetPreferences_Call struct {
	*mock.Call
}

// GetPreferences is a helper method to define mock.On call
//   - email string
func (_e *MockAccessor_Expecter) GetPreferences(email interface{}) *MockAccessor_GetPreferences_Call {
	return &MockAccessor_GetPreferences_Call{Call: _e.mock.On("GetPreferences", email)}
}

func (_c *MockAccessor_GetPreferences_Call) Run(run func(email string)) *MockAccessor_GetPreferences_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockAccessor_GetPreferences_Call) Return(_a0 Preferences, _a1 error) *MockAccessor_GetPreferences_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockAccessor_GetPreferences_Call) RunAndReturn(run func(string) (Preferences, error)) *MockAccessor_GetPreferences_Call {
	_c.Call.Return(run)
	return _c
}

// GetUpcomingFridays provides a mock function with given fields: daysAhead
func (_m *MockAccessor) GetUpcomingFridays(daysAhead int) ([]Friday, error) {
	ret := _m.Called(daysAhead)

	if len(ret) == 0 {
		panic("no return value specified for GetUpcomingFridays")
	}

	var r0 []Friday
	var r1 error
	if rf, ok := ret.Get(0).(func(int) ([]Friday, error)); ok {
		return rf(daysAhead)
	}
	if rf, ok := ret.Get(0).(func(int) []Friday); ok {
		r0 = rf(daysAhead)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]Friday)
		}
	}

	if rf, ok := ret.Get(1).(func(int) error); ok {
		r1 = rf(daysAhead)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockAccessor_GetUpcomingFridays_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetUpcomingFridays'
type MockAccessor_GetUpcomingFridays_Call struct {
	*mock.Call
}

// GetUpcomingFridays is a helper method to define mock.On call
//   - daysAhead int
func (_e *MockAccessor_Expecter) GetUpcomingFridays(daysAhead interface{}) *MockAccessor_GetUpcomingFridays_Call {
	return &MockAccessor_GetUpcomingFridays_Call{Call: _e.mock.On("GetUpcomingFridays", daysAhead)}
}

func (_c *MockAccessor_GetUpcomingFridays_Call) Run(run func(daysAhead int)) *MockAccessor_GetUpcomingFridays_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int))
	})
	return _c
}

func (_c *MockAccessor_GetUpcomingFridays_Call) Return(_a0 []Friday, _a1 error) *MockAccessor_GetUpcomingFridays_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockAccessor_GetUpcomingFridays_Call) RunAndReturn(run func(int) ([]Friday, error)) *MockAccessor_GetUpcomingFridays_Call {
	_c.Call.Return(run)
	return _c
}

// GetUpcomingFridaysAfter provides a mock function with given fields: after, daysAhead
func (_m *MockAccessor) GetUpcomingFridaysAfter(after time.Time, daysAhead int) ([]Friday, error) {
	ret := _m.Called(after, daysAhead)

	if len(ret) == 0 {
		panic("no return value specified for GetUpcomingFridaysAfter")
	}

	var r0 []Friday
	var r1 error
	if rf, ok := ret.Get(0).(func(time.Time, int) ([]Friday, error)); ok {
		return rf(after, daysAhead)
	}
	if rf, ok := ret.Get(0).(func(time.Time, int) []Friday); ok {
		r0 = rf(after, daysAhead)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]Friday)
		}
	}

	if rf, ok := ret.Get(1).(func(time.Time, int) error); ok {
		r1 = rf(after, daysAhead)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockAccessor_GetUpcomingFridaysAfter_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetUpcomingFridaysAfter'
type MockAccessor_GetUpcomingFridaysAfter_Call struct {
	*mock.Call
}

// GetUpcomingFridaysAfter is a helper method to define mock.On call
//   - after time.Time
//   - daysAhead int
func (_e *MockAccessor_Expecter) GetUpcomingFridaysAfter(after interface{}, daysAhead interface{}) *MockAccessor_GetUpcomingFridaysAfter_Call {
	return &MockAccessor_GetUpcomingFridaysAfter_Call{Call: _e.mock.On("GetUpcomingFridaysAfter", after, daysAhead)}
}

func (_c *MockAccessor_GetUpcomingFridaysAfter_Call) Run(run func(after time.Time, daysAhead int)) *MockAccessor_GetUpcomingFridaysAfter_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(time.Time), args[1].(int))
	})
	return _c
}

func (_c *MockAccessor_GetUpcomingFridaysAfter_Call) Return(_a0 []Friday, _a1 error) *MockAccessor_GetUpcomingFridaysAfter_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockAccessor_GetUpcomingFridaysAfter_Call) RunAndReturn(run func(time.Time, int) ([]Friday, error)) *MockAccessor_GetUpcomingFridaysAfter_Call {
	_c.Call.Return(run)
	return _c
}

// ListFridays provides a mock function with no fields
func (_m *MockAccessor) ListFridays() ([]Friday, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ListFridays")
	}

	var r0 []Friday
	var r1 error
	if rf, ok := ret.Get(0).(func() ([]Friday, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() []Friday); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]Friday)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockAccessor_ListFridays_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListFridays'
type MockAccessor_ListFridays_Call struct {
	*mock.Call
}

// ListFridays is a helper method to define mock.On call
func (_e *MockAccessor_Expecter) ListFridays() *MockAccessor_ListFridays_Call {
	return &MockAccessor_ListFridays_Call{Call: _e.mock.On("ListFridays")}
}

func (_c *MockAccessor_ListFridays_Call) Run(run func()) *MockAccessor_ListFridays_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockAccessor_ListFridays_Call) Return(_a0 []Friday, _a1 error) *MockAccessor_ListFridays_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockAccessor_ListFridays_Call) RunAndReturn(run func() ([]Friday, error)) *MockAccessor_ListFridays_Call {
	_c.Call.Return(run)
	return _c
}

// RemoveFriday provides a mock function with given fields: date
func (_m *MockAccessor) RemoveFriday(date time.Time) error {
	ret := _m.Called(date)

	if len(ret) == 0 {
		panic("no return value specified for RemoveFriday")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(time.Time) error); ok {
		r0 = rf(date)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockAccessor_RemoveFriday_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RemoveFriday'
type MockAccessor_RemoveFriday_Call struct {
	*mock.Call
}

// RemoveFriday is a helper method to define mock.On call
//   - date time.Time
func (_e *MockAccessor_Expecter) RemoveFriday(date interface{}) *MockAccessor_RemoveFriday_Call {
	return &MockAccessor_RemoveFriday_Call{Call: _e.mock.On("RemoveFriday", date)}
}

func (_c *MockAccessor_RemoveFriday_Call) Run(run func(date time.Time)) *MockAccessor_RemoveFriday_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(time.Time))
	})
	return _c
}

func (_c *MockAccessor_RemoveFriday_Call) Return(_a0 error) *MockAccessor_RemoveFriday_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockAccessor_RemoveFriday_Call) RunAndReturn(run func(time.Time) error) *MockAccessor_RemoveFriday_Call {
	_c.Call.Return(run)
	return _c
}

// SetPreferences provides a mock function with given fields: email, preferences
func (_m *MockAccessor) SetPreferences(email string, preferences Preferences) error {
	ret := _m.Called(email, preferences)

	if len(ret) == 0 {
		panic("no return value specified for SetPreferences")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, Preferences) error); ok {
		r0 = rf(email, preferences)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockAccessor_SetPreferences_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetPreferences'
type MockAccessor_SetPreferences_Call struct {
	*mock.Call
}

// SetPreferences is a helper method to define mock.On call
//   - email string
//   - preferences Preferences
func (_e *MockAccessor_Expecter) SetPreferences(email interface{}, preferences interface{}) *MockAccessor_SetPreferences_Call {
	return &MockAccessor_SetPreferences_Call{Call: _e.mock.On("SetPreferences", email, preferences)}
}

func (_c *MockAccessor_SetPreferences_Call) Run(run func(email string, preferences Preferences)) *MockAccessor_SetPreferences_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(Preferences))
	})
	return _c
}

func (_c *MockAccessor_SetPreferences_Call) Return(_a0 error) *MockAccessor_SetPreferences_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockAccessor_SetPreferences_Call) RunAndReturn(run func(string, Preferences) error) *MockAccessor_SetPreferences_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateFriday provides a mock function with given fields: friday
func (_m *MockAccessor) UpdateFriday(friday Friday) error {
	ret := _m.Called(friday)

	if len(ret) == 0 {
		panic("no return value specified for UpdateFriday")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(Friday) error); ok {
		r0 = rf(friday)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockAccessor_UpdateFriday_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateFriday'
type MockAccessor_UpdateFriday_Call struct {
	*mock.Call
}

// UpdateFriday is a helper method to define mock.On call
//   - friday Friday
func (_e *MockAccessor_Expecter) UpdateFriday(friday interface{}) *MockAccessor_UpdateFriday_Call {
	return &MockAccessor_UpdateFriday_Call{Call: _e.mock.On("UpdateFriday", friday)}
}

func (_c *MockAccessor_UpdateFriday_Call) Run(run func(friday Friday)) *MockAccessor_UpdateFriday_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(Friday))
	})
	return _c
}

func (_c *MockAccessor_UpdateFriday_Call) Return(_a0 error) *MockAccessor_UpdateFriday_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockAccessor_UpdateFriday_Call) RunAndReturn(run func(Friday) error) *MockAccessor_UpdateFriday_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockAccessor creates a new instance of MockAccessor. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockAccessor(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockAccessor {
	mock := &MockAccessor{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
