// SPDX-License-Identifier: Apache-2.0

package errs_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	"github.com/nathanbrophy/glacier/errs"
)

// customErr is a simple typed error for testing errors.As.
type customErr struct{ msg string }

func (e *customErr) Error() string { return e.msg }

// --- Wrapper typed-nil safe-method cases ---

func TestWrapperTypedNilMethods(t *testing.T) {
	t.Parallel()
	var tw *errs.Wrapper

	t.Run("Error() on typed nil returns empty string", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, tw.Error(), "")
	})

	t.Run("Unwrap() on typed nil returns nil", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, tw.Unwrap())
	})

	t.Run("WithStackTrace() on typed nil returns nil", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, tw.WithStackTrace())
	})
}

// --- Wrap(nil) ---

func TestWrapNilReturnsNil(t *testing.T) {
	t.Parallel()
	w := errs.Wrap(nil, "x")
	assert.Nil(t, w)
	// Calling methods on typed nil must be safe.
	var tw *errs.Wrapper
	_ = tw.Error()
	_ = tw.Unwrap()
	_ = tw.WithStackTrace()
}

// --- Wrap format and chain cases ---

func TestWrapFormatAndChain(t *testing.T) {
	t.Parallel()
	type tc struct {
		name      string
		inner     error
		prefix    string
		wantMsg   string // empty means skip format check
		wantIsErr bool   // check errors.Is(result, inner)
	}
	cases := []tc{
		{
			name:      "preserves unwrap chain via errors.Is",
			inner:     io.EOF,
			prefix:    "x",
			wantIsErr: true,
		},
		{
			name:    "formats prefix colon inner message",
			inner:   io.EOF,
			prefix:  "pkg: act",
			wantMsg: "pkg: act: EOF",
		},
		{
			name:    "empty prefix produces leading colon separator",
			inner:   io.EOF,
			prefix:  "",
			wantMsg: ": EOF",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := errs.Wrap(c.inner, c.prefix)
			if c.wantMsg != "" {
				require.Equal(t, got.Error(), c.wantMsg)
			}
			if c.wantIsErr {
				require.True(t, errors.Is(got, c.inner), "expected errors.Is(Wrap(e, "+c.prefix+"), e) == true")
			}
		})
	}
}

// --- errors.As ---

func TestWrapErrorsAs(t *testing.T) {
	t.Parallel()
	inner := &customErr{msg: "inner"}
	wrapped := errs.Wrap(inner, "x")
	var target *customErr
	require.True(t, errors.As(wrapped, &target), "expected errors.As to find *customErr through Wrapper")
	assert.Equal(t, target, inner)
}

// --- Unwrap ---

func TestWrapperUnwrapReturnsInner(t *testing.T) {
	t.Parallel()
	inner := io.EOF
	w := errs.Wrap(inner, "pkg: act")
	require.Equal(t, w.Unwrap(), inner)
}

// --- WithStackTrace / StackOf ---

func TestStackTraceCapture(t *testing.T) {
	t.Parallel()

	t.Run("captures frames — first frame names this test", func(t *testing.T) {
		t.Parallel()
		w := errs.Wrap(io.EOF, "x").WithStackTrace()
		frames := errs.StackOf(w)
		require.True(t, len(frames) > 0, "expected at least 1 frame, got 0")
		first := frames[0]
		assert.True(t, strings.Contains(first.Function, "TestStackTraceCapture"),
			"first frame Function = "+first.Function+" want it to contain TestStackTraceCapture")
	})

	t.Run("frame count capped at 32", func(t *testing.T) {
		t.Parallel()
		var deepWrap func(depth int) *errs.Wrapper
		deepWrap = func(depth int) *errs.Wrapper {
			if depth == 0 {
				return errs.Wrap(io.EOF, "deep").WithStackTrace()
			}
			return deepWrap(depth - 1)
		}
		w := deepWrap(100)
		frames := errs.StackOf(w)
		assert.True(t, len(frames) <= 32, "expected <=32 frames")
		assert.True(t, len(frames) > 0, "expected at least 1 frame")
	})

	t.Run("double WithStackTrace preserves frame count", func(t *testing.T) {
		t.Parallel()
		w := errs.Wrap(io.EOF, "x").WithStackTrace()
		n1 := len(errs.StackOf(w))
		w2 := w.WithStackTrace()
		n2 := len(errs.StackOf(w2))
		assert.NotNil(t, w2, "expected non-nil after second WithStackTrace")
		assert.Equal(t, n1, n2)
	})
}

// --- StackOf edge cases ---

func TestStackOfEdgeCases(t *testing.T) {
	t.Parallel()
	type tc struct {
		name    string
		build   func() error
		wantNil bool
	}
	cases := []tc{
		{
			name:    "StackOf(nil) returns nil",
			build:   func() error { return nil },
			wantNil: true,
		},
		{
			name:    "no WithStackTrace in chain — returns nil",
			build:   func() error { return errs.Wrap(errs.Wrap(io.EOF, "a"), "b") },
			wantNil: true,
		},
		{
			name: "inner WithStackTrace found via outer plain Wrap",
			build: func() error {
				inner := errs.Wrap(io.EOF, "inner").WithStackTrace()
				return errs.Wrap(inner, "outer")
			},
			wantNil: false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := errs.StackOf(c.build())
			if c.wantNil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.True(t, len(got) > 0, "expected at least 1 frame")
			}
		})
	}
}
