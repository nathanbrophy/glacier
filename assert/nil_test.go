// SPDX-License-Identifier: Apache-2.0

package assert

import "testing"

// §21.4 F3

func TestNilUntyped(t *testing.T) {
	mt := &mockTB{}
	True(t, Nil(mt, nil), "Nil(nil) = true")
	Equal(t, mt.errorfCalls, 0)
}

func TestNilTypedNil(t *testing.T) {
	type S struct{ X int }
	mt := &mockTB{}
	var p *S
	True(t, Nil(mt, p), "Nil((*S)(nil)) = true (typed-nil-aware)")
	Equal(t, mt.errorfCalls, 0)
}

func TestNilNonNil(t *testing.T) {
	type S struct{ X int }
	mt := &mockTB{}
	False(t, Nil(mt, &S{}), "Nil(&S{}) = false")
	Equal(t, mt.errorfCalls, 1)
}

func TestNotNil(t *testing.T) {
	type S struct{ X int }
	mt := &mockTB{}
	True(t, NotNil(mt, &S{}), "NotNil(&S{}) = true")
	Equal(t, mt.errorfCalls, 0)
	mt.reset()
	False(t, NotNil(mt, nil), "NotNil(nil) = false")
	Equal(t, mt.errorfCalls, 1)
}

func TestNilSliceNil(t *testing.T) {
	mt := &mockTB{}
	var s []int
	True(t, Nil(mt, s), "Nil(nil slice) = true (typed-nil-aware)")
}

func TestNilMapNil(t *testing.T) {
	mt := &mockTB{}
	var m map[string]int
	True(t, Nil(mt, m), "Nil(nil map) = true")
}

// L-add-12: nil channel len returns 0; no panic.
func TestNilChan(t *testing.T) {
	mt := &mockTB{}
	var ch chan int
	True(t, Nil(mt, ch), "Nil(nil chan) = true")
}
