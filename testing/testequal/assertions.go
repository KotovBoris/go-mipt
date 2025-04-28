//go:build !solution

package testequal

import (
	"fmt"
	"reflect"
)

// AssertEqual checks that expected and actual are equal.
//
// Marks caller function as having failed but continues execution.
//
// Returns true iff arguments are equal.
func AssertEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()

	switch expected.(type) { // type switch: val.(type) определяет динамический тип val
	case struct{}:
		return false
	}

	result := reflect.DeepEqual(expected, actual)

	message := fmt.Sprintf(`equal:\nexpected: %v\nactual  : %v\nmessage : %v != %v`, expected, actual, expected, actual)
	if len(msgAndArgs) > 0 {
		message += msgAndArgs[0].(string)
		message = fmt.Sprintf(message, msgAndArgs[1:]...)
	}

	if !result {
		t.Errorf(message)
	}

	return result
}

// AssertNotEqual checks that expected and actual are not equal.
//
// Marks caller function as having failed but continues execution.
//
// Returns true iff arguments are not equal.
func AssertNotEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()

	switch expected.(type) { // type switch: val.(type) определяет динамический тип val
	case struct{}:
		return true
	}

	result := !reflect.DeepEqual(expected, actual)

	message := fmt.Sprintf(`not equal:\nexpected: %v\nactual  : %v\nmessage : %v != %v`, expected, actual, expected, actual)
	if len(msgAndArgs) > 0 {
		message += msgAndArgs[0].(string)
		message = fmt.Sprintf(message, msgAndArgs[1:]...)
	}

	if !result {
		t.Errorf(message)
	}

	return result
}

// RequireEqual does the same as AssertEqual but fails caller test immediately.
func RequireEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()

	if !AssertEqual(t, expected, actual, msgAndArgs...) {
		t.FailNow()
	}
}

// RequireNotEqual does the same as AssertNotEqual but fails caller test immediately.
func RequireNotEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()

	if !AssertNotEqual(t, expected, actual, msgAndArgs...) {
		t.FailNow()
	}
}
