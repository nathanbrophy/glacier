// SPDX-License-Identifier: Apache-2.0

package assert

import "testing"

// §21.4 F3

func TestContainsString(t *testing.T) {
	mt := &mockTB{}
	True(t, Contains(mt, "hello world", "world"), "Contains: string basic")
	Equal(t, mt.errorfCalls, 0)
	mt.reset()
	False(t, Contains(mt, "hello world", "xyz"), "Contains: string miss")
	Equal(t, mt.errorfCalls, 1)
}

func TestContainsSliceWithSmartEqual(t *testing.T) {
	type S struct{ V int }
	mt := &mockTB{}
	True(t, Contains(mt, []S{{1}, {2}, {3}}, S{2}), "Contains: slice element found")
	mt.reset()
	False(t, Contains(mt, []S{{1}, {2}}, S{99}), "Contains: slice element not found")
	Equal(t, mt.errorfCalls, 1)
}

func TestContainsMapKey(t *testing.T) {
	mt := &mockTB{}
	True(t, Contains(mt, map[string]int{"a": 1, "b": 2}, "a"), "Contains: map key found")
	mt.reset()
	False(t, Contains(mt, map[string]int{"a": 1}, "z"), "Contains: map key not found")
	Equal(t, mt.errorfCalls, 1)
}

func TestContainsWithIgnoreCaseOption(t *testing.T) {
	mt := &mockTB{}
	True(t, Contains(mt, "HELLO WORLD", "world", IgnoreCase()), "Contains: string IgnoreCase")
}

func TestContainsNilHaystack(t *testing.T) {
	mt := &mockTB{}
	False(t, Contains(mt, nil, "x"), "Contains: nil haystack returns false")
	Equal(t, mt.errorfCalls, 1)
}

func TestContainsUnsupportedType(t *testing.T) {
	mt := &mockTB{}
	False(t, Contains(mt, 42, "x"), "Contains: unsupported type returns false")
	Equal(t, mt.errorfCalls, 1)
}
