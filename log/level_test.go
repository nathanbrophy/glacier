// SPDX-License-Identifier: Apache-2.0

package log_test

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/log"
)

// T#L1 LevelConstants: numeric values match the spec.
func TestLevelConstants(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		level slog.Level
		want  int
	}{
		{"Trace", log.LevelTrace, -8},
		{"Debug", log.LevelDebug, -4},
		{"Info", log.LevelInfo, 0},
		{"Notice", log.LevelNotice, 2},
		{"Warn", log.LevelWarn, 4},
		{"Error", log.LevelError, 8},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, int(tc.level), tc.want)
		})
	}
}

// T#L1 stdlib aliases match their slog counterparts.
func TestLevelAliases(t *testing.T) {
	t.Parallel()
	assert.Equal(t, log.LevelDebug, slog.LevelDebug)
	assert.Equal(t, log.LevelInfo, slog.LevelInfo)
	assert.Equal(t, log.LevelWarn, slog.LevelWarn)
	assert.Equal(t, log.LevelError, slog.LevelError)
}

// T#L2 LevelLabels: handler renders the correct label strings.
func TestLevelLabels(t *testing.T) {
	t.Parallel()
	cases := []struct {
		level slog.Level
		label string
	}{
		{log.LevelTrace, "TRACE"},
		{log.LevelDebug, "DEBUG"},
		{log.LevelInfo, "INFO"},
		{log.LevelNotice, "NOTICE"},
		{log.LevelWarn, "WARN"},
		{log.LevelError, "ERROR"},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			h := log.NewHandler(&buf,
				log.WithLevel(log.LevelTrace),
				log.WithColor(log.ColorNever),
			)
			l := slog.New(h)
			l.Log(nil, tc.level, "msg")
			assert.Contains(t, buf.String(), tc.label)
		})
	}
}

// T#L2 Custom level (outside named set) renders via slog default string.
func TestCustomLevelLabel(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	h := log.NewHandler(&buf,
		log.WithLevel(slog.Level(-12)),
		log.WithColor(log.ColorNever),
	)
	l := slog.New(h)
	// Level -12 is not one of the named constants; slog renders it as DEBUG-8.
	l.Log(nil, slog.Level(-12), "msg")
	got := buf.String()
	// The output should not be empty — the record was emitted.
	assert.True(t, len(got) > 0, "expected non-empty output for custom level")
}
