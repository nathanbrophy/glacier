// SPDX-License-Identifier: Apache-2.0

package option

import (
	"errors"
	"fmt"
)

// Option configures a value of type T. The unexported apply method
// forbids out-of-module implementations; consumers compose options
// via OptionFunc.
type Option[T any] interface {
	apply(*T) error
}

// OptionFunc is the function adapter that satisfies Option.
//
// Per-package WithX constructors return OptionFunc-wrapped functions:
//
//	func WithLogger(l *slog.Logger) option.Option[config] {
//	    return option.OptionFunc[config](func(c *config) error {
//	        if l == nil {
//	            return errors.New("pkg: WithLogger: logger is nil")
//	        }
//	        c.logger = l
//	        return nil
//	    })
//	}
//
// A nil OptionFunc (one whose underlying func is nil) will panic when Apply
// invokes it. This is a caller error; document it in the With* constructor.
type OptionFunc[T any] func(*T) error

// apply implements Option.
func (f OptionFunc[T]) apply(t *T) error { return f(t) }

// Mode configures Apply's error-handling semantics.
//
// Construct via Strict(). The zero value of Mode is the default
// (short-circuit on first option error).
type Mode struct {
	strict bool
}

// Strict returns a Mode that causes Apply to apply every option even
// when some fail, returning errors.Join over every collected failure.
//
// Use Strict when the caller wants to see every option problem in one
// pass (e.g., displaying configuration errors to a user) rather than
// fixing them one at a time.
//
// Concurrency: Strict() is a pure function; its return value is safe
// to share across goroutines.
func Strict() Mode { return Mode{strict: true} }

// Apply applies opts to a zero-valued T and returns the configured T
// plus any error.
//
// Default behavior: Apply returns at the first option that errors,
// with T reflecting the partial state accumulated up to that point.
// With Strict() mode: Apply applies every option and returns
// errors.Join over all collected failures; T reflects every
// successful option applied.
//
// Nil options in opts are skipped silently. Duplicate options follow
// last-wins semantics by virtue of in-order application. When multiple
// modes are supplied, the last one wins. An empty opts slice returns
// the zero value of T and nil.
//
// Panics propagate: Apply does not recover from options that panic.
//
// Preconditions: none. opts may be nil or empty.
// Postconditions: returned T is always valid (zero-valued or partially/fully
// configured). error is nil when all options succeed.
// Error contract: returns the first option's error (default mode) or
// errors.Join of all failures (Strict mode). Error strings from
// individual options follow each package's own library register.
// Concurrency: goroutine-safe. Apply reads opts and each Option value
// but does not mutate them. Multiple goroutines may call Apply over
// the same []Option[T] concurrently.
// Allocations: zero on the happy path (no errors, no Strict mode).
func Apply[T any](opts []Option[T], mode ...Mode) (T, error) {
	var m Mode
	if n := len(mode); n > 0 {
		m = mode[n-1]
	}
	var t T
	var errs []error
	for _, o := range opts {
		if o == nil {
			continue
		}
		if err := o.apply(&t); err != nil {
			if m.strict {
				errs = append(errs, err)
				continue
			}
			return t, err
		}
	}
	if len(errs) > 0 {
		return t, errors.Join(errs...)
	}
	return t, nil
}

// Validator validates a fully-applied T. Validators run after Apply
// has populated T from options; they check correctness invariants
// that span multiple fields or that depend on a fully-applied state.
//
// Concurrency: Validator[T] is a function type; goroutine-safe by design.
type Validator[T any] func(*T) error

// Validate runs validators against t and returns errors.Join over
// every validator that fails. Nil validators are skipped silently.
//
// Validate is intended to be called by package constructors after
// Apply, like:
//
//	cfg, err := option.Apply(opts)
//	if err != nil {
//	    return nil, err
//	}
//	if err := option.Validate(&cfg, requiredValidators...); err != nil {
//	    return nil, err
//	}
//
// Preconditions: t must not be nil.
// Postconditions: nil returned when all validators pass or validators is empty.
// Error contract: "option: validate: target is nil" when t is nil.
// Otherwise, errors.Join of every failing validator's error.
// Panics propagate: Validate does not recover from validators that panic.
// Concurrency: goroutine-safe when called with independent t values.
func Validate[T any](t *T, validators ...Validator[T]) error {
	if t == nil {
		return errors.New("option: validate: target is nil")
	}
	var errs []error
	for _, v := range validators {
		if v == nil {
			continue
		}
		if err := v(t); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Required returns a Validator that fails if getter returns nil (for
// pointer and interface types) or the zero value of any (for value
// types where nil signals absence). The error message is:
//
//	option: required: field "<name>" not set
//
// Required is the most common validator and ships in the kernel for
// convenience; package authors may write their own Validator functions
// for more complex checks (range checks, mutual exclusion, format
// validation).
//
// The getter receives a *T so it can navigate the config struct.
// Return the field value as any; Required checks whether it is nil.
// For non-pointer fields, return nil explicitly to signal absence:
//
//	option.Required[config]("logger", func(c *config) any { return c.logger })
//
// T is load-bearing in Required: the getter is typed to *T, giving
// compile-time safety that the getter navigates the correct config type.
//
// Preconditions: name may be empty; produces field "" not set (allowed, tested).
// Concurrency: the returned Validator[T] is a closure; goroutine-safe as long
// as the getter does not mutate t.
func Required[T any](name string, getter func(*T) any) Validator[T] {
	return func(t *T) error {
		if getter(t) == nil {
			return fmt.Errorf("option: required: field %q not set", name)
		}
		return nil
	}
}
