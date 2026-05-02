// SPDX-License-Identifier: Apache-2.0

package assert

import "testing"

// §21.4 F8

func TestBytesEqIdentical(t *testing.T) {
	mt := &mockTB{}
	True(t, BytesEq(mt, []byte("hello"), []byte("hello")), "BytesEq: identical")
}

func TestBytesEqDifferent(t *testing.T) {
	mt := &mockTB{}
	False(t, BytesEq(mt, []byte("hello"), []byte("world")), "BytesEq: different")
	Equal(t, mt.errorfCalls, 1)
}

func TestBytesEqEmpty(t *testing.T) {
	mt := &mockTB{}
	True(t, BytesEq(mt, []byte{}, []byte{}), "BytesEq: both empty")
}

func TestBytesEqNilVsEmpty(t *testing.T) {
	mt := &mockTB{}
	True(t, BytesEq(mt, nil, []byte{}), "BytesEq: nil == empty (bytes.Equal semantics)")
	mt.reset()
	True(t, BytesEq(mt, []byte{}, nil), "BytesEq: empty == nil")
}
