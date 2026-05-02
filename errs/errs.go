// SPDX-License-Identifier: Apache-2.0

package errs

import (
	"errors"
	"fmt"
	"iter"
	"runtime"
	"strconv"
	"strings"
)

// Wrapper is the concrete error type returned by Wrap. It supports an
// optional fluent stack-trace capture via WithStackTrace and preserves
// the unwrap chain so errors.Is and errors.As traverse through it.
type Wrapper struct {
	prefix string
	err    error
	stack  []runtime.Frame
}

// Error implements error. Format: "<prefix>: <err.Error()>".
// Returns "" on a nil receiver.
func (w *Wrapper) Error() string {
	if w == nil {
		return ""
	}
	return w.prefix + ": " + w.err.Error()
}

// Unwrap returns the wrapped error, enabling errors.Is and errors.As traversal.
// Returns nil on a nil receiver.
func (w *Wrapper) Unwrap() error {
	if w == nil {
		return nil
	}
	return w.err
}

// WithStackTrace captures the call stack at the call site (up to 32 frames)
// and attaches it to w. Returns w for chaining; nil-receiver-safe.
// Subsequent calls to StackOf will return the captured frames.
//
// Use sparingly: stack capture allocates and adds latency.
func (w *Wrapper) WithStackTrace() *Wrapper {
	if w == nil {
		return nil
	}
	if w.stack != nil {
		// Already captured; do not overwrite.
		return w
	}
	const max = 32
	pcs := make([]uintptr, max)
	n := runtime.Callers(2, pcs) // skip: runtime.Callers + WithStackTrace
	frames := runtime.CallersFrames(pcs[:n])
	w.stack = make([]runtime.Frame, 0, n)
	for {
		f, more := frames.Next()
		w.stack = append(w.stack, f)
		if !more || len(w.stack) >= max {
			break
		}
	}
	return w
}

// Wrap returns a Wrapper that prepends prefix and preserves err's unwrap chain.
// Returns nil if err is nil. Opt in to stack capture via .WithStackTrace().
//
// Prefix should follow Glacier's library register: lowercase, no trailing
// period, "package: action" shape. Example:
//
//	return errs.Wrap(err, "cli: parse")
//	return errs.Wrap(err, "sandbox: spawn").WithStackTrace()
func Wrap(err error, prefix string) *Wrapper {
	if err == nil {
		return nil
	}
	return &Wrapper{prefix: prefix, err: err}
}

// StackOf returns the stack frames captured by WithStackTrace anywhere in
// err's unwrap chain, or nil if none were captured.
func StackOf(err error) []runtime.Frame {
	for err != nil {
		if w, ok := err.(*Wrapper); ok && w.stack != nil {
			return w.stack
		}
		err = errors.Unwrap(err)
	}
	return nil
}

// Join composes multiple errors, dropping nil entries. Returns nil if all
// inputs are nil or the input is empty. Returns the single surviving non-nil
// error directly (preserving identity) when exactly one remains. Otherwise
// delegates to errors.Join.
func Join(errs ...error) error {
	var nonNil []error
	for _, e := range errs {
		if e != nil {
			nonNil = append(nonNil, e)
		}
	}
	switch len(nonNil) {
	case 0:
		return nil
	case 1:
		return nonNil[0]
	default:
		return errors.Join(nonNil...)
	}
}

// Chain returns an iterator that yields every error in err's tree, depth-first,
// walking both Unwrap() error (linear chain) and Unwrap() []error (fan-out).
// The first yielded error is err itself. Yields nothing for a nil err.
//
//	for e := range errs.Chain(err) {
//	    if myErr := (*MyError)(nil); errors.As(e, &myErr) { … }
//	}
func Chain(err error) iter.Seq[error] {
	return func(yield func(error) bool) {
		if err == nil {
			return
		}
		walk(err, yield)
	}
}

// walk is the recursive DFS worker for Chain.
func walk(err error, yield func(error) bool) bool {
	if !yield(err) {
		return false
	}
	type multiUnwrapper interface{ Unwrap() []error }
	if mu, ok := err.(multiUnwrapper); ok {
		for _, child := range mu.Unwrap() {
			if child == nil {
				continue
			}
			if !walk(child, yield) {
				return false
			}
		}
		return true
	}
	if next := errors.Unwrap(err); next != nil {
		return walk(next, yield)
	}
	return true
}

// validRegister checks Glacier's library-register format: non-empty, contains
// at least one ':', no trailing '.', no ASCII uppercase letters (A–Z).
// Returns a non-empty violation description on failure, "" on success.
// Unicode non-ASCII uppercase is allowed (D-ERR-4).
func validRegister(text string) string {
	if text == "" {
		return "text must be non-empty"
	}
	for i := range len(text) {
		c := text[i]
		if c >= 'A' && c <= 'Z' {
			return fmt.Sprintf("text contains ASCII uppercase letter %q", c)
		}
	}
	if !strings.Contains(text, ":") {
		return "text must contain at least one ':' separator"
	}
	if strings.HasSuffix(text, ".") {
		return "text must not end with '.'"
	}
	return ""
}

type sentinelError struct{ text string }

func (s *sentinelError) Error() string { return s.text }

// Sentinel constructs a sentinel error with stable text. The text MUST conform
// to Glacier's library register (lowercase, contains ':', no trailing period).
// Misformatted text panics at construction time — sentinels are always declared
// at package level, so violations surface immediately in any test that imports
// the package.
//
//	var ErrCancelled   = errs.Sentinel("cli: cancelled")
//	var ErrUnknownFlag = errs.Sentinel("cli: unknown flag")
func Sentinel(text string) error {
	if reason := validRegister(text); reason != "" {
		panic("errs: Sentinel: text " + strconv.Quote(text) +
			" does not conform to the Glacier library register: " + reason)
	}
	return &sentinelError{text: text}
}

// IsAny reports whether errors.Is(err, t) is true for any t in targets.
//
//	if errs.IsAny(err, cli.ErrCancelled, context.Canceled, context.DeadlineExceeded) {
//	    return cleanShutdown()
//	}
func IsAny(err error, targets ...error) bool {
	for _, t := range targets {
		if errors.Is(err, t) {
			return true
		}
	}
	return false
}

// retryableError is the marker type used by MarkRetryable.
type retryableError struct{ err error }

func (r *retryableError) Error() string   { return r.err.Error() }
func (r *retryableError) Unwrap() error   { return r.err }
func (r *retryableError) Retryable() bool { return true }

// MarkRetryable wraps err with a marker indicating the operation is safe to
// retry. Returns nil if err is nil. Custom error types may implement
// Retryable() bool directly to participate without this wrapper.
func MarkRetryable(err error) error {
	if err == nil {
		return nil
	}
	return &retryableError{err: err}
}

// Retryable reports whether err (or any error in its chain) is marked
// retryable, i.e. implements Retryable() bool and returns true.
func Retryable(err error) bool {
	type retryabler interface{ Retryable() bool }
	var target retryabler
	if errors.As(err, &target) {
		return target.Retryable()
	}
	return false
}

// Coded is the optional interface for errors that carry a stable machine-readable
// code alongside their human-readable message. Conventionally formatted
// ^[A-Z][A-Z0-9_]*$ (e.g. "E_TIMEOUT") but Code does not validate the format.
type Coded interface {
	error
	Code() string
}

// Code returns the code of the first Coded error in err's chain, or "" if none.
func Code(err error) string {
	var c Coded
	if errors.As(err, &c) {
		return c.Code()
	}
	return ""
}
