// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"testing"

	"github.com/nathanbrophy/glacier/assert"
)

// TestExitCodeStability verifies every exit-code constant matches D-S27.
// Adding or changing a constant here is intentional; the test will break
// to force a spec update.
func TestExitCodeStability(t *testing.T) {
	t.Parallel()

	rows := []struct {
		name string
		got  int
		want int
	}{
		{"success", exitSuccess, 0},
		{"generic", exitGeneric, 1},
		{"usage", exitUsage, 2},
		{"generate_failed", exitGenerateFailed, 64},
		{"lint_findings", exitLintFindings, 65},
		{"tests_failed", exitTestsFailed, 66},
		{"scaffold_failed", exitScaffoldFailed, 67},
		{"version_check", exitVersionCheck, 68},
		{"codegen_drift", exitCodegenDrift, 69},
		{"subprocess", exitSubprocess, 70},
		{"interrupted", exitInterrupted, 130},
		{"terminated", exitTerminated, 143},
	}

	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, tc.got)
		})
	}
}
