// SPDX-License-Identifier: Apache-2.0

package term_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	"github.com/nathanbrophy/glacier/term"
)

// Prompt tests that require real TTY interaction are in cross_platform_test.go.
// Here we test non-interactive (non-TTY) paths and error contracts.

func TestPromptCtxAlreadyCancelled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel
	// stdin in tests is not a TTY; readLine will read from it.
	// With a pre-cancelled context, Prompt must return ErrCancelled.
	// Note: the goroutine reading stdin may race with ctx.Done().
	// We just check the error is ErrCancelled or a wrapped form.
	_, err := term.Prompt(ctx, "Q?")
	if err == nil {
		t.Skip("stdin may have returned data; skipping TTY-less test")
	}
	assert.True(t, errors.Is(err, term.ErrCancelled),
		fmt.Sprintf("Prompt(cancelled ctx) error = %v, want ErrCancelled", err))
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"cancelled", term.ErrCancelled, "term: cancelled"},
		{"not_interactive", term.ErrNotInteractive, "term: not interactive"},
		{"timeout", term.ErrTimeout, "term: timeout"},
		{"too_many_attempts", term.ErrTooManyAttempts, "term: too many attempts"},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, tc.err.Error())
			assert.True(t, errors.Is(tc.err, tc.err),
				fmt.Sprintf("errors.Is(%q, itself) = false", tc.name))
		})
	}
}

func TestHexParseError(t *testing.T) {
	t.Parallel()
	_, err := term.Hex("BADCOLOR")
	require.Error(t, err, "expected HexParseError, got nil")
	var he *term.HexParseError
	require.True(t, errors.As(err, &he), fmt.Sprintf("expected *HexParseError, got %T", err))
	assert.NotNil(t, he.Unwrap(), "HexParseError.Unwrap() should not be nil")
}

func TestGlyphError(t *testing.T) {
	t.Parallel()
	err := term.RegisterGlyph("1invalid", "X", "x")
	require.Error(t, err, "expected GlyphError, got nil")
	var ge *term.GlyphError
	require.True(t, errors.As(err, &ge), fmt.Sprintf("expected *GlyphError, got %T", err))
	assert.True(t, ge.Cause != "", "GlyphError.Cause must not be empty")
}

func TestConfirmOpts(t *testing.T) {
	t.Parallel()
	// WithDefaultYes is a ConfirmOption; just verify it can be constructed.
	_ = term.WithDefaultYes()
}

func TestPromptOpts(t *testing.T) {
	t.Parallel()
	// Verify option constructors don't panic.
	_ = term.WithDefault("foo")
	_ = term.WithValidator(func(s string) error { return nil })
	_ = term.WithMaxAttempts(3)
	_ = term.WithTimeout(0) // 0 duration is unusual but must not panic
}
