// SPDX-License-Identifier: Apache-2.0

// Package report_test tests the report package. Tests that call SetWriter are
// run sequentially because report uses a global writer (safe for real use, but
// t.Parallel() would cause races between test bodies and their assertions).
package report_test

import (
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/report"
)

func TestStatus_WritesToWriter(t *testing.T) {
	var buf strings.Builder
	report.SetWriter(&buf)
	defer report.SetWriter(nil)

	report.Status(report.Calm, "hello world")
	got := buf.String()

	assert.True(t, strings.Contains(got, "hello world"),
		"expected output to contain message, got: %q", got)
}

func TestStatus_AllLevels(t *testing.T) {
	cases := []struct {
		name  string
		level report.Level
	}{
		{"calm", report.Calm},
		{"confident", report.Confident},
		{"thinking", report.Thinking},
		{"alarmed", report.Alarmed},
		{"err", report.Err},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf strings.Builder
			report.SetWriter(&buf)
			defer report.SetWriter(nil)

			report.Status(tc.level, "test message")
			got := buf.String()
			assert.True(t, strings.Contains(got, "test message"),
				"level %s: expected output to contain message, got: %q", tc.name, got)
			assert.True(t, strings.Contains(got, "ʕ"),
				"level %s: expected kaomoji in output, got: %q", tc.name, got)
		})
	}
}

func TestStatus_Format(t *testing.T) {
	var buf strings.Builder
	report.SetWriter(&buf)
	defer report.SetWriter(nil)

	report.Status(report.Confident, "glacier version")
	got := buf.String()

	// Should be "<kaomoji> glacier version\n"
	assert.True(t, strings.HasSuffix(got, "\n"),
		"expected trailing newline, got: %q", got)
	parts := strings.Fields(got)
	assert.True(t, len(parts) >= 2,
		"expected at least kaomoji + message, got: %q", got)
}

func TestSetWriter_Nil_RestoresStderr(t *testing.T) {
	// SetWriter(nil) restores the default (stderr). Must not panic.
	report.SetWriter(nil)
	report.Status(report.Calm, "nil writer test — should go to stderr")
	// No assertion; just verify no panic and no nil-write crash.
}

// Example is the canonical package example test.
func Example() {
	report.Status(report.Calm, "glacier version")
}
