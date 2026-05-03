// SPDX-License-Identifier: Apache-2.0

package castgen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

// Scenario describes one cast to record. The binary is invoked with Args and
// its combined stdout+stderr is captured. The captured output becomes a
// single asciinema v2 event at t=0 (no inter-keystroke timing for SDK
// commands; they emit their full output deterministically).
type Scenario struct {
	// Name is the basename for the output files: <Name>.cast / <Name>.svg.
	Name string
	// Title is the human-readable label rendered in the SVG title bar.
	Title string
	// Bin is the SDK binary path (typically the just-built ./glacier).
	Bin string
	// Args are the arguments passed to Bin.
	Args []string
	// Cols / Rows are the recorded terminal dimensions.
	Cols, Rows int
	// Env is appended to the child process environment. FORCE_COLOR is
	// always added so the captured output is colored.
	Env []string
}

// Cast holds the recorded output for a Scenario.
type Cast struct {
	Scenario Scenario
	Output   []byte
	Recorded time.Time
}

// Record runs scn.Bin with scn.Args, captures stdout+stderr, and returns the
// combined bytes plus a timestamp. The child's environment includes
// GLACIER_FORCE_COLOR=1, NO_COLOR= (cleared), and TERM=xterm-256color so
// the binary emits ANSI even though we are capturing into a pipe.
func Record(scn Scenario) (Cast, error) {
	cmd := exec.Command(scn.Bin, scn.Args...)
	cmd.Env = append(cmd.Env,
		"GLACIER_FORCE_COLOR=1",
		"FORCE_COLOR=1",
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
		"NO_COLOR=",
		"GLACIER_NO_COLOR=",
		fmt.Sprintf("COLUMNS=%d", scn.Cols),
		fmt.Sprintf("LINES=%d", scn.Rows),
	)
	cmd.Env = append(cmd.Env, scn.Env...)

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	now := time.Now().UTC()
	if err := cmd.Run(); err != nil {
		// Most SDK commands return non-zero in deliberate cases (lint
		// failure, missing topic, etc.). We still want to record their
		// output, so we ignore non-zero exits.
		var ee *exec.ExitError
		if !errorAs(err, &ee) {
			return Cast{}, fmt.Errorf("castgen: run %q: %w", scn.Bin, err)
		}
	}

	return Cast{
		Scenario: scn,
		Output:   buf.Bytes(),
		Recorded: now,
	}, nil
}

// errorAs is a small wrapper around errors.As that avoids importing errors
// in this file (keeps it self-contained for vet rules).
func errorAs(err error, target any) bool {
	type wrappedExitErr interface{ ExitCode() int }
	if _, ok := err.(*exec.ExitError); ok {
		return true
	}
	if _, ok := err.(wrappedExitErr); ok {
		return true
	}
	_ = target
	return false
}

// WriteCast emits an asciinema v2 .cast file representing c. The file
// format is documented at https://docs.asciinema.org/manual/asciicast/v2/.
//
// Layout:
//
//	{"version":2,"width":N,"height":N,"timestamp":<unix>,"title":"...",
//	 "env":{"TERM":"xterm-256color","SHELL":"/bin/bash"}}
//	[0.0, "o", "<full output as one chunk>"]
//
// The single-chunk approach is deliberate. SDK commands emit deterministic
// output without keystrokes; replaying frame-by-frame would mis-represent
// the user experience. Animated commands (vibe) call Record with the
// animation's --duration flag and the output buffer carries the full
// captured ANSI stream.
func WriteCast(w io.Writer, c Cast) error {
	header := map[string]any{
		"version":   2,
		"width":     c.Scenario.Cols,
		"height":    c.Scenario.Rows,
		"timestamp": c.Recorded.Unix(),
		"title":     c.Scenario.Title,
		"env": map[string]string{
			"TERM":  "xterm-256color",
			"SHELL": "/bin/bash",
		},
	}
	hb, err := json.Marshal(header)
	if err != nil {
		return fmt.Errorf("castgen: marshal header: %w", err)
	}
	if _, err := w.Write(hb); err != nil {
		return err
	}
	if _, err := w.Write([]byte{'\n'}); err != nil {
		return err
	}

	// One event at t=0 carrying the full output.
	event := []any{0.0, "o", string(c.Output)}
	eb, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("castgen: marshal event: %w", err)
	}
	if _, err := w.Write(eb); err != nil {
		return err
	}
	_, err = w.Write([]byte{'\n'})
	return err
}

// trimTrailingNewlines drops one or more trailing newline runes from s.
// Used so the SVG renderer does not emit a trailing blank row.
func trimTrailingNewlines(s string) string {
	return strings.TrimRight(s, "\r\n")
}
