// SPDX-License-Identifier: Apache-2.0

package assert

import "testing"

// §21.4 F9, E19

func TestSubset(t *testing.T) {
	mt := &mockTB{}
	True(t, Subset(mt, []int{1, 2, 3, 4}, []int{2, 3}), "Subset: [2,3] ⊆ [1,2,3,4]")
}

func TestSubsetMissingElement(t *testing.T) {
	mt := &mockTB{}
	False(t, Subset(mt, []int{1, 2}, []int{3}), "Subset: [3] ⊄ [1,2]")
	Equal(t, mt.errorfCalls, 1)
}

func TestSubsetSmartEqual(t *testing.T) {
	type S struct{ V int }
	mt := &mockTB{}
	True(t, Subset(mt, []S{{1}, {2}, {3}}, []S{{1}, {3}}), "Subset: struct slice smart-equal")
}

func TestSubsetEmptyWantAlwaysTrue(t *testing.T) {
	mt := &mockTB{}
	True(t, Subset(mt, []int{1, 2, 3}, []int{}), "Subset: empty want → always true")
	Equal(t, mt.errorfCalls, 0)
}

func TestSubsetEmptyGot(t *testing.T) {
	mt := &mockTB{}
	False(t, Subset(mt, []int{}, []int{1}), "Subset: empty got, non-empty want → false")
}
