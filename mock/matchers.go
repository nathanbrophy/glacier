// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"fmt"
	"reflect"

	"github.com/nathanbrophy/glacier/assert"
)

// anyMatcher is the internal untyped interface used within Expectation.
// Matcher[T] implements anyMatcher via its unexported methods. The type
// assertion is checked at registration.
type anyMatcher interface {
	matchAny(v any) bool
	String() string
}

// Matcher[T] is a type-safe predicate over values of type T.
// It is the unit of argument matching in the expectation builder.
//
// invariant: match is never nil.
// invariant: String() returns a stable, human-readable description.
type Matcher[T any] struct {
	match func(T) bool // invariant: non-nil
	desc  string       // invariant: stable across calls
}

// Match reports whether v satisfies the matcher.
func (m Matcher[T]) Match(v T) bool { return m.match(v) }

// String returns a human-readable description of this matcher.
func (m Matcher[T]) String() string { return m.desc }

// matchAny implements anyMatcher. It type-asserts v to T; if the assertion
// fails, the matcher returns false (type mismatch → no match).
func (m Matcher[T]) matchAny(v any) bool {
	// Handle nil interface case.
	if v == nil {
		var zero T
		rv := reflect.ValueOf(&zero).Elem()
		// If T is a reference type (pointer, interface, slice, map, func, chan),
		// treat nil as matching a zero value.
		switch rv.Kind() {
		case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map,
			reflect.Func, reflect.Chan:
			return m.match(zero)
		}
		return false
	}
	typed, ok := v.(T)
	if !ok {
		return false
	}
	return m.match(typed)
}

// Eq returns a Matcher[T] that passes only when the argument is deeply equal
// to want (using reflect.DeepEqual).
//
// Example: Eq[string]("alice") matches only the string "alice".
func Eq[T any](want T) Matcher[T] {
	return Matcher[T]{
		match: func(got T) bool { return reflect.DeepEqual(got, want) },
		desc:  fmt.Sprintf("Eq(%v)", want),
	}
}

// Any returns a Matcher[T] that passes for every value of type T.
//
// Example: Any[int]() matches any int argument.
func Any[T any]() Matcher[T] {
	var zero T
	return Matcher[T]{
		match: func(T) bool { return true },
		desc:  fmt.Sprintf("Any[%T]", zero),
	}
}

// Pred returns a Matcher[T] backed by a custom predicate function.
// fn must be non-nil; Pred panics (library-register format) otherwise.
//
// Example: Pred[User](func(u User) bool { return u.ID > 0 })
func Pred[T any](fn func(T) bool) Matcher[T] {
	if fn == nil {
		panic("mock.Pred: predicate function must not be nil")
	}
	var zero T
	return Matcher[T]{
		match: fn,
		desc:  fmt.Sprintf("Pred[%T](<func>)", zero),
	}
}

// Ref returns a Matcher[T] that compares the argument to want using
// assert.Equal's smart-equal algorithm, which supports functional options
// such as assert.IgnoreFields and assert.IgnoreOrder.
//
// opts are passed verbatim to assert.Equal. Ref panics (library-register)
// if want is nil for a non-pointer T.
//
// Example: Ref[[]int]([]int{1, 2, 3}, assert.IgnoreOrder())
func Ref[T any](want T, opts ...assert.EqualOption) Matcher[T] {
	return Matcher[T]{
		match: func(got T) bool {
			// Use a no-op TB to capture the result from assert.Equal without
			// calling t.Errorf; Equal returns bool directly.
			return assert.Equal[T](noopTB{}, got, want, opts...)
		},
		desc: fmt.Sprintf("Ref(%v)", want),
	}
}

// noopTB is a minimal assert.TB that discards all output.
// Used internally by Ref to drive assert.Equal without side effects.
type noopTB struct{}

// Helper implements assert.TB; no-op.
func (noopTB) Helper() {}

// Errorf implements assert.TB; no-op.
func (noopTB) Errorf(string, ...any) {}

// Fatalf implements assert.TB; no-op.
func (noopTB) Fatalf(string, ...any) {}

// FailNow implements assert.TB; no-op.
func (noopTB) FailNow() {}

// Cleanup implements assert.TB; no-op.
func (noopTB) Cleanup(fn func()) {}

// Name implements assert.TB and returns an empty name.
func (noopTB) Name() string { return "" }

// Nil returns a Matcher[any] that passes only when the argument is nil
// (interface nil, pointer nil, or nil slice/map/channel/func).
// Use this for untyped nil checks; for typed nil use Eq[*T](nil).
func Nil() Matcher[any] {
	return Matcher[any]{
		match: func(v any) bool { return isNilAny(v) },
		desc:  "Nil()",
	}
}

// NotNil returns a Matcher[any] that passes only when the argument is
// non-nil.
func NotNil() Matcher[any] {
	return Matcher[any]{
		match: func(v any) bool { return !isNilAny(v) },
		desc:  "NotNil()",
	}
}

// MatchFn returns an untyped matcher backed by fn. Use this as an escape
// hatch when a typed matcher is inconvenient.
//
// fn must be non-nil; MatchFn panics (library-register) otherwise.
func MatchFn(fn func(any) bool) Matcher[any] {
	if fn == nil {
		panic("mock.MatchFn: function must not be nil")
	}
	return Matcher[any]{
		match: fn,
		desc:  "MatchFn(<func>)",
	}
}

// isNilAny reports whether v is nil in any of Go's nilable forms.
func isNilAny(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map,
		reflect.Func, reflect.Chan:
		return rv.IsNil()
	}
	return false
}
