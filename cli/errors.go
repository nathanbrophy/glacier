// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
)

// ExitCoder is implemented by errors that carry a specific process exit code.
// cli.App.Main inspects the returned error chain (errors.As) and, if any error
// in the chain implements ExitCoder, calls os.Exit with that code instead of 1.
//
// Library and SDK code should implement ExitCoder on structured errors that
// have a stable mapping to a documented exit code (e.g. "tests failed" → 66).
// Generic errors that do not implement ExitCoder map to exit 1 as before.
type ExitCoder interface {
	ExitCode() int
}

// ErrCancelled is returned when ctx is already done before dispatch.
var ErrCancelled = &cancelledError{}

// cancelledError is the type backing ErrCancelled. It implements ExitCoder so
// SIGINT-induced cancellation maps to exit 130 per sysexits convention.
type cancelledError struct{}

// Error implements error.
func (*cancelledError) Error() string { return "glacier/cli: context cancelled before dispatch" }

// ExitCode returns 130 (sysexits SIGINT-on-shell convention).
func (*cancelledError) ExitCode() int { return 130 }

// FlagParseError is returned when argv contains a flag whose value cannot be
// coerced to the declared type (e.g. --port abc for an int flag).
// Replaces the former ParseError name (§23.15 rename).
type FlagParseError struct {
	Name string
	Err  error
}

// Error implements error.
func (e *FlagParseError) Error() string {
	return fmt.Sprintf("cli: flag parse %q: %s", e.Name, e.Err)
}

// Unwrap returns the underlying cause.
func (e *FlagParseError) Unwrap() error { return e.Err }

// ExitCode returns 2 (sysexits EX_USAGE) for any flag-parsing failure.
func (*FlagParseError) ExitCode() int { return 2 }

// ErrUnknownCommand is returned by Run/Main when argv names a command path
// that has no registration.
type ErrUnknownCommand struct{ Path string }

// Error implements error.
func (e *ErrUnknownCommand) Error() string {
	return fmt.Sprintf("cli: unknown command: %q", e.Path)
}

// ExitCode returns 2 (sysexits EX_USAGE).
func (*ErrUnknownCommand) ExitCode() int { return 2 }

// ErrUnknownFlag is returned when argv contains an unrecognized --flag.
type ErrUnknownFlag struct{ Name string }

// Error implements error.
func (e *ErrUnknownFlag) Error() string {
	return fmt.Sprintf("cli: unknown flag: %q", e.Name)
}

// ExitCode returns 2 (sysexits EX_USAGE).
func (*ErrUnknownFlag) ExitCode() int { return 2 }

// ErrMultipleRoots is returned by Register when more than one command is
// marked root.
type ErrMultipleRoots struct{ First, Second string }

// Error implements error.
func (e *ErrMultipleRoots) Error() string {
	return fmt.Sprintf("cli: multiple root commands: %q and %q", e.First, e.Second)
}

// ErrUnresolvedParent is returned by Register when the declared parent path
// does not correspond to any registered command.
type ErrUnresolvedParent struct{ Child, Parent string }

// Error implements error.
func (e *ErrUnresolvedParent) Error() string {
	return fmt.Sprintf("cli: unresolved parent: command %q parent %q not found", e.Child, e.Parent)
}

// ErrPanic wraps a value recovered from a handler panic.
type ErrPanic struct {
	Value any
	Stack []byte // runtime.Stack output
}

// Error implements error.
func (e *ErrPanic) Error() string {
	return fmt.Sprintf("cli: panic: %v", e.Value)
}

// RequiredError is returned when a required flag is missing from all sources.
type RequiredError struct{ Name string }

// Error implements error.
func (e *RequiredError) Error() string {
	return fmt.Sprintf("cli: required flag missing: %q", e.Name)
}

// ExitCode returns 2 (sysexits EX_USAGE).
func (*RequiredError) ExitCode() int { return 2 }

// ChoicesError is returned when a flag value is not in the declared choices set.
type ChoicesError struct {
	Name    string
	Value   string
	Choices []string
}

// Error implements error.
func (e *ChoicesError) Error() string {
	return fmt.Sprintf("cli: flag %q value %q not in choices %v", e.Name, e.Value, e.Choices)
}

// ExitCode returns 2 (sysexits EX_USAGE).
func (*ChoicesError) ExitCode() int { return 2 }
