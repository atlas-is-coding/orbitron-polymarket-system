// Package require implements the same assertions as the assert package
// but stops test execution when a test fails.
package require

import (
	"github.com/stretchr/testify/assert"
)

// TestingT is an interface wrapper around *testing.T
type TestingT interface {
	Errorf(format string, args ...interface{})
	FailNow()
}

type tHelper interface {
	Helper()
}

func helper(t TestingT) {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
}

// Equal asserts that two objects are equal.
func Equal(t TestingT, expected, actual interface{}, msgAndArgs ...interface{}) {
	helper(t)
	if !assert.Equal(t, expected, actual, msgAndArgs...) {
		t.FailNow()
	}
}

// NotEqual asserts that two objects are not equal.
func NotEqual(t TestingT, expected, actual interface{}, msgAndArgs ...interface{}) {
	helper(t)
	if !assert.NotEqual(t, expected, actual, msgAndArgs...) {
		t.FailNow()
	}
}

// NoError asserts that an error is nil.
func NoError(t TestingT, err error, msgAndArgs ...interface{}) {
	helper(t)
	if !assert.NoError(t, err, msgAndArgs...) {
		t.FailNow()
	}
}

// Error asserts that an error is not nil.
func Error(t TestingT, err error, msgAndArgs ...interface{}) {
	helper(t)
	if !assert.Error(t, err, msgAndArgs...) {
		t.FailNow()
	}
}

// Nil asserts that the specified object is nil.
func Nil(t TestingT, object interface{}, msgAndArgs ...interface{}) {
	helper(t)
	if !assert.Nil(t, object, msgAndArgs...) {
		t.FailNow()
	}
}

// NotNil asserts that the specified object is NOT nil.
func NotNil(t TestingT, object interface{}, msgAndArgs ...interface{}) {
	helper(t)
	if !assert.NotNil(t, object, msgAndArgs...) {
		t.FailNow()
	}
}

// True asserts that the specified value is true.
func True(t TestingT, value bool, msgAndArgs ...interface{}) {
	helper(t)
	if !assert.True(t, value, msgAndArgs...) {
		t.FailNow()
	}
}

// False asserts that the specified value is false.
func False(t TestingT, value bool, msgAndArgs ...interface{}) {
	helper(t)
	if !assert.False(t, value, msgAndArgs...) {
		t.FailNow()
	}
}

// Contains asserts that the specified string contains the specified substring.
func Contains(t TestingT, s, contains interface{}, msgAndArgs ...interface{}) {
	helper(t)
	if !assert.Contains(t, s, contains, msgAndArgs...) {
		t.FailNow()
	}
}

// Len asserts that the specified object has specific length.
func Len(t TestingT, object interface{}, length int, msgAndArgs ...interface{}) {
	helper(t)
	if !assert.Len(t, object, length, msgAndArgs...) {
		t.FailNow()
	}
}

// Greater asserts that the first element is greater than the second.
func Greater(t TestingT, e1 interface{}, e2 interface{}, msgAndArgs ...interface{}) {
	helper(t)
	if !assert.Greater(t, e1, e2, msgAndArgs...) {
		t.FailNow()
	}
}

// LessOrEqual asserts that the first element is less than or equal to the second.
func LessOrEqual(t TestingT, e1 interface{}, e2 interface{}, msgAndArgs ...interface{}) {
	helper(t)
	if !assert.LessOrEqual(t, e1, e2, msgAndArgs...) {
		t.FailNow()
	}
}

// ErrorIs asserts that errors.Is(err, target) is true.
func ErrorIs(t TestingT, err, target error, msgAndArgs ...interface{}) {
	helper(t)
	if !assert.ErrorIs(t, err, target, msgAndArgs...) {
		t.FailNow()
	}
}

// NotEmpty asserts that the specified object is NOT empty.
func NotEmpty(t TestingT, object interface{}, msgAndArgs ...interface{}) {
	helper(t)
	if !assert.NotEmpty(t, object, msgAndArgs...) {
		t.FailNow()
	}
}

// Empty asserts that the specified object is empty.
func Empty(t TestingT, object interface{}, msgAndArgs ...interface{}) {
	helper(t)
	if !assert.Empty(t, object, msgAndArgs...) {
		t.FailNow()
	}
}

// EqualError asserts that a function returned an error equal to the provided string.
func EqualError(t TestingT, theError error, errString string, msgAndArgs ...interface{}) {
	helper(t)
	if !assert.EqualError(t, theError, errString, msgAndArgs...) {
		t.FailNow()
	}
}

// ErrorContains checks that a function returned an error with a message that contains the string.
func ErrorContains(t TestingT, theError error, contains string, msgAndArgs ...interface{}) {
	helper(t)
	if !assert.ErrorContains(t, theError, contains, msgAndArgs...) {
		t.FailNow()
	}
}

// Fail reports a failure through.
func Fail(t TestingT, failureMessage string, msgAndArgs ...interface{}) {
	helper(t)
	assert.Fail(t, failureMessage, msgAndArgs...)
	t.FailNow()
}

// Assertions provides assertion methods around the TestingT interface.
type Assertions struct {
	t TestingT
}

// New makes a new Assertions object for the specified TestingT.
func New(t TestingT) *Assertions {
	return &Assertions{t: t}
}

func (a *Assertions) Equal(expected, actual interface{}, msgAndArgs ...interface{}) {
	Equal(a.t, expected, actual, msgAndArgs...)
}
func (a *Assertions) NotEqual(expected, actual interface{}, msgAndArgs ...interface{}) {
	NotEqual(a.t, expected, actual, msgAndArgs...)
}
func (a *Assertions) NoError(err error, msgAndArgs ...interface{}) {
	NoError(a.t, err, msgAndArgs...)
}
func (a *Assertions) Error(err error, msgAndArgs ...interface{}) {
	Error(a.t, err, msgAndArgs...)
}
func (a *Assertions) Nil(object interface{}, msgAndArgs ...interface{}) {
	Nil(a.t, object, msgAndArgs...)
}
func (a *Assertions) NotNil(object interface{}, msgAndArgs ...interface{}) {
	NotNil(a.t, object, msgAndArgs...)
}
func (a *Assertions) True(value bool, msgAndArgs ...interface{}) {
	True(a.t, value, msgAndArgs...)
}
func (a *Assertions) False(value bool, msgAndArgs ...interface{}) {
	False(a.t, value, msgAndArgs...)
}
func (a *Assertions) Contains(s, contains interface{}, msgAndArgs ...interface{}) {
	Contains(a.t, s, contains, msgAndArgs...)
}
func (a *Assertions) Len(object interface{}, length int, msgAndArgs ...interface{}) {
	Len(a.t, object, length, msgAndArgs...)
}
func (a *Assertions) Greater(e1, e2 interface{}, msgAndArgs ...interface{}) {
	Greater(a.t, e1, e2, msgAndArgs...)
}
func (a *Assertions) NotEmpty(object interface{}, msgAndArgs ...interface{}) {
	NotEmpty(a.t, object, msgAndArgs...)
}
func (a *Assertions) Empty(object interface{}, msgAndArgs ...interface{}) {
	Empty(a.t, object, msgAndArgs...)
}
func (a *Assertions) EqualError(theError error, errString string, msgAndArgs ...interface{}) {
	EqualError(a.t, theError, errString, msgAndArgs...)
}
func (a *Assertions) ErrorContains(theError error, contains string, msgAndArgs ...interface{}) {
	ErrorContains(a.t, theError, contains, msgAndArgs...)
}
func (a *Assertions) Fail(failureMessage string, msgAndArgs ...interface{}) {
	Fail(a.t, failureMessage, msgAndArgs...)
}
func (a *Assertions) LessOrEqual(e1, e2 interface{}, msgAndArgs ...interface{}) {
	LessOrEqual(a.t, e1, e2, msgAndArgs...)
}
func (a *Assertions) ErrorIs(err, target error, msgAndArgs ...interface{}) {
	ErrorIs(a.t, err, target, msgAndArgs...)
}
