// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"errors"
	"fmt"
)

// ErrCancelled is returned when ctx is already done before dispatch.
var ErrCancelled = errors.New("glacier/cli: context cancelled before dispatch")

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

// ErrUnknownCommand is returned by Run/Main when argv names a command path
// that has no registration.
type ErrUnknownCommand struct{ Path string }

// Error implements error.
func (e *ErrUnknownCommand) Error() string {
	return fmt.Sprintf("cli: unknown command: %q", e.Path)
}

// ErrUnknownFlag is returned when argv contains an unrecognized --flag.
type ErrUnknownFlag struct{ Name string }

// Error implements error.
func (e *ErrUnknownFlag) Error() string {
	return fmt.Sprintf("cli: unknown flag: %q", e.Name)
}

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
