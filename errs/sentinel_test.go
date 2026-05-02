// SPDX-License-Identifier: Apache-2.0

package errs_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	"github.com/nathanbrophy/glacier/errs"
)

// mustPanic calls f and returns the recovered panic value (as a string).
// If f does not panic, mustPanic fails the test.
func mustPanic(t *testing.T, name string, f func()) (msg string) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			msg = r.(string)
		} else {
			require.True(t, false, name+": expected panic but did not panic")
		}
	}()
	f()
	return
}

// mustNotPanic calls f and fails the test if f panics.
func mustNotPanic(t *testing.T, name string, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			require.True(t, false, fmt.Sprintf("%s: unexpected panic: %v", name, r))
		}
	}()
	f()
}

// TestSentinelValidCases covers inputs that must not panic.
func TestSentinelValidCases(t *testing.T) {
	t.Parallel()
	type tc struct {
		name    string
		text    string
		wantMsg string // if non-empty, verify .Error() == wantMsg
	}
	cases := []tc{
		{
			name:    "standard package-colon-action format",
			text:    "pkg: cause",
			wantMsg: "pkg: cause",
		},
		{
			name: "trailing empty action — just colon is valid",
			text: "pkg:",
		},
		{
			name: "non-ASCII uppercase allowed (D-ERR-4)",
			text: "pkg: Ünicode",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			var s error
			mustNotPanic(t, c.name, func() { s = errs.Sentinel(c.text) })
			if c.wantMsg != "" {
				assert.Equal(t, s.Error(), c.wantMsg)
			}
		})
	}
}

// TestSentinelPanicCases covers inputs that must panic.
func TestSentinelPanicCases(t *testing.T) {
	t.Parallel()
	type tc struct {
		name            string
		text            string
		wantRegisterRef bool // panic msg must contain "register" or "Register"
		wantErrPrefix   string
	}
	cases := []tc{
		{
			name:            "ASCII uppercase first letter panics with register reference",
			text:            "Pkg: cause",
			wantRegisterRef: true,
		},
		{
			name: "trailing period panics",
			text: "pkg: cause.",
		},
		{
			name: "no colon panics",
			text: "nocolon",
		},
		{
			name: "empty string panics",
			text: "",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			msg := mustPanic(t, c.name, func() { _ = errs.Sentinel(c.text) })
			assert.True(t, msg != "", "panic message was empty")
			if c.wantRegisterRef {
				found := assert.Contains(t, msg, "register") || assert.Contains(t, msg, "Register")
				assert.True(t, found, "panic message does not contain 'register': "+msg)
			}
		})
	}
}

// TestErrorRegisterConformance_errs: the panic string from errs starts with "errs:".
func TestErrorRegisterConformance_errs(t *testing.T) {
	t.Parallel()
	msg := mustPanic(t, "register conformance", func() {
		_ = errs.Sentinel("BAD")
	})
	const prefix = "errs:"
	assert.True(t, len(msg) >= len(prefix) && msg[:len(prefix)] == prefix,
		"panic message does not start with "+prefix+": "+msg)
}

// TestSentinelConcurrentRegisterValidation: multiple Sentinel calls from
// parallel goroutines must not race.
func TestSentinelConcurrentRegisterValidation(t *testing.T) {
	var wg sync.WaitGroup
	texts := []string{
		"pkg: a",
		"pkg: b",
		"pkg: c",
		"other: d",
		"x:",
	}
	for _, text := range texts {
		text := text
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = errs.Sentinel(text)
		}()
	}
	wg.Wait()
}
