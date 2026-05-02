// SPDX-License-Identifier: Apache-2.0

package assert

import "testing"

// §21.4 F5

func TestGreater(t *testing.T) {
	mt := &mockTB{}
	True(t, Greater(mt, 5, 4), "Greater(5, 4) = true")
	Equal(t, mt.errorfCalls, 0)
	mt.reset()
	False(t, Greater(mt, 4, 5), "Greater(4, 5) = false")
	Equal(t, mt.errorfCalls, 1)
	mt.reset()
	False(t, Greater(mt, 5, 5), "Greater(5, 5) = false")
}

func TestLess(t *testing.T) {
	mt := &mockTB{}
	True(t, Less(mt, 3, 10), "Less(3, 10) = true")
	mt.reset()
	False(t, Less(mt, 10, 3), "Less(10, 3) = false")
	Equal(t, mt.errorfCalls, 1)
}

func TestGreaterOrEqual(t *testing.T) {
	mt := &mockTB{}
	True(t, GreaterOrEqual(mt, 5, 5), "GreaterOrEqual(5, 5) = true")
	mt.reset()
	True(t, GreaterOrEqual(mt, 6, 5), "GreaterOrEqual(6, 5) = true")
	mt.reset()
	False(t, GreaterOrEqual(mt, 4, 5), "GreaterOrEqual(4, 5) = false")
	Equal(t, mt.errorfCalls, 1)
}

func TestLessOrEqual(t *testing.T) {
	mt := &mockTB{}
	True(t, LessOrEqual(mt, 5, 5), "LessOrEqual(5, 5) = true")
	mt.reset()
	True(t, LessOrEqual(mt, 3, 5), "LessOrEqual(3, 5) = true")
	mt.reset()
	False(t, LessOrEqual(mt, 6, 5), "LessOrEqual(6, 5) = false")
}

func TestOrderingOnString(t *testing.T) {
	mt := &mockTB{}
	True(t, Greater(mt, "b", "a"), "Greater('b','a') string lexicographic")
	mt.reset()
	True(t, Less(mt, "a", "b"), "Less('a','b') string lexicographic")
}

func TestOrderingOnFloat(t *testing.T) {
	mt := &mockTB{}
	True(t, Greater(mt, 1.1, 1.0), "Greater(1.1, 1.0) float")
}
