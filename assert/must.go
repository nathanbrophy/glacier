// SPDX-License-Identifier: Apache-2.0

package assert

import "fmt"

// Must returns v if err is nil. If err is non-nil, Must panics with a
// value that wraps err. Use only for initialization-time invariants:
// package init, main() setup, test harness setup. Never use in library
// hot paths — library code must not panic (see CLAUDE.md directive 4).
//
// Example:
//
//	var rePhone = assert.Must(regexp.Compile(`^\+?[0-9 -]+$`))
//
// Concurrency: safe if v and err are already goroutine-safe; Must itself
// introduces no shared state.
//
// §21.4 F19, E21, E22
func Must[T any](v T, err error) T {
	if err != nil {
		panic(fmt.Errorf("assert: Must: %w", err))
	}
	return v
}

// Must2 is Must for two-value-plus-error returns. Returns (a, b) if err
// is nil; panics with err otherwise. Same usage rules as Must.
//
// Example:
//
//	n, buf := assert.Must2(io.ReadFull(r, p))
//
// §21.4 F20
func Must2[A, B any](a A, b B, err error) (A, B) {
	if err != nil {
		panic(fmt.Errorf("assert: Must2: %w", err))
	}
	return a, b
}

// Mustf panics with a formatted message if cond is false. The panic value
// is a plain string of the form fmt.Sprintf(format, args...). If cond is
// true, Mustf returns without effect.
//
// Same usage rules as Must: initialization-time invariants only.
//
// Example:
//
//	assert.Mustf(len(os.Args) > 1, "usage: %s <config>", os.Args[0])
//
// §21.4 F21, E22
func Mustf(cond bool, format string, args ...any) {
	if !cond {
		panic(fmt.Sprintf(format, args...))
	}
}
