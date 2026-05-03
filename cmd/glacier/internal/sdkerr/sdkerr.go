// SPDX-License-Identifier: Apache-2.0

// Package sdkerr defines SDK-internal error types and sentinel values.
// All error messages conform to Glacier's library register: lowercase,
// contains ':', no trailing period.
package sdkerr

import (
	"fmt"

	"github.com/nathanbrophy/glacier/errs"
)

// ErrUserCancelled is returned when the user explicitly cancels an interactive
// prompt (e.g. Ctrl-C during glacier init).
var ErrUserCancelled = errs.Sentinel("sdk: user cancelled")

// ErrAlreadyInitialized is returned by glacier init when the target directory
// already contains files that would be overwritten by the scaffold.
type ErrAlreadyInitialized struct {
	// Path is the directory that was targeted.
	Path string
	// Conflict lists the files that would be overwritten.
	Conflict []string
}

// Error implements error.
func (e *ErrAlreadyInitialized) Error() string {
	return fmt.Sprintf(
		"sdk: init: %s already contains %d file(s) that would be overwritten",
		e.Path, len(e.Conflict),
	)
}

// ErrNoModule is returned by glacier new when no go.mod is found in the
// working directory or any ancestor.
type ErrNoModule struct {
	// Cwd is the directory in which the search started.
	Cwd string
}

// Error implements error.
func (e *ErrNoModule) Error() string {
	return fmt.Sprintf("sdk: new: no go.mod found in %s or any ancestor", e.Cwd)
}

// ErrInvalidName is returned when a user-supplied identifier (command name,
// project name, etc.) fails validation.
type ErrInvalidName struct {
	// Kind describes what was being named (e.g. "command", "project").
	Kind string
	// Name is the value that was rejected.
	Name string
	// Why explains the constraint that was violated.
	Why string
}

// Error implements error.
func (e *ErrInvalidName) Error() string {
	return fmt.Sprintf("sdk: %s: %q is invalid: %s", e.Kind, e.Name, e.Why)
}

// ErrCodegenDrift is returned by glacier generate --check when one or more
// generated files are stale relative to their inputs.
type ErrCodegenDrift struct {
	// StaleFiles lists the paths of out-of-date generated files.
	StaleFiles []string
}

// Error implements error.
func (e *ErrCodegenDrift) Error() string {
	return fmt.Sprintf("sdk: generate: %d generated file(s) are stale", len(e.StaleFiles))
}

// ErrTestsFailed is returned by glacier test when the test run reports
// failures or benchmark regressions.
type ErrTestsFailed struct {
	// Failed is the count of failed test cases.
	Failed int
	// Regressed is the count of benchmark regressions.
	Regressed int
	// FirstFailure is the name of the first failing test, if known.
	FirstFailure string
}

// Error implements error.
func (e *ErrTestsFailed) Error() string {
	return fmt.Sprintf(
		"sdk: test: %d test(s) failed, %d benchmark(s) regressed",
		e.Failed, e.Regressed,
	)
}
