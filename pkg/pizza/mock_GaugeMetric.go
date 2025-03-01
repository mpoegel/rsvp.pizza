// Code generated by mockery v2.52.3. DO NOT EDIT.

package pizza

import mock "github.com/stretchr/testify/mock"

// MockGaugeMetric is an autogenerated mock type for the GaugeMetric type
type MockGaugeMetric struct {
	mock.Mock
}

type MockGaugeMetric_Expecter struct {
	mock *mock.Mock
}

func (_m *MockGaugeMetric) EXPECT() *MockGaugeMetric_Expecter {
	return &MockGaugeMetric_Expecter{mock: &_m.Mock}
}

// Set provides a mock function with given fields: _a0
func (_m *MockGaugeMetric) Set(_a0 float64) {
	_m.Called(_a0)
}

// MockGaugeMetric_Set_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Set'
type MockGaugeMetric_Set_Call struct {
	*mock.Call
}

// Set is a helper method to define mock.On call
//   - _a0 float64
func (_e *MockGaugeMetric_Expecter) Set(_a0 interface{}) *MockGaugeMetric_Set_Call {
	return &MockGaugeMetric_Set_Call{Call: _e.mock.On("Set", _a0)}
}

func (_c *MockGaugeMetric_Set_Call) Run(run func(_a0 float64)) *MockGaugeMetric_Set_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(float64))
	})
	return _c
}

func (_c *MockGaugeMetric_Set_Call) Return() *MockGaugeMetric_Set_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockGaugeMetric_Set_Call) RunAndReturn(run func(float64)) *MockGaugeMetric_Set_Call {
	_c.Run(run)
	return _c
}

// NewMockGaugeMetric creates a new instance of MockGaugeMetric. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockGaugeMetric(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockGaugeMetric {
	mock := &MockGaugeMetric{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
