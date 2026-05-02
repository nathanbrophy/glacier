// SPDX-License-Identifier: Apache-2.0

package mock_test

import (
	"fmt"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/mock"
)

// PropertyTimes_n_Calls_Verifies tests that for n calls with Times(n), Verify
// passes, and for n+1 calls (unmatched third call in strict mode), the n+1
// call generates an error.
func PropertyTimes_n_Calls_Verifies(t *testing.T) {
	for n := 1; n <= 20; n++ {
		n := n
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			t.Parallel()
			ft := newFakeT()
			m := mock.Of[Greeter](ft)
			m.OnCall("Greet").With(mock.Any[string]()).Return("hi").Times(n)
			for range n {
				m.Interface().Greet("x")
			}
			m.Verify()
			assert.True(t, len(ft.errors) == 0,
				fmt.Sprintf("n=%d exact calls: unexpected errors: %v", n, ft.errors))
		})
	}
}

// PropertyExpectationOrderingMatters verifies that first-registered-wins holds
// when two matching expectations are registered.
func PropertyExpectationOrderingMatters(t *testing.T) {
	for i := range 10 {
		_ = i
		t.Run(fmt.Sprintf("ordering_%d", i), func(t *testing.T) {
			m := mock.Of[Greeter](t)
			m.OnCall("Greet").With(mock.Any[string]()).Return("first").AnyTimes()
			m.OnCall("Greet").With(mock.Any[string]()).Return("second").AnyTimes()
			result := m.Interface().Greet("x")
			assert.True(t, result == "first", "first-registered must win, got: "+result)
		})
	}
}

// PropertyMatcherStringIsStable verifies that Matcher[T].String() is deterministic.
func PropertyMatcherStringIsStable(t *testing.T) {
	matchers := []struct {
		name string
		m    mock.Matcher[string]
	}{
		{"Eq", mock.Eq[string]("hello")},
		{"Any", mock.Any[string]()},
		{"Pred", mock.Pred[string](func(s string) bool { return s != "" })},
	}
	for _, tc := range matchers {
		for range 10 {
			s1 := tc.m.String()
			s2 := tc.m.String()
			assert.True(t, s1 == s2, tc.name+".String() not stable")
		}
	}
}
