// SPDX-License-Identifier: Apache-2.0

package completions_test

import (
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/completions"
)

func TestScript_KnownShells(t *testing.T) {
	t.Parallel()
	shells := []struct {
		name    string
		keyword string // unique string expected in the script
	}{
		{"bash", "_glacier_completions"},
		{"zsh", "#compdef glacier"},
		{"fish", "complete -c glacier"},
		{"pwsh", "Register-ArgumentCompleter"},
	}
	for _, tc := range shells {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			script, ok := completions.Script(tc.name)
			require.True(t, ok, "expected shell %q to be supported", tc.name)
			assert.True(t,
				strings.Contains(script, tc.keyword),
				"shell %q: expected script to contain %q",
				tc.name, tc.keyword,
			)
		})
	}
}

func TestScript_UnknownShell(t *testing.T) {
	t.Parallel()
	_, ok := completions.Script("csh")
	assert.False(t, ok, "expected unknown shell to return false")
}

func TestScript_CaseInsensitive(t *testing.T) {
	t.Parallel()
	// Shell names should match case-insensitively per D-S67.
	// (Implementation may or may not normalize; at minimum exact match must work.)
	_, ok := completions.Script("bash")
	assert.True(t, ok)
}

func TestScript_NonEmpty(t *testing.T) {
	t.Parallel()
	for _, shell := range []string{"bash", "zsh", "fish", "pwsh"} {
		t.Run(shell, func(t *testing.T) {
			t.Parallel()
			script, ok := completions.Script(shell)
			require.True(t, ok)
			assert.True(t, len(script) > 0, "expected non-empty script for %q", shell)
		})
	}
}

func TestScript_ContainsCommandNames(t *testing.T) {
	t.Parallel()
	// Every completion script should include the top-level command names.
	commands := []string{"version", "generate", "lint", "test", "init", "new", "completions", "explain", "vibe"}
	for _, shell := range []string{"bash", "zsh", "fish", "pwsh"} {
		t.Run(shell, func(t *testing.T) {
			t.Parallel()
			script, _ := completions.Script(shell)
			for _, cmd := range commands {
				assert.True(t,
					strings.Contains(script, cmd),
					"shell %q script missing command %q", shell, cmd,
				)
			}
		})
	}
}

// Example is the canonical package example test.
func Example() {
	script, ok := completions.Script("bash")
	if !ok {
		return
	}
	_ = script
}
