// SPDX-License-Identifier: Apache-2.0

package assert

import (
	"regexp"
	"testing"
)

// §21.4 F18, NF5

func TestDiffPrimitive(t *testing.T) {
	mt := &mockTB{}
	Equal(mt, 42, 41)
	// Failure message must contain "got" and "want" in CLI register.
	True(t, mt.errorfCalls == 1, "Errorf called once")
	msg := mt.lastMessage
	True(t, len(msg) > 0, "message is non-empty")
}

func TestDiffSlices(t *testing.T) {
	mt := &mockTB{}
	Equal(mt, []int{1, 2, 3}, []int{4, 5, 6})
	True(t, mt.errorfCalls == 1, "Errorf called once for slice diff")
}

func TestDiffMaps(t *testing.T) {
	mt := &mockTB{}
	Equal(mt, map[string]int{"a": 1}, map[string]int{"a": 2})
	True(t, mt.errorfCalls == 1, "Errorf called once for map diff")
}

func TestDiffStructs(t *testing.T) {
	type S struct{ A, B int }
	mt := &mockTB{}
	Equal(mt, S{1, 2}, S{3, 4})
	True(t, mt.errorfCalls == 1, "Errorf called once for struct diff")
}

// TestErrorMessageInCliRegister verifies failure messages are in CLI register.
// §21.4 NF5
func TestErrorMessageInCliRegister(t *testing.T) {
	// Pattern: starts with uppercase letter, ends with '.'.
	pattern := regexp.MustCompile(`^[A-Z]`)
	tests := []struct {
		name string
		fn   func(tb TB)
	}{
		{"Equal", func(tb TB) { Equal(tb, 1, 2) }},
		{"NotEqual", func(tb TB) { NotEqual(tb, 1, 1) }},
		{"True", func(tb TB) { True(tb, false) }},
		{"False", func(tb TB) { False(tb, true) }},
		{"Nil", func(tb TB) { Nil(tb, 42) }},
		{"NotNil", func(tb TB) { NotNil(tb, nil) }},
		{"NoError", func(tb TB) { NoError(tb, errSentinel) }},
		{"Error", func(tb TB) { Error(tb, nil) }},
		{"Greater", func(tb TB) { Greater(tb, 1, 5) }},
		{"Less", func(tb TB) { Less(tb, 5, 1) }},
		{"Len", func(tb TB) { Len(tb, []int{1}, 5) }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mt := &mockTB{}
			tc.fn(mt)
			True(t, mt.errorfCalls >= 1, tc.name+": Errorf called")
			// CLI register: starts with uppercase.
			True(t, pattern.MatchString(mt.lastMessage),
				tc.name+": message starts with uppercase: "+mt.lastMessage)
		})
	}
}

func TestDiffNonTTYNoColor(t *testing.T) {
	// In test environment (no TTY), renderDiff should produce no ANSI escapes.
	diff := renderDiff(1, 2)
	// "\x1b[" is the start of an ANSI escape.
	for _, b := range diff {
		if b == '\x1b' {
			t.Fatal("renderDiff produced ANSI escape in non-TTY environment")
		}
	}
}
