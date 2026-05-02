// SPDX-License-Identifier: Apache-2.0

package assert

import (
	"math/rand/v2"
	"testing"
)

// §21.4 F2, F4, F7, F9

// PropertyEqualReflexive verifies that for any x, Equal(t, x, x) == true.
func TestPropertyEqualReflexive(t *testing.T) {
	mt := &mockTB{}
	values := []any{
		0, 1, -1, 42,
		"", "hello", "abc",
		true, false,
		[]int{1, 2, 3},
		map[string]int{"a": 1},
	}
	for _, v := range values {
		mt.reset()
		Equal[any](mt, v, v)
		if mt.errorfCalls != 0 {
			t.Errorf("PropertyEqualReflexive: Equal(%v, %v) failed", v, v)
		}
	}
}

// PropertyEqualSymmetric verifies Equal(x, y) == Equal(y, x).
func TestPropertyEqualSymmetric(t *testing.T) {
	pairs := [][2]any{
		{1, 2}, {1, 1}, {"a", "b"}, {"hello", "hello"},
		{nil, nil}, {nil, 42},
	}
	for _, p := range pairs {
		x, y := p[0], p[1]
		mt1 := &mockTB{}
		mt2 := &mockTB{}
		r1 := Equal[any](mt1, x, y)
		r2 := Equal[any](mt2, y, x)
		if r1 != r2 {
			t.Errorf("PropertyEqualSymmetric: Equal(%v,%v)=%v != Equal(%v,%v)=%v",
				x, y, r1, y, x, r2)
		}
	}
}

// PropertyEqualTransitiveOnPrimitives verifies if Equal(a,b) && Equal(b,c) then Equal(a,c).
func TestPropertyEqualTransitiveOnPrimitives(t *testing.T) {
	for range 100 {
		v := rand.IntN(5) // small range so collisions happen
		a, b, c := v, v, v
		mt := &mockTB{}
		if !Equal(mt, a, b) || !Equal(mt, b, c) {
			continue
		}
		mt2 := &mockTB{}
		if !Equal(mt2, a, c) {
			t.Errorf("PropertyEqualTransitive: Equal(%v,%v) && Equal(%v,%v) but !Equal(%v,%v)",
				a, b, b, c, a, c)
		}
	}
}

// PropertyMatchEmptyPatternNeverMatches verifies Match(t, nonEmpty, "") == false.
func TestPropertyMatchEmptyPatternNeverMatches(t *testing.T) {
	nonEmptyStrings := []string{"a", "abc", " ", "hello world", "1"}
	for _, s := range nonEmptyStrings {
		mt := &mockTB{}
		if Match(mt, s, "") {
			t.Errorf("PropertyMatchEmptyPattern: Match(%q, \"\") = true, want false", s)
		}
	}
}

// PropertySubsetReflexive verifies Subset(t, x, x) == true.
func TestPropertySubsetReflexive(t *testing.T) {
	slices := [][]int{
		{},
		{1},
		{1, 2, 3},
		{1, 1, 2},
	}
	for _, s := range slices {
		mt := &mockTB{}
		if !Subset(mt, s, s) {
			t.Errorf("PropertySubsetReflexive: Subset(%v, %v) = false", s, s)
		}
	}
}
