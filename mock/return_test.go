// SPDX-License-Identifier: Apache-2.0

package mock_test

import (
	"fmt"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	"github.com/nathanbrophy/glacier/mock"
)

func TestExpectationReturnSeq(t *testing.T) {
	m := mock.Of[Queue](t)
	m.OnCall("Pop").ReturnSeq([][]any{
		{1, true},
		{2, true},
		{3, false},
	}).AnyTimes()

	q := m.Interface()
	checkPop := func(wantV int, wantOK bool) {
		t.Helper()
		v, ok := q.Pop()
		assert.Equal(t, wantV, v)
		assert.Equal(t, wantOK, ok)
	}
	checkPop(1, true)
	checkPop(2, true)
	checkPop(3, false)
}

func TestReturnSeqCycleDefault(t *testing.T) {
	m := mock.Of[Queue](t)
	m.OnCall("Pop").ReturnSeq([][]any{
		{1, true},
		{2, true},
	}).AnyTimes() // default: SeqCycle

	q := m.Interface()
	for expected := range []int{1, 2, 1, 2, 1} {
		_ = expected
	}
	v1, _ := q.Pop()
	v2, _ := q.Pop()
	v3, _ := q.Pop() // wraps around
	assert.Equal(t, 1, v1)
	assert.Equal(t, 2, v2)
	assert.True(t, v3 == 1, fmt.Sprintf("SeqCycle: expected wrap-around to 1, got %d", v3))
}

func TestReturnSeqExhaustMode(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Queue](ft, mock.LenientMode())
	m.OnCall("Pop").ReturnSeq([][]any{
		{1, true},
	}, mock.SeqExhaust).AnyTimes()

	q := m.Interface()
	v1, ok1 := q.Pop() // uses index 0
	assert.Equal(t, 1, v1)
	assert.True(t, ok1)
	v2, _ := q.Pop() // exhausted → error
	assert.True(t, v2 == 0, fmt.Sprintf("exhausted Pop: expected zero value, got %d", v2))
	assert.True(t, len(ft.errors) > 0, "SeqExhaust: expected Errorf on exhaustion")
}

func TestReturnSeqEmptyPanics(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Queue](ft)
	defer func() {
		r := recover()
		require.NotNil(t, r, "expected panic for empty ReturnSeq")
	}()
	m.OnCall("Pop").ReturnSeq(nil)
}

func TestReturnMutualExclusivity(t *testing.T) {
	t.Run("Return then ReturnSeq", func(t *testing.T) {
		ft := newFakeT()
		m := mock.Of[Queue](ft)
		defer func() {
			if r := recover(); r == nil {
				require.True(t, false, "expected panic")
			}
		}()
		e := m.OnCall("Pop").Return(1, true)
		e.ReturnSeq([][]any{{2, true}})
	})

	t.Run("ReturnSeq then Return", func(t *testing.T) {
		ft := newFakeT()
		m := mock.Of[Queue](ft)
		defer func() {
			if r := recover(); r == nil {
				require.True(t, false, "expected panic")
			}
		}()
		e := m.OnCall("Pop").ReturnSeq([][]any{{1, true}})
		e.Return(2, true)
	})

	t.Run("Return then Do", func(t *testing.T) {
		ft := newFakeT()
		m := mock.Of[Queue](ft)
		defer func() {
			if r := recover(); r == nil {
				require.True(t, false, "expected panic")
			}
		}()
		e := m.OnCall("Pop").Return(1, true)
		e.Do(func() (int, bool) { return 2, true })
	})
}

func TestWithArityMismatch(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	defer func() {
		r := recover()
		require.NotNil(t, r, "expected panic for With arity mismatch")
	}()
	// Greet takes 1 param; passing 2 matchers should panic.
	m.OnCall("Greet").With(mock.Any[string](), mock.Any[string]())
}
