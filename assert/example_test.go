// SPDX-License-Identifier: Apache-2.0

package assert_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/nathanbrophy/glacier/assert"
)

// ExampleEqual_smart demonstrates smart equality with struct fields.
func ExampleEqual_smart() {
	var t *testing.T

	type User struct {
		ID      string
		Name    string
		Created time.Time
	}
	got := User{ID: "u-42", Name: "Ada", Created: time.Now()}
	want := User{ID: "u-42", Name: "Ada"}

	// Without IgnoreFields: false (Created differs).
	// With IgnoreFields: true (Created excluded).
	assert.Equal(t, got, want, assert.IgnoreFields("Created"))
}

// ExampleIgnoreOrder demonstrates multiset slice comparison.
func ExampleIgnoreOrder() {
	var t *testing.T

	got := []int{3, 1, 2}
	want := []int{1, 2, 3}

	assert.Equal(t, got, want, assert.IgnoreOrder()) // true: multiset
}

// ExampleIgnoreFields demonstrates struct field exclusion.
func ExampleIgnoreFields() {
	var t *testing.T

	type Point struct{ X, Y, Z int }
	got := Point{1, 2, 99}
	want := Point{1, 2, 0}

	assert.Equal(t, got, want, assert.IgnoreFields("Z")) // true
}

// ExampleMatchRegex demonstrates regex matching.
func ExampleMatchRegex() {
	var t *testing.T

	assert.Match(t, "user-123", `^user-[0-9]+$`, assert.MatchRegex()) // true
}

// ExampleJSONEq demonstrates JSON equality.
func ExampleJSONEq() {
	var t *testing.T

	got := []byte(`{"name":"Ada","age":36}`)
	want := []byte(`{"age":36,"name":"Ada"}`)

	assert.JSONEq(t, got, want) // true: key order ignored
}

// ExampleMust demonstrates runtime Must at initialization time.
func ExampleMust() {
	var rePhone *regexp.Regexp
	rePhone = assert.Must(regexp.Compile(`^\+?[0-9 -]+$`))
	_ = rePhone
}
