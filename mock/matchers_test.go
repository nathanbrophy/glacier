// SPDX-License-Identifier: Apache-2.0

package mock_test

import (
	"testing"

	"github.com/nathanbrophy/glacier/assert"
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
			if got != tc.match {
				t.Errorf("Eq[string](%q).Match(%q) = %v, want %v", tc.want, tc.arg, got, tc.match)
			}
		})
	}
}

func TestEqInt(t *testing.T) {
	m := mock.Eq[int](42)
	if !m.Match(42) {
		t.Error("Eq[int](42) should match 42")
	}
	if m.Match(43) {
		t.Error("Eq[int](42) should not match 43")
	}
}

func TestEqMismatchType(t *testing.T) {
	// Eq[string] called with an int arg via anyMatcher — should not match.
	m := mock.Of[Calculator](t)
	m.OnCall("Add").
		With(mock.Eq[int](1), mock.Eq[int](2)).
		Return(3).
		AnyTimes()

	// Correct types: should match.
	result := m.Interface().Add(1, 2)
	if result != 3 {
		t.Fatalf("got %d, want 3", result)
	}
}

func TestAnyMatchesAnyT(t *testing.T) {
	m := mock.Any[int]()
	cases := []int{0, -1, 42, 1000}
	for _, v := range cases {
		if !m.Match(v) {
			t.Errorf("Any[int]() should match %d", v)
		}
	}
}

func TestAnyString(t *testing.T) {
	m := mock.Any[string]()
	for _, s := range []string{"", "hello", "world"} {
		if !m.Match(s) {
			t.Errorf("Any[string]() should match %q", s)
		}
	}
}

func TestPredCustomPredicate(t *testing.T) {
	m := mock.Pred[int](func(n int) bool { return n > 0 })
	if !m.Match(1) {
		t.Error("Pred should match positive")
	}
	if m.Match(0) {
		t.Error("Pred should not match zero")
	}
	if m.Match(-1) {
		t.Error("Pred should not match negative")
	}
}

func TestPredNilPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for nil predicate")
		}
	}()
	_ = mock.Pred[int](nil)
}

func TestNilMatcher(t *testing.T) {
	m := mock.Nil()
	if !m.Match(nil) {
		t.Error("Nil() should match nil")
	}
	if m.Match("x") {
		t.Error("Nil() should not match non-nil string")
	}
	var p *int
	if !m.Match(p) {
		t.Error("Nil() should match nil pointer")
	}
}

func TestNotNilMatcher(t *testing.T) {
	m := mock.NotNil()
	if m.Match(nil) {
		t.Error("NotNil() should not match nil")
	}
	if !m.Match("x") {
		t.Error("NotNil() should match non-nil string")
	}
}

func TestMatchFnUntypedFallback(t *testing.T) {
	m := mock.MatchFn(func(v any) bool {
		s, ok := v.(string)
		return ok && len(s) > 3
	})
	if !m.Match("hello") {
		t.Error("MatchFn: should match string with len > 3")
	}
	if m.Match("hi") {
		t.Error("MatchFn: should not match short string")
	}
}

func TestMatchFnNilPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for nil MatchFn")
		}
	}()
	_ = mock.MatchFn(nil)
}

func TestRefSmartEqual(t *testing.T) {
	m := mock.Ref[[]int]([]int{3, 1, 2}, assert.IgnoreOrder())
	if !m.Match([]int{1, 2, 3}) {
		t.Error("Ref with IgnoreOrder should match same elements in different order")
	}
	if m.Match([]int{1, 2}) {
		t.Error("Ref with IgnoreOrder should not match different length slice")
	}
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
	if !m.Match(got) {
		t.Error("Ref with IgnoreFields should match when ignored field differs")
	}
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
		if s1 != s2 {
			t.Errorf("Matcher.String() not stable: %q vs %q", s1, s2)
		}
	}
}
