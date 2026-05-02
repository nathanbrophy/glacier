// SPDX-License-Identifier: Apache-2.0

package assert

import (
	"testing"
)

// §21.4 F2, F11; §23.5, §23.13, §23.17

// Bootstrap subset — uses bare-if only. These test Equal before assert can be
// used to test itself.

// TestEqual_Bootstrap_PrimitiveInt verifies Equal(5, 5) returns true.
// §21.4 F2, F11; §23.5
func TestEqual_Bootstrap_PrimitiveInt(t *testing.T) {
	mt := &mockTB{}
	got := Equal(mt, 5, 5)
	if !got {
		t.Fatalf("Equal(5,5) = false, want true")
	}
	if mt.errorfCalls != 0 {
		t.Fatalf("Errorf called %d times, want 0", mt.errorfCalls)
	}
}

// TestEqual_Bootstrap_PrimitiveString verifies Equal("a", "a") == true.
// §21.4 F2; §23.5
func TestEqual_Bootstrap_PrimitiveString(t *testing.T) {
	mt := &mockTB{}
	got := Equal(mt, "a", "a")
	if !got {
		t.Fatalf("Equal(\"a\",\"a\") = false, want true")
	}
	if mt.errorfCalls != 0 {
		t.Fatalf("Errorf called %d times, want 0", mt.errorfCalls)
	}
}

// TestEqual_Bootstrap_NilNil verifies Equal(nil, nil) == true.
// §21.4 E1
func TestEqual_Bootstrap_NilNil(t *testing.T) {
	mt := &mockTB{}
	got := Equal[any](mt, nil, nil)
	if !got {
		t.Fatalf("Equal(nil,nil) = false, want true")
	}
}

// TestEqual_Bootstrap_TypedNilNil verifies typed nil pointers are equal.
// §21.4 E2
func TestEqual_Bootstrap_TypedNilNil(t *testing.T) {
	type S struct{ X int }
	mt := &mockTB{}
	var a, b *S
	got := Equal(mt, a, b)
	if !got {
		t.Fatalf("Equal((*S)(nil), (*S)(nil)) = false, want true")
	}
	if mt.errorfCalls != 0 {
		t.Fatalf("Errorf called %d times, want 0", mt.errorfCalls)
	}
}

// TestEqual_Bootstrap_Mismatch verifies Equal(5, 4) == false and Errorf called.
// §21.4 F2
func TestEqual_Bootstrap_Mismatch(t *testing.T) {
	mt := &mockTB{}
	got := Equal(mt, 5, 4)
	if got {
		t.Fatalf("Equal(5,4) = true, want false")
	}
	if mt.errorfCalls != 1 {
		t.Fatalf("Errorf called %d times, want 1", mt.errorfCalls)
	}
}

// TestEqual_Bootstrap_TypeMismatchAtTop verifies nil pointers of different types are unequal.
// §21.4 E3 — via any
func TestEqual_Bootstrap_TypeMismatchAtTop(t *testing.T) {
	type A struct{ X int }
	type B struct{ Y int }
	mt := &mockTB{}
	got := Equal[any](mt, (*A)(nil), (*B)(nil))
	if got {
		t.Fatalf("Equal[any]((*A)(nil), (*B)(nil)) = true, want false")
	}
}

// TestPrimitiveFastPathTypeNotComparable verifies slices bypass the fast path.
// §23.5
func TestPrimitiveFastPathTypeNotComparable(t *testing.T) {
	// []int is not comparable; primitiveEqual must return false.
	got := primitiveEqual([]int{1, 2}, []int{1, 2})
	if got {
		t.Fatalf("primitiveEqual([]int, []int) = true, want false (slices are not comparable)")
	}
}

// --- Composition tests — use assert.Equal / assert.True / assert.False ---

// TestEqualPointerDeref verifies pointer dereferencing.
// §21.4 F11, E4, E5
func TestEqualPointerDeref(t *testing.T) {
	type S struct{ X int }
	tests := []struct {
		name   string
		got    *S
		want   *S
		wantEq bool
	}{
		{"equal pointers", &S{1}, &S{1}, true},
		{"unequal pointers", &S{1}, &S{2}, false},
		{"nil vs non-nil", nil, &S{1}, false},
		{"both nil", nil, nil, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mt := &mockTB{}
			got := Equal(mt, tc.got, tc.want)
			True(t, got == tc.wantEq, tc.name, "got result", got, "want", tc.wantEq)
		})
	}
}

// TestEqualSliceOrdered verifies order-sensitive slice comparison.
// §21.4 E6
func TestEqualSliceOrdered(t *testing.T) {
	mt := &mockTB{}
	False(t, Equal(mt, []int{1, 2, 3}, []int{3, 2, 1}), "ordered slices [1,2,3] vs [3,2,1]")
}

// TestEqualSliceIgnoreOrder verifies IgnoreOrder makes multiset comparison.
// §21.4 F12, E6
func TestEqualSliceIgnoreOrder(t *testing.T) {
	mt := &mockTB{}
	True(t, Equal(mt, []int{1, 2, 3}, []int{3, 2, 1}, IgnoreOrder()), "IgnoreOrder: [1,2,3] == [3,2,1]")
}

// TestEqualSliceIgnoreOrderMultisetCount verifies multiset count semantics.
// §21.4 E7
func TestEqualSliceIgnoreOrderMultisetCount(t *testing.T) {
	mt := &mockTB{}
	False(t, Equal(mt, []int{1, 2, 3}, []int{1, 2, 3, 3}, IgnoreOrder()), "multiset count mismatch")
}

// TestEqualMapDefault verifies map equality is always order-insensitive.
// §21.4 E8
func TestEqualMapDefault(t *testing.T) {
	mt := &mockTB{}
	True(t, Equal(mt, map[string]int{"a": 1, "b": 2}, map[string]int{"b": 2, "a": 1}), "maps are order-insensitive")
}

// TestEqualMapIgnoreCaseKeys verifies case-insensitive map key comparison.
// §21.4 F13, E9
func TestEqualMapIgnoreCaseKeys(t *testing.T) {
	mt := &mockTB{}
	True(t, Equal(mt, map[string]int{"FOO": 1, "Bar": 2}, map[string]int{"foo": 1, "bar": 2}, IgnoreCase()),
		"IgnoreCase: map key fold")
}

// TestEqualStructWithDelta verifies float delta tolerance in struct fields.
// §21.4 F15, E12
func TestEqualStructWithDelta(t *testing.T) {
	type Point struct{ X, Y float64 }
	mt := &mockTB{}
	True(t, Equal(mt, Point{1.0001, 2.0001}, Point{1.0, 2.0}, WithDelta(0.001)), "WithDelta struct fields")
	mt.reset()
	False(t, Equal(mt, Point{1.01, 2.0}, Point{1.0, 2.0}, WithDelta(0.001)), "WithDelta struct field exceeds delta")
}

// TestEqualStructIgnoreFields verifies field exclusion.
// §21.4 F16
func TestEqualStructIgnoreFields(t *testing.T) {
	type User struct {
		ID   string
		Name string
		Age  int
	}
	mt := &mockTB{}
	True(t, Equal(mt, User{"u1", "Alice", 30}, User{"u1", "Alice", 99}, IgnoreFields("Age")),
		"IgnoreFields: Age excluded")
	mt.reset()
	False(t, Equal(mt, User{"u1", "Alice", 30}, User{"u1", "Bob", 99}, IgnoreFields("Age")),
		"IgnoreFields: Name still compared")
}

// TestEqualStructIgnoreFieldsRecursive verifies recursive field exclusion.
// §21.4 F16
func TestEqualStructIgnoreFieldsRecursive(t *testing.T) {
	type Inner struct {
		Created string
		Val     int
	}
	type Outer struct {
		Inner Inner
	}
	mt := &mockTB{}
	True(t, Equal(mt, Outer{Inner{"now", 1}}, Outer{Inner{"then", 1}}, IgnoreFields("Created")),
		"IgnoreFields recursive: Created skipped at every level")
}

// TestEqualCyclic verifies cycle detection prevents stack overflow.
// §21.4 NF2, E10
func TestEqualCyclic(t *testing.T) {
	type Node struct {
		Val  int
		Next *Node
	}
	a := &Node{Val: 1}
	b := &Node{Val: 1}
	a.Next = a // self-cycle
	b.Next = b
	mt := &mockTB{}
	// Both are self-referential with same Val; should return true.
	True(t, Equal(mt, a, b), "cyclic: same structure both sides → equal")
}

// TestEqualCustomMethod verifies custom Equal(any) bool method is invoked.
// §21.4 F11, E11
func TestEqualCustomMethod(t *testing.T) {
	type Custom struct{ val int }
	// Wrap in a type that has Equal method.
	type Wrapper struct {
		inner       Custom
		alwaysEqual bool
	}
	// Use a different approach: a type with Equal(any) bool.
	type CE struct{ V int }
	// Instead define at package scope via interface check.
	mt := &mockTB{}
	c1 := &customEqualer{val: 42, returnEqual: true}
	c2 := &customEqualer{val: 99, returnEqual: true} // different val but Equal returns true
	True(t, Equal(mt, c1, c2), "custom Equal method returns true → assert equal")
}

// TestEqualCustomMethodReturnsFalse verifies custom Equal returning false.
// §21.4 E11
func TestEqualCustomMethodReturnsFalse(t *testing.T) {
	mt := &mockTB{}
	c1 := &customEqualer{val: 42, returnEqual: false}
	c2 := &customEqualer{val: 42, returnEqual: false}
	False(t, Equal(mt, c1, c2), "custom Equal method returns false → assert not equal")
}

// customEqualer is a test type with a custom Equal(any) bool method.
type customEqualer struct {
	val         int
	returnEqual bool
}

func (c *customEqualer) Equal(other any) bool {
	return c.returnEqual
}

// TestEqualNilVsNonNilDifferentTypes verifies nil/non-nil permutations.
// §21.4 E1, E2, E3
func TestEqualNilVsNonNilDifferentTypes(t *testing.T) {
	type S struct{ X int }
	tests := []struct {
		name   string
		got    any
		want   any
		wantEq bool
	}{
		{"nil nil", nil, nil, true},
		{"typed nil vs nil", (*S)(nil), nil, false},
		{"nil vs typed nil", nil, (*S)(nil), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mt := &mockTB{}
			got := Equal[any](mt, tc.got, tc.want)
			True(t, got == tc.wantEq, tc.name)
		})
	}
}

// TestEqualInterfaceUnwrapping verifies interface unwrapping.
// §21.4 F11
func TestEqualInterfaceUnwrapping(t *testing.T) {
	mt := &mockTB{}
	var a any = 5
	var b any = 5
	True(t, Equal(mt, a, b), "any-wrapped int 5 == 5")
}

// TestEqualInterfaceWrappedDifferentDynamicTypes verifies different dynamic types are not equal.
// §21.4 F11
func TestEqualInterfaceWrappedDifferentDynamicTypes(t *testing.T) {
	mt := &mockTB{}
	var a any = int(5)
	var b any = int64(5)
	False(t, Equal(mt, a, b), "any(int(5)) != any(int64(5)): different types")
}

// TestEqualNaNVsNaN verifies NaN != NaN (Go semantics).
// §21.4 E13
func TestEqualNaNVsNaN(t *testing.T) {
	mt := &mockTB{}
	nan := float64NaN()
	False(t, Equal(mt, nan, nan), "NaN != NaN")
}

// TestEqualWithDeltaNaNHandling verifies WithDelta + NaN is still false.
// §21.4 E13
func TestEqualWithDeltaNaNHandling(t *testing.T) {
	mt := &mockTB{}
	nan := float64NaN()
	False(t, Equal(mt, nan, 1.0, WithDelta(100)), "NaN with delta → false")
}

func float64NaN() float64 {
	// Produce NaN without importing math in test file.
	var x float64
	return x / x // 0/0 = NaN
}

// TestEqualIgnoreWhitespace verifies whitespace normalization.
// §21.4 F14
func TestEqualIgnoreWhitespace(t *testing.T) {
	mt := &mockTB{}
	True(t, Equal(mt, "hello\nworld", "hello   world", IgnoreWhitespace()), "IgnoreWhitespace: normalize")
	mt.reset()
	True(t, Equal(mt, "  foo  bar  ", "foo bar", IgnoreWhitespace()), "IgnoreWhitespace: trim + collapse")
}

// TestEqualChannelByIdentity verifies channel equality by pointer identity.
// §21.4 F11
func TestEqualChannelByIdentity(t *testing.T) {
	ch := make(chan int)
	mt := &mockTB{}
	True(t, Equal(mt, ch, ch), "same channel is equal")
	mt.reset()
	ch2 := make(chan int)
	False(t, Equal(mt, ch, ch2), "different channels are not equal")
}

// TestEqualFuncByIdentity verifies func equality by pointer identity.
// §21.4 F11
func TestEqualFuncByIdentity(t *testing.T) {
	mt := &mockTB{}
	// nil funcs are equal.
	var f1, f2 func()
	True(t, Equal(mt, f1, f2), "nil funcs are equal")
}

// TestEqualLargeRecursive verifies no stack overflow on 100-deep nesting.
// §21.4 NF3
func TestEqualLargeRecursive(t *testing.T) {
	type Node struct {
		Val  int
		Next *Node
	}
	// Build a 100-deep chain.
	build := func(depth, val int) *Node {
		var root, prev *Node
		for i := range depth {
			n := &Node{Val: val + i}
			if prev == nil {
				root = n
			} else {
				prev.Next = n
			}
			prev = n
		}
		return root
	}
	got := build(100, 1)
	want := build(100, 1)
	mt := &mockTB{}
	True(t, Equal(mt, got, want), "100-deep linked list: equal")
}

// TestNotEqualBasic verifies NotEqual basic pass/fail.
// §21.4 F2
func TestNotEqualBasic(t *testing.T) {
	mt := &mockTB{}
	True(t, NotEqual(mt, 1, 2), "NotEqual(1, 2) == true")
	mt.reset()
	False(t, NotEqual(mt, 1, 1), "NotEqual(1, 1) == false")
	True(t, mt.errorfCalls == 1, "Errorf called once on failure")
}

// TestEqualCustomMethodIgnoresStructFields verifies custom Equal short-circuits struct walk.
// §21.4 E11
func TestEqualCustomMethodIgnoresStructFields(t *testing.T) {
	mt := &mockTB{}
	// If custom Equal says "equal", the struct field comparison is skipped.
	c1 := &customEqualer{val: 1, returnEqual: true}
	c2 := &customEqualer{val: 999, returnEqual: true}
	True(t, Equal(mt, c1, c2), "custom Equal true → struct fields not compared")
}

// L-add-5: IgnoreFields with non-existent field name is silently ignored.
func TestEqualIgnoreFieldsNonExistent(t *testing.T) {
	type S struct{ A, B int }
	mt := &mockTB{}
	True(t, Equal(mt, S{1, 2}, S{1, 2}, IgnoreFields("NonExistent")), "non-existent field is silently ignored")
}

// L-add-4: slice of *T where pointers differ but values are equal.
func TestEqualSliceOfPointersDeref(t *testing.T) {
	type S struct{ V int }
	a := []*S{{1}, {2}}
	b := []*S{{1}, {2}}
	mt := &mockTB{}
	True(t, Equal(mt, a, b), "slice of *S with same values: equal")
}

// L-add-2: struct with unexported fields.
func TestEqualStructUnexportedFields(t *testing.T) {
	// unexported fields use reflect.DeepEqual fallback.
	// Two zero-valued structs with unexported fields should be equal.
	type s struct{ x int }
	mt := &mockTB{}
	True(t, Equal(mt, s{}, s{}), "struct with unexported fields (both zero) == equal")
}
