// SPDX-License-Identifier: Apache-2.0

package log_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/log"
)

// T#L17 NewJSONHandler emits valid JSON lines.
func TestJSONHandlerValidJSON(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	h := log.NewJSONHandler(&buf, log.WithLevel(log.LevelInfo))
	l := slog.New(h)
	l.Info("hello world", "key", "value")

	line := buf.Bytes()
	// Trim trailing newline if present.
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}
	var m map[string]any
	err := json.Unmarshal(line, &m)
	assert.NoError(t, err)
	assert.Equal(t, m["msg"].(string), "hello world")
	assert.Equal(t, m["key"].(string), "value")
}

// T#L18 Custom level labels (TRACE, NOTICE) appear correctly in JSON.
func TestJSONHandlerCustomLevelLabels(t *testing.T) {
	t.Parallel()
	cases := []struct {
		level slog.Level
		label string
	}{
		{log.LevelTrace, "TRACE"},
		{log.LevelNotice, "NOTICE"},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			h := log.NewJSONHandler(&buf, log.WithLevel(log.LevelTrace))
			l := slog.New(h)
			l.Log(nil, tc.level, "msg")

			line := buf.Bytes()
			if len(line) > 0 && line[len(line)-1] == '\n' {
				line = line[:len(line)-1]
			}
			var m map[string]any
			err := json.Unmarshal(line, &m)
			assert.NoError(t, err)
			assert.Equal(t, m["level"].(string), tc.label)
		})
	}
}

// T#L19 WithSource adds source field to JSON output.
func TestJSONHandlerWithSource(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	h := log.NewJSONHandler(&buf,
		log.WithLevel(log.LevelInfo),
		log.WithSource(),
	)
	l := slog.New(h)
	l.Info("sourced")

	line := buf.Bytes()
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}
	var m map[string]any
	err := json.Unmarshal(line, &m)
	assert.NoError(t, err)
	// The "source" key must be present.
	_, ok := m["source"]
	assert.True(t, ok, "expected source field in JSON output")
}
