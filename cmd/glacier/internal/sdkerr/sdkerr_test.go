// SPDX-License-Identifier: Apache-2.0

package sdkerr_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/sdkerr"
)

// errFormat checks the invariant "lowercase, contains ':', no trailing period".
func errFormat(t *testing.T, err error) {
	t.Helper()
	msg := err.Error()
	assert.True(t, strings.Contains(msg, ":"),
		"error %q must contain ':' per library register format", msg)
	assert.True(t, msg == strings.ToLower(msg),
		"error %q must be lowercase per library register format", msg)
	assert.False(t, strings.HasSuffix(msg, "."),
		"error %q must not end with '.' per library register format", msg)
}

func TestErrUserCancelled_Format(t *testing.T) {
	t.Parallel()
	errFormat(t, sdkerr.ErrUserCancelled)
}

func TestErrUserCancelled_IsSentinel(t *testing.T) {
	t.Parallel()
	// Sentinel errors should be comparable with errors.Is.
	err := sdkerr.ErrUserCancelled
	assert.True(t, errors.Is(err, sdkerr.ErrUserCancelled))
}

func TestErrAlreadyInitialized_Format(t *testing.T) {
	t.Parallel()
	err := &sdkerr.ErrAlreadyInitialized{
		Path:     "/tmp/myapp",
		Conflict: []string{"main.go", "go.mod"},
	}
	errFormat(t, err)
}

func TestErrAlreadyInitialized_Message(t *testing.T) {
	t.Parallel()
	err := &sdkerr.ErrAlreadyInitialized{
		Path:     "/tmp/myapp",
		Conflict: []string{"main.go", "go.mod"},
	}
	msg := err.Error()
	assert.True(t, strings.Contains(msg, "/tmp/myapp"), "expected path in error: %q", msg)
	assert.True(t, strings.Contains(msg, "2"), "expected conflict count in error: %q", msg)
}

func TestErrNoModule_Format(t *testing.T) {
	t.Parallel()
	err := &sdkerr.ErrNoModule{Cwd: "/home/user/project"}
	errFormat(t, err)
}

func TestErrNoModule_Message(t *testing.T) {
	t.Parallel()
	err := &sdkerr.ErrNoModule{Cwd: "/home/user/project"}
	msg := err.Error()
	assert.True(t, strings.Contains(msg, "/home/user/project"), "expected cwd in error: %q", msg)
	assert.True(t, strings.Contains(msg, "go.mod"), "expected 'go.mod' in error: %q", msg)
}

func TestErrInvalidName_Format(t *testing.T) {
	t.Parallel()
	// Use a lowercase name so the full message satisfies the lowercase invariant.
	// (The Name field is quoted verbatim; mixed-case user input is valid.)
	err := &sdkerr.ErrInvalidName{Kind: "command", Name: "my-cmd", Why: "name is required"}
	errFormat(t, err)
}

func TestErrInvalidName_Message(t *testing.T) {
	t.Parallel()
	err := &sdkerr.ErrInvalidName{Kind: "package", Name: "", Why: "name is required"}
	msg := err.Error()
	assert.True(t, strings.Contains(msg, "package"), "expected kind in error: %q", msg)
	assert.True(t, strings.Contains(msg, "name is required"), "expected reason in error: %q", msg)
}

func TestErrCodegenDrift_Format(t *testing.T) {
	t.Parallel()
	err := &sdkerr.ErrCodegenDrift{StaleFiles: []string{"zz_generated_cli.go"}}
	errFormat(t, err)
}

func TestErrCodegenDrift_Message(t *testing.T) {
	t.Parallel()
	err := &sdkerr.ErrCodegenDrift{StaleFiles: []string{"a.go", "b.go"}}
	msg := err.Error()
	assert.True(t, strings.Contains(msg, "2"), "expected stale count in error: %q", msg)
}

func TestErrTestsFailed_Format(t *testing.T) {
	t.Parallel()
	err := &sdkerr.ErrTestsFailed{Failed: 3, Regressed: 1}
	errFormat(t, err)
}

func TestErrTestsFailed_Message(t *testing.T) {
	t.Parallel()
	err := &sdkerr.ErrTestsFailed{Failed: 5, Regressed: 2}
	msg := err.Error()
	assert.True(t, strings.Contains(msg, "5"), "expected failed count in error: %q", msg)
	assert.True(t, strings.Contains(msg, "2"), "expected regressed count in error: %q", msg)
}

// Example is the canonical package example test.
func Example() {
	err := &sdkerr.ErrInvalidName{Kind: "command", Name: "", Why: "name is required"}
	_ = err.Error()
}
