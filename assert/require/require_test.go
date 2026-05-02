// SPDX-License-Identifier: Apache-2.0

package require_test

import (
	"errors"
	"testing"
	"time"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
)

// mockTB is a recording TB for require tests.
// It records FailNow calls via a channel to detect whether subsequent
// code ran (simulating t.FailNow halting the goroutine).
type mockTB struct {
	errorfCalls  int
	failNowCalls int
	helperCalls  int
	lastMessage  string
}

func (m *mockTB) Helper() { m.helperCalls++ }
func (m *mockTB) Errorf(format string, args ...any) {
	m.errorfCalls++
	m.lastMessage = format
}
func (m *mockTB) Fatalf(format string, args ...any) {}
func (m *mockTB) FailNow()                          { m.failNowCalls++ }
func (m *mockTB) Cleanup(fn func())                 {}
func (m *mockTB) Name() string                      { return "mockTB" }

// §21.4 F22, E24, T20

// TestRequireEqualHaltsOnFailure verifies require.Equal calls Errorf then FailNow on failure.
// R1
func TestRequireEqualHaltsOnFailure(t *testing.T) {
	mt := &mockTB{}
	require.Equal(mt, 1, 2)
	assert.True(t, mt.errorfCalls == 1, "Errorf called once")
	assert.True(t, mt.failNowCalls == 1, "FailNow called once")
}

// TestRequireEqualPassNoHalt verifies require.Equal on success does not call FailNow.
// R2
func TestRequireEqualPassNoHalt(t *testing.T) {
	mt := &mockTB{}
	ok := require.Equal(mt, 1, 1)
	assert.True(t, ok, "Equal(1,1) returns true")
	assert.True(t, mt.failNowCalls == 0, "FailNow not called on pass")
}

// TestRequireEqualGenericMirror verifies generic-for-generic mirror.
// R4; §23.17
func TestRequireEqualGenericMirror(t *testing.T) {
	mt := &mockTB{}
	// int
	ok := require.Equal[int](mt, 5, 5)
	assert.True(t, ok, "require.Equal[int] pass")
	assert.True(t, mt.failNowCalls == 0, "FailNow not called")
	// string
	ok = require.Equal[string](mt, "hello", "hello")
	assert.True(t, ok, "require.Equal[string] pass")
}

// TestRequireGreater verifies require.Greater mirrors assert.Greater.
// R5
func TestRequireGreater(t *testing.T) {
	mt := &mockTB{}
	ok := require.Greater(mt, 5, 4)
	assert.True(t, ok)
	assert.Equal(t, mt.failNowCalls, 0)

	mt2 := &mockTB{}
	require.Greater(mt2, 4, 5)
	assert.Equal(t, mt2.failNowCalls, 1)
}

// TestRequireMatch verifies require.Match mirrors assert.Match.
// R6
func TestRequireMatch(t *testing.T) {
	mt := &mockTB{}
	ok := require.Match(mt, "hello world", "hello *")
	assert.True(t, ok)
	assert.Equal(t, mt.failNowCalls, 0)
}

// TestRequireJSONEq verifies require.JSONEq mirrors assert.JSONEq.
// R7
func TestRequireJSONEq(t *testing.T) {
	mt := &mockTB{}
	got := []byte(`{"a":1}`)
	want := []byte(`{"a":1}`)
	ok := require.JSONEq(mt, got, want)
	assert.True(t, ok)
	assert.Equal(t, mt.failNowCalls, 0)

	mt2 := &mockTB{}
	require.JSONEq(mt2, []byte(`{"a":1}`), []byte(`{"a":2}`))
	assert.Equal(t, mt2.failNowCalls, 1)
}

// TestRequireForEveryAssertMirror verifies each require function halts on fail.
// R3
func TestRequireForEveryAssertMirror(t *testing.T) {
	type testCase struct {
		name string
		fn   func(tb assert.TB)
	}
	sentinel := errors.New("sentinel")
	cases := []testCase{
		{"Equal", func(tb assert.TB) { require.Equal(tb, 1, 2) }},
		{"NotEqual", func(tb assert.TB) { require.NotEqual(tb, 1, 1) }},
		{"True", func(tb assert.TB) { require.True(tb, false) }},
		{"False", func(tb assert.TB) { require.False(tb, true) }},
		{"Nil", func(tb assert.TB) { require.Nil(tb, 42) }},
		{"NotNil", func(tb assert.TB) { require.NotNil(tb, nil) }},
		{"NoError", func(tb assert.TB) { require.NoError(tb, sentinel) }},
		{"Error", func(tb assert.TB) { require.Error(tb, nil) }},
		{"ErrorIs", func(tb assert.TB) { require.ErrorIs(tb, errors.New("x"), sentinel) }},
		{"Contains", func(tb assert.TB) { require.Contains(tb, "abc", "z") }},
		{"Len", func(tb assert.TB) { require.Len(tb, []int{1}, 5) }},
		{"Eventually", func(tb assert.TB) {
			require.Eventually(tb, func() bool { return false }, 20*time.Millisecond, 5*time.Millisecond)
		}},
		{"Match", func(tb assert.TB) { require.Match(tb, "abc", "xyz") }},
		{"Greater", func(tb assert.TB) { require.Greater(tb, 1, 5) }},
		{"Less", func(tb assert.TB) { require.Less(tb, 5, 1) }},
		{"GreaterOrEqual", func(tb assert.TB) { require.GreaterOrEqual(tb, 1, 5) }},
		{"LessOrEqual", func(tb assert.TB) { require.LessOrEqual(tb, 5, 1) }},
		{"InDelta", func(tb assert.TB) { require.InDelta(tb, 1.0, 2.0, 0.1) }},
		{"JSONEq", func(tb assert.TB) { require.JSONEq(tb, []byte(`1`), []byte(`2`)) }},
		{"BytesEq", func(tb assert.TB) { require.BytesEq(tb, []byte("a"), []byte("b")) }},
		{"Subset", func(tb assert.TB) { require.Subset(tb, []int{1}, []int{99}) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mt := &mockTB{}
			tc.fn(mt)
			assert.True(t, mt.errorfCalls >= 1, tc.name+": Errorf called")
			assert.True(t, mt.failNowCalls == 1, tc.name+": FailNow called once on failure")
		})
	}
}
