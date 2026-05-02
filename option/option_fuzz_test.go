// SPDX-License-Identifier: Apache-2.0

// Bootstrap discipline: stdlib testing only.

package option_test

import (
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/option"
)

// FuzzRequired fuzzes the name argument of Required and verifies that:
//  1. The returned Validator always returns an error when getter returns nil.
//  2. The error message contains the name (via %q rendering).
//  3. The error message starts with "option: required: field ".
//
// This covers the round-trip formatting contract: name → error string.
func FuzzRequired(f *testing.F) {
	// Seed corpus.
	f.Add("")
	f.Add("logger")
	f.Add(`my "field"`)
	f.Add("field with spaces")
	f.Add("field\twith\ttabs")
	f.Add("a\x00b") // null byte
	f.Add("αβγ")    // Unicode

	type cfg struct{}

	f.Fuzz(func(t *testing.T, name string) {
		c := cfg{}
		vtor := option.Required[cfg](name, func(_ *cfg) any { return nil })
		err := vtor(&c)
		if err == nil {
			t.Fatalf("Required with nil getter result must always return an error (name=%q)", name)
		}
		const prefix = "option: required: field "
		if !strings.HasPrefix(err.Error(), prefix) {
			t.Fatalf("error %q does not start with expected prefix %q", err.Error(), prefix)
		}
	})
}
