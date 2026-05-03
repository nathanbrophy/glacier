// SPDX-License-Identifier: Apache-2.0

package mock_test

import (
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	"github.com/nathanbrophy/glacier/mock"
)

func TestEqTyped(t *testing.T) {
	cases := []struct {
		name  string
		want  string
		arg   string
		match bool
	}{
		{"equal", "u-42", "u-42", true},
		{"not equal", "u-42", "u-43", false},
		{"empty", "", "", true},
		{"empty vs non-empty", "", "x", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := mock.Eq[string](tc.want)
			got := m.Match(tc.arg)
			assert.Equal(t, tc.match, got)
		})
	}
}

func TestEqInt(t *testing.T) {
	m := mock.Eq[int](42)
	assert.True(t, m.Match(42), "Eq[int](42) should match 42")
	assert.False(t, m.Match(43), "Eq[int](42) should not match 43")
}

func TestEqMismatchType(t *testing.T) {
	// Eq[string] called with an int arg via anyMatcher :  should not match.
	m := mock.Of[Calculator](t)
	m.OnCall("Add").
		With(mock.Eq[int](1), mock.Eq[int](2)).
		Return(3).
		AnyTimes()

	// Correct types: should match.
	result := m.Interface().Add(1, 2)
	require.Equal(t, 3, result)
}

func TestAnyMatchesAnyT(t *testing.T) {
	m := mock.Any[int]()
	cases := []int{0, -1, 42, 1000}
	for _, v := range cases {
		assert.True(t, m.Match(v), "Any[int]() should match any int")
	}
}

func TestAnyString(t *testing.T) {
	m := mock.Any[string]()
	for _, s := range []string{"", "hello", "world"} {
		assert.True(t, m.Match(s), "Any[string]() should match "+s)
	}
}

func TestPredCustomPredicate(t *testing.T) {
	m := mock.Pred[int](func(n int) bool { return n > 0 })
	assert.True(t, m.Match(1), "Pred should match positive")
	assert.False(t, m.Match(0), "Pred should not match zero")
	assert.False(t, m.Match(-1), "Pred should not match negative")
}

func TestPredNilPanics(t *testing.T) {
	defer func() {
		r := recover()
		require.NotNil(t, r, "expected panic for nil predicate")
	}()
	_ = mock.Pred[int](nil)
}

func TestNilMatcher(t *testing.T) {
	m := mock.Nil()
	assert.True(t, m.Match(nil), "Nil() should match nil")
	assert.False(t, m.Match("x"), "Nil() should not match non-nil string")
	var p *int
	assert.True(t, m.Match(p), "Nil() should match nil pointer")
}

func TestNotNilMatcher(t *testing.T) {
	m := mock.NotNil()
	assert.False(t, m.Match(nil), "NotNil() should not match nil")
	assert.True(t, m.Match("x"), "NotNil() should match non-nil string")
}

func TestMatchFnUntypedFallback(t *testing.T) {
	m := mock.MatchFn(func(v any) bool {
		s, ok := v.(string)
		return ok && len(s) > 3
	})
	assert.True(t, m.Match("hello"), "MatchFn: should match string with len > 3")
	assert.False(t, m.Match("hi"), "MatchFn: should not match short string")
}

func TestMatchFnNilPanics(t *testing.T) {
	defer func() {
		r := recover()
		require.NotNil(t, r, "expected panic for nil MatchFn")
	}()
	_ = mock.MatchFn(nil)
}

func TestRefSmartEqual(t *testing.T) {
	m := mock.Ref[[]int]([]int{3, 1, 2}, assert.IgnoreOrder())
	assert.True(t, m.Match([]int{1, 2, 3}), "Ref with IgnoreOrder should match same elements in different order")
	assert.False(t, m.Match([]int{1, 2}), "Ref with IgnoreOrder should not match different length slice")
}

func TestRefIgnoreFields(t *testing.T) {
	type Entity struct {
		ID        int
		Name      string
		UpdatedAt string
	}
	want := Entity{ID: 1, Name: "Alice", UpdatedAt: "2026-01-01"}
	m := mock.Ref[Entity](want, assert.IgnoreFields("UpdatedAt"))
	got := Entity{ID: 1, Name: "Alice", UpdatedAt: "different"}
	assert.True(t, m.Match(got), "Ref with IgnoreFields should match when ignored field differs")
}

func TestMatcherStringIsStable(t *testing.T) {
	matchers := []mock.Matcher[string]{
		mock.Eq[string]("hello"),
		mock.Any[string](),
		mock.Pred[string](func(s string) bool { return s != "" }),
	}
	for _, m := range matchers {
		s1 := m.String()
		s2 := m.String()
		assert.True(t, s1 == s2, "Matcher.String() not stable: got "+s1+" then "+s2)
	}
}
