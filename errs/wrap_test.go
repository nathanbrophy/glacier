// SPDX-License-Identifier: Apache-2.0

package errs_test

import (
	"errors"
	"io"
	"strings"
	"testing"

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
		if got := tw.Error(); got != "" {
			t.Fatalf("nil.Error() = %q, want \"\"", got)
		}
	})

	t.Run("Unwrap() on typed nil returns nil", func(t *testing.T) {
		t.Parallel()
		if got := tw.Unwrap(); got != nil {
			t.Fatalf("nil.Unwrap() = %v, want nil", got)
		}
	})

	t.Run("WithStackTrace() on typed nil returns nil", func(t *testing.T) {
		t.Parallel()
		if got := tw.WithStackTrace(); got != nil {
			t.Fatalf("nil.WithStackTrace() = %v, want nil", got)
		}
	})
}

// --- Wrap(nil) ---

func TestWrapNilReturnsNil(t *testing.T) {
	t.Parallel()
	w := errs.Wrap(nil, "x")
	if w != nil {
		t.Fatalf("expected nil, got %v", w)
	}
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
			if c.wantMsg != "" && got.Error() != c.wantMsg {
				t.Fatalf("Error() = %q, want %q", got.Error(), c.wantMsg)
			}
			if c.wantIsErr && !errors.Is(got, c.inner) {
				t.Fatalf("expected errors.Is(Wrap(e, %q), e) == true, got false", c.prefix)
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
	if !errors.As(wrapped, &target) {
		t.Fatalf("expected errors.As to find *customErr through Wrapper")
	}
	if target != inner {
		t.Fatalf("expected target == inner, got %v", target)
	}
}

// --- Unwrap ---

func TestWrapperUnwrapReturnsInner(t *testing.T) {
	t.Parallel()
	inner := io.EOF
	w := errs.Wrap(inner, "pkg: act")
	if w.Unwrap() != inner {
		t.Fatalf("Unwrap() = %v, want %v", w.Unwrap(), inner)
	}
}

// --- WithStackTrace / StackOf ---

func TestStackTraceCapture(t *testing.T) {
	t.Parallel()

	t.Run("captures frames — first frame names this test", func(t *testing.T) {
		t.Parallel()
		w := errs.Wrap(io.EOF, "x").WithStackTrace()
		frames := errs.StackOf(w)
		if len(frames) == 0 {
			t.Fatal("expected at least 1 frame, got 0")
		}
		first := frames[0]
		if !strings.Contains(first.Function, "TestStackTraceCapture") {
			t.Fatalf("first frame Function = %q, want it to contain TestStackTraceCapture", first.Function)
		}
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
		if len(frames) > 32 {
			t.Fatalf("expected <=32 frames, got %d", len(frames))
		}
		if len(frames) == 0 {
			t.Fatal("expected at least 1 frame")
		}
	})

	t.Run("double WithStackTrace preserves frame count", func(t *testing.T) {
		t.Parallel()
		w := errs.Wrap(io.EOF, "x").WithStackTrace()
		n1 := len(errs.StackOf(w))
		w2 := w.WithStackTrace()
		n2 := len(errs.StackOf(w2))
		if w2 == nil {
			t.Fatal("expected non-nil after second WithStackTrace")
		}
		if n1 != n2 {
			t.Fatalf("second WithStackTrace changed frame count: before=%d after=%d", n1, n2)
		}
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
			if c.wantNil && got != nil {
				t.Fatalf("expected nil, got %v", got)
			}
			if !c.wantNil && got == nil {
				t.Fatal("expected non-nil frames, got nil")
			}
			if !c.wantNil && len(got) == 0 {
				t.Fatal("expected at least 1 frame")
			}
		})
	}
}
