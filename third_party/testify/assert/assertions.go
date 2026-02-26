// Package assert provides assertion functions for use in tests.
package assert

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// TestingT is an interface wrapper around *testing.T
type TestingT interface {
	Errorf(format string, args ...interface{})
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
func Equal(t TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	helper(t)
	if !objectsAreEqual(expected, actual) {
		t.Errorf("Not equal: \n"+
			"expected: %v\n"+
			"actual  : %v\n%s", expected, actual, formatMsgAndArgs(msgAndArgs...))
		return false
	}
	return true
}

// NotEqual asserts that two objects are not equal.
func NotEqual(t TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	helper(t)
	if objectsAreEqual(expected, actual) {
		t.Errorf("Should not be equal: %v\n%s", actual, formatMsgAndArgs(msgAndArgs...))
		return false
	}
	return true
}

// NoError asserts that an error is nil.
func NoError(t TestingT, err error, msgAndArgs ...interface{}) bool {
	helper(t)
	if err != nil {
		t.Errorf("Received unexpected error:\n%+v\n%s", err, formatMsgAndArgs(msgAndArgs...))
		return false
	}
	return true
}

// Error asserts that an error is not nil.
func Error(t TestingT, err error, msgAndArgs ...interface{}) bool {
	helper(t)
	if err == nil {
		t.Errorf("An error is expected but got nil.\n%s", formatMsgAndArgs(msgAndArgs...))
		return false
	}
	return true
}

// Nil asserts that the specified object is nil.
func Nil(t TestingT, object interface{}, msgAndArgs ...interface{}) bool {
	helper(t)
	if isNil(object) {
		return true
	}
	t.Errorf("Expected nil, but got: %#v\n%s", object, formatMsgAndArgs(msgAndArgs...))
	return false
}

// NotNil asserts that the specified object is NOT nil.
func NotNil(t TestingT, object interface{}, msgAndArgs ...interface{}) bool {
	helper(t)
	if !isNil(object) {
		return true
	}
	t.Errorf("Expected value not to be nil.\n%s", formatMsgAndArgs(msgAndArgs...))
	return false
}

// True asserts that the specified value is true.
func True(t TestingT, value bool, msgAndArgs ...interface{}) bool {
	helper(t)
	if !value {
		t.Errorf("Should be true\n%s", formatMsgAndArgs(msgAndArgs...))
		return false
	}
	return true
}

// False asserts that the specified value is false.
func False(t TestingT, value bool, msgAndArgs ...interface{}) bool {
	helper(t)
	if value {
		t.Errorf("Should be false\n%s", formatMsgAndArgs(msgAndArgs...))
		return false
	}
	return true
}

// Contains asserts that the specified string contains the specified substring.
func Contains(t TestingT, s, contains interface{}, msgAndArgs ...interface{}) bool {
	helper(t)
	ok, found := includeElement(s, contains)
	if !ok {
		t.Errorf("%#v could not be applied builtin len()\n%s", s, formatMsgAndArgs(msgAndArgs...))
		return false
	}
	if !found {
		t.Errorf("%#v does not contain %#v\n%s", s, contains, formatMsgAndArgs(msgAndArgs...))
		return false
	}
	return true
}

// NotContains asserts that the specified string does NOT contain the specified substring.
func NotContains(t TestingT, s, contains interface{}, msgAndArgs ...interface{}) bool {
	helper(t)
	ok, found := includeElement(s, contains)
	if !ok {
		t.Errorf("%#v could not be applied builtin len()\n%s", s, formatMsgAndArgs(msgAndArgs...))
		return false
	}
	if found {
		t.Errorf("%#v should not contain %#v\n%s", s, contains, formatMsgAndArgs(msgAndArgs...))
		return false
	}
	return true
}

// Len asserts that the specified object has specific length.
func Len(t TestingT, object interface{}, length int, msgAndArgs ...interface{}) bool {
	helper(t)
	ok, l := getLen(object)
	if !ok {
		t.Errorf("\"%v\" could not be applied builtin len()\n%s", object, formatMsgAndArgs(msgAndArgs...))
		return false
	}
	if l != length {
		t.Errorf("\"%v\" should have %d item(s), but has %d\n%s", object, length, l, formatMsgAndArgs(msgAndArgs...))
		return false
	}
	return true
}

// Greater asserts that the first element is greater than the second.
func Greater(t TestingT, e1 interface{}, e2 interface{}, msgAndArgs ...interface{}) bool {
	helper(t)
	return compareTwoValues(t, e1, e2, []CompareType{compareGreater}, "\"%v\" is not greater than \"%v\"", msgAndArgs...)
}

// GreaterOrEqual asserts that the first element is greater than or equal to the second.
func GreaterOrEqual(t TestingT, e1 interface{}, e2 interface{}, msgAndArgs ...interface{}) bool {
	helper(t)
	return compareTwoValues(t, e1, e2, []CompareType{compareGreater, compareEqual}, "\"%v\" is not greater than or equal to \"%v\"", msgAndArgs...)
}

// Less asserts that the first element is less than the second.
func Less(t TestingT, e1 interface{}, e2 interface{}, msgAndArgs ...interface{}) bool {
	helper(t)
	return compareTwoValues(t, e1, e2, []CompareType{compareLess}, "\"%v\" is not less than \"%v\"", msgAndArgs...)
}

// LessOrEqual asserts that the first element is less than or equal to the second.
func LessOrEqual(t TestingT, e1 interface{}, e2 interface{}, msgAndArgs ...interface{}) bool {
	helper(t)
	return compareTwoValues(t, e1, e2, []CompareType{compareLess, compareEqual}, "\"%v\" is not less than or equal to \"%v\"", msgAndArgs...)
}

// ErrorIs asserts that errors.Is(err, target) is true.
func ErrorIs(t TestingT, err, target error, msgAndArgs ...interface{}) bool {
	helper(t)
	if errors.Is(err, target) {
		return true
	}
	t.Errorf("Target error should be in err chain:\n"+
		"expected : %q\n"+
		"in chain : %s\n%s",
		target, buildErrorChainString(err), formatMsgAndArgs(msgAndArgs...))
	return false
}

func buildErrorChainString(err error) string {
	if err == nil {
		return ""
	}
	var chain []string
	for e := err; e != nil; e = errors.Unwrap(e) {
		chain = append(chain, fmt.Sprintf("%q", e))
	}
	return strings.Join(chain, "\n\t")
}

// NotEmpty asserts that the specified object is NOT empty.
func NotEmpty(t TestingT, object interface{}, msgAndArgs ...interface{}) bool {
	helper(t)
	pass := !isEmpty(object)
	if !pass {
		t.Errorf("Should NOT be empty, but was\n%s", formatMsgAndArgs(msgAndArgs...))
	}
	return pass
}

// Empty asserts that the specified object is empty.
func Empty(t TestingT, object interface{}, msgAndArgs ...interface{}) bool {
	helper(t)
	pass := isEmpty(object)
	if !pass {
		t.Errorf("Should be empty, but was %v\n%s", object, formatMsgAndArgs(msgAndArgs...))
	}
	return pass
}

// EqualError asserts that a function returned an error (i.e. not nil) and that it is equal to the provided error.
func EqualError(t TestingT, theError error, errString string, msgAndArgs ...interface{}) bool {
	helper(t)
	if !Error(t, theError, msgAndArgs...) {
		return false
	}
	expected := errString
	actual := theError.Error()
	if expected != actual {
		t.Errorf("Error message not equal:\nexpected: %q\nactual  : %q\n%s", expected, actual, formatMsgAndArgs(msgAndArgs...))
		return false
	}
	return true
}

// ErrorContains checks that a function returned an error with a message that contains the string.
func ErrorContains(t TestingT, theError error, contains string, msgAndArgs ...interface{}) bool {
	helper(t)
	if !Error(t, theError, msgAndArgs...) {
		return false
	}
	if !strings.Contains(theError.Error(), contains) {
		t.Errorf("Error %#v does not contain %#v\n%s", theError, contains, formatMsgAndArgs(msgAndArgs...))
		return false
	}
	return true
}

// Fail reports a failure through.
func Fail(t TestingT, failureMessage string, msgAndArgs ...interface{}) bool {
	helper(t)
	t.Errorf("%s\n%s", failureMessage, formatMsgAndArgs(msgAndArgs...))
	return false
}

// FailNow fails test.
func FailNow(t TestingT, failureMessage string, msgAndArgs ...interface{}) bool {
	helper(t)
	Fail(t, failureMessage, msgAndArgs...)
	// We call FailNow on T if available
	if ft, ok := t.(interface{ FailNow() }); ok {
		ft.FailNow()
	}
	return false
}

// --- helpers ---

func objectsAreEqual(expected, actual interface{}) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}
	exp, ok := expected.([]byte)
	if !ok {
		return reflect.DeepEqual(expected, actual)
	}
	act, ok := actual.([]byte)
	if !ok {
		return false
	}
	if exp == nil || act == nil {
		return exp == nil && act == nil
	}
	return reflect.DeepEqual(exp, act)
}

func isNil(object interface{}) bool {
	if object == nil {
		return true
	}
	value := reflect.ValueOf(object)
	kind := value.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice && value.IsNil() {
		return true
	}
	return false
}

func isEmpty(object interface{}) bool {
	if object == nil {
		return true
	}
	objValue := reflect.ValueOf(object)
	switch objValue.Kind() {
	case reflect.Chan, reflect.Map, reflect.Slice:
		return objValue.Len() == 0
	case reflect.Ptr:
		if objValue.IsNil() {
			return true
		}
		deref := objValue.Elem().Interface()
		return isEmpty(deref)
	default:
		zero := reflect.Zero(objValue.Type())
		return reflect.DeepEqual(object, zero.Interface())
	}
}

func getLen(x interface{}) (ok bool, length int) {
	v := reflect.ValueOf(x)
	defer func() {
		if e := recover(); e != nil {
			ok = false
		}
	}()
	return true, v.Len()
}

func includeElement(list interface{}, element interface{}) (ok, found bool) {
	listValue := reflect.ValueOf(list)
	listKind := listValue.Type().Kind()
	if listKind == reflect.String {
		elementValue := reflect.ValueOf(element)
		return true, strings.Contains(listValue.String(), elementValue.String())
	}
	if listKind == reflect.Map {
		mapKeys := listValue.MapKeys()
		for i := 0; i < len(mapKeys); i++ {
			if objectsAreEqual(mapKeys[i].Interface(), element) {
				return true, true
			}
		}
		return true, false
	}
	for i := 0; i < listValue.Len(); i++ {
		if objectsAreEqual(listValue.Index(i).Interface(), element) {
			return true, true
		}
	}
	return true, false
}

type CompareType int

const (
	compareLess    CompareType = -1
	compareEqual   CompareType = 0
	compareGreater CompareType = 1
)

func compareTwoValues(t TestingT, e1 interface{}, e2 interface{}, allowedComparesResults []CompareType, failMessage string, msgAndArgs ...interface{}) bool {
	helper(t)
	e1Kind := reflect.ValueOf(e1).Kind()
	e2Kind := reflect.ValueOf(e2).Kind()
	if e1Kind != e2Kind {
		t.Errorf("Elements should be the same type\n%s", formatMsgAndArgs(msgAndArgs...))
		return false
	}
	compareResult, isComparable := compare(e1, e2, e1Kind)
	if !isComparable {
		t.Errorf("Can not compare type \"%s\"\n%s", reflect.TypeOf(e1), formatMsgAndArgs(msgAndArgs...))
		return false
	}
	for _, allowedResult := range allowedComparesResults {
		if compareResult == allowedResult {
			return true
		}
	}
	t.Errorf(failMessage+"\n%s", e1, e2, formatMsgAndArgs(msgAndArgs...))
	return false
}

func compare(obj1, obj2 interface{}, kind reflect.Kind) (CompareType, bool) {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		{
			intobj1, ok := obj1.(int64)
			if !ok {
				intobj1 = reflect.ValueOf(obj1).Int()
			}
			intobj2, ok := obj2.(int64)
			if !ok {
				intobj2 = reflect.ValueOf(obj2).Int()
			}
			if intobj1 > intobj2 {
				return compareGreater, true
			}
			if intobj1 == intobj2 {
				return compareEqual, true
			}
			if intobj1 < intobj2 {
				return compareLess, true
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		{
			uintobj1, ok := obj1.(uint64)
			if !ok {
				uintobj1 = reflect.ValueOf(obj1).Uint()
			}
			uintobj2, ok := obj2.(uint64)
			if !ok {
				uintobj2 = reflect.ValueOf(obj2).Uint()
			}
			if uintobj1 > uintobj2 {
				return compareGreater, true
			}
			if uintobj1 == uintobj2 {
				return compareEqual, true
			}
			if uintobj1 < uintobj2 {
				return compareLess, true
			}
		}
	case reflect.Float32, reflect.Float64:
		{
			floatobj1, ok := obj1.(float64)
			if !ok {
				floatobj1 = reflect.ValueOf(obj1).Float()
			}
			floatobj2, ok := obj2.(float64)
			if !ok {
				floatobj2 = reflect.ValueOf(obj2).Float()
			}
			if floatobj1 > floatobj2 {
				return compareGreater, true
			}
			if floatobj1 == floatobj2 {
				return compareEqual, true
			}
			if floatobj1 < floatobj2 {
				return compareLess, true
			}
		}
	case reflect.String:
		{
			stringobj1, ok := obj1.(string)
			if !ok {
				stringobj1 = reflect.ValueOf(obj1).String()
			}
			stringobj2, ok := obj2.(string)
			if !ok {
				stringobj2 = reflect.ValueOf(obj2).String()
			}
			if stringobj1 > stringobj2 {
				return compareGreater, true
			}
			if stringobj1 == stringobj2 {
				return compareEqual, true
			}
			if stringobj1 < stringobj2 {
				return compareLess, true
			}
		}
	}
	return compareEqual, false
}

func formatMsgAndArgs(msgAndArgs ...interface{}) string {
	if len(msgAndArgs) == 0 {
		return ""
	}
	if len(msgAndArgs) == 1 {
		msg := msgAndArgs[0]
		if msgAsStr, ok := msg.(string); ok {
			return msgAsStr
		}
		return fmt.Sprintf("%+v", msg)
	}
	if len(msgAndArgs) > 1 {
		return fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
	return ""
}

// Assertions provides assertion methods around the TestingT interface.
type Assertions struct {
	t TestingT
}

// New makes a new Assertions object for the specified TestingT.
func New(t TestingT) *Assertions {
	return &Assertions{t: t}
}

func (a *Assertions) Equal(expected, actual interface{}, msgAndArgs ...interface{}) bool {
	return Equal(a.t, expected, actual, msgAndArgs...)
}
func (a *Assertions) NotEqual(expected, actual interface{}, msgAndArgs ...interface{}) bool {
	return NotEqual(a.t, expected, actual, msgAndArgs...)
}
func (a *Assertions) NoError(err error, msgAndArgs ...interface{}) bool {
	return NoError(a.t, err, msgAndArgs...)
}
func (a *Assertions) Error(err error, msgAndArgs ...interface{}) bool {
	return Error(a.t, err, msgAndArgs...)
}
func (a *Assertions) Nil(object interface{}, msgAndArgs ...interface{}) bool {
	return Nil(a.t, object, msgAndArgs...)
}
func (a *Assertions) NotNil(object interface{}, msgAndArgs ...interface{}) bool {
	return NotNil(a.t, object, msgAndArgs...)
}
func (a *Assertions) True(value bool, msgAndArgs ...interface{}) bool {
	return True(a.t, value, msgAndArgs...)
}
func (a *Assertions) False(value bool, msgAndArgs ...interface{}) bool {
	return False(a.t, value, msgAndArgs...)
}
func (a *Assertions) Contains(s, contains interface{}, msgAndArgs ...interface{}) bool {
	return Contains(a.t, s, contains, msgAndArgs...)
}
func (a *Assertions) NotContains(s, contains interface{}, msgAndArgs ...interface{}) bool {
	return NotContains(a.t, s, contains, msgAndArgs...)
}
func (a *Assertions) Len(object interface{}, length int, msgAndArgs ...interface{}) bool {
	return Len(a.t, object, length, msgAndArgs...)
}
func (a *Assertions) Greater(e1, e2 interface{}, msgAndArgs ...interface{}) bool {
	return Greater(a.t, e1, e2, msgAndArgs...)
}
func (a *Assertions) NotEmpty(object interface{}, msgAndArgs ...interface{}) bool {
	return NotEmpty(a.t, object, msgAndArgs...)
}
func (a *Assertions) Empty(object interface{}, msgAndArgs ...interface{}) bool {
	return Empty(a.t, object, msgAndArgs...)
}
func (a *Assertions) EqualError(theError error, errString string, msgAndArgs ...interface{}) bool {
	return EqualError(a.t, theError, errString, msgAndArgs...)
}
func (a *Assertions) ErrorContains(theError error, contains string, msgAndArgs ...interface{}) bool {
	return ErrorContains(a.t, theError, contains, msgAndArgs...)
}
func (a *Assertions) Fail(failureMessage string, msgAndArgs ...interface{}) bool {
	return Fail(a.t, failureMessage, msgAndArgs...)
}
func (a *Assertions) FailNow(failureMessage string, msgAndArgs ...interface{}) bool {
	return FailNow(a.t, failureMessage, msgAndArgs...)
}
func (a *Assertions) LessOrEqual(e1, e2 interface{}, msgAndArgs ...interface{}) bool {
	return LessOrEqual(a.t, e1, e2, msgAndArgs...)
}
func (a *Assertions) ErrorIs(err, target error, msgAndArgs ...interface{}) bool {
	return ErrorIs(a.t, err, target, msgAndArgs...)
}

// --- testing.T wrappers ---
type tTesting struct {
	t *testing.T
}

func (tt *tTesting) Errorf(format string, args ...interface{}) {
	tt.t.Helper()
	tt.t.Errorf(format, args...)
}
