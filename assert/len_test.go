// SPDX-License-Identifier: Apache-2.0

package assert

import "testing"

// §21.4 F3; §23.17

func TestLenSlice(t *testing.T) {
	mt := &mockTB{}
	True(t, Len(mt, []int{1, 2, 3}, 3), "Len: slice of 3")
	Equal(t, mt.errorfCalls, 0)
	mt.reset()
	False(t, Len(mt, []int{1, 2, 3}, 2), "Len: mismatch")
	Equal(t, mt.errorfCalls, 1)
}

func TestLenMap(t *testing.T) {
	mt := &mockTB{}
	True(t, Len(mt, map[string]int{"a": 1, "b": 2}, 2), "Len: map of 2")
}

func TestLenString(t *testing.T) {
	mt := &mockTB{}
	True(t, Len(mt, "hello", 5), "Len: string of 5 bytes")
}

func TestLenChan(t *testing.T) {
	ch := make(chan int, 3)
	ch <- 1
	ch <- 2
	mt := &mockTB{}
	True(t, Len(mt, ch, 2), "Len: buffered chan with 2 elements")
}

// L-add-12: nil channel len returns 0.
func TestLenNilChan(t *testing.T) {
	mt := &mockTB{}
	var ch chan int
	True(t, Len(mt, ch, 0), "Len: nil channel has len 0")
}

func TestLenNonContainer(t *testing.T) {
	mt := &mockTB{}
	False(t, Len(mt, 42, 3), "Len: int is unsupported")
	Equal(t, mt.errorfCalls, 1)
}

func TestLenArray(t *testing.T) {
	mt := &mockTB{}
	arr := [4]int{1, 2, 3, 4}
	True(t, Len(mt, arr, 4), "Len: array of 4")
}
