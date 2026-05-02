// SPDX-License-Identifier: Apache-2.0

package assert

import "reflect"

// Nil reports whether v is nil, including typed-nil pointers and interfaces
// whose value is nil. The check is reflect-aware: a (*T)(nil) value passed
// as any is detected as nil. On failure reports via t.Errorf.
//
// Concurrency: goroutine-safe.
//
// §21.4 F3
func Nil(t TB, v any, msg ...any) bool {
	t.Helper()
	if isNil(v) {
		return true
	}
	suffix := fmtMsg(msg)
	t.Errorf("Nil failed: expected nil but got %s.%s", formatValue(v), suffix)
	return false
}

// NotNil reports whether v is non-nil. Typed-nil-aware (see Nil). On
// failure reports via t.Errorf.
//
// Concurrency: goroutine-safe.
//
// §21.4 F3
func NotNil(t TB, v any, msg ...any) bool {
	t.Helper()
	if !isNil(v) {
		return true
	}
	suffix := fmtMsg(msg)
	t.Errorf("NotNil failed: expected non-nil value.%s", suffix)
	return false
}

// True reports whether cond is true. On failure reports via t.Errorf.
// msg is optional context appended to the failure message.
//
// Concurrency: goroutine-safe.
//
// §21.4 F3
func True(t TB, cond bool, msg ...any) bool {
	t.Helper()
	if cond {
		return true
	}
	suffix := fmtMsg(msg)
	t.Errorf("True failed: condition is false.%s", suffix)
	return false
}

// False reports whether cond is false. On failure reports via t.Errorf.
//
// Concurrency: goroutine-safe.
//
// §21.4 F3
func False(t TB, cond bool, msg ...any) bool {
	t.Helper()
	if !cond {
		return true
	}
	suffix := fmtMsg(msg)
	t.Errorf("False failed: condition is true.%s", suffix)
	return false
}

// isNil reports whether v is nil, including typed-nil pointers and interfaces.
func isNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map,
		reflect.Chan, reflect.Func:
		return rv.IsNil()
	}
	return false
}

// fmtMsg formats optional message args into a suffix for failure messages.
// Returns " <msg>" if msg is non-empty, otherwise "".
func fmtMsg(msg []any) string {
	if len(msg) == 0 {
		return ""
	}
	s := formatValue(msg[0])
	for _, m := range msg[1:] {
		s += " " + formatValue(m)
	}
	return " " + s
}
