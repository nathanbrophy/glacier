// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/internal/reflectx"
)

// methodNameRe is the allowed method-name pattern per §23.9 row 20.
var methodNameRe = regexp.MustCompile(`^[A-Z][A-Za-z0-9_]*$`)

// maxMethodNameBytes is the maximum allowed byte length for a method name.
const maxMethodNameBytes = 64

// mockState holds all mutable state for a Mock[T] instance behind a pointer
// so that Mock[T] can be returned by value without copying a sync.RWMutex.
type mockState struct {
	mu           sync.RWMutex
	tb           assert.TB
	expectations []*Expectation
	calls        []*Call
	unmatched    []*Call
	mode         strictMode
	// closed transitions false→true exactly once, under closedMu.
	closed   bool
	closedMu sync.Mutex
	// retTypes maps method name → slice of return reflect.Types.
	retTypes map[string][]reflect.Type
	// methodSet is the set of method names in the interface (for validation).
	methodSet map[string]bool
	// methodTypes maps method name → reflect.Method (from the interface type).
	methodTypes map[string]reflect.Method
}

// Mock is the central handle returned by Of[T]. It holds all registered
// expectations, the call log, and the configuration for one mock instance.
//
// invariant: T must be an interface type; Of[T] panics at runtime otherwise.
// invariant: state.mu is held (write-locked) for the entire duration of every
// method call so that match, increment, and return happen atomically (§23.14).
// invariant: once Verify has run, state.closed is true; subsequent Close calls
// are no-ops.
// invariant: state.expectations is append-only after construction; no reordering.
type Mock[T any] struct {
	state  *mockState
	ifaceT T // the synthesized value satisfying T
}

// Of returns a new Mock[T] for the interface type T and registers a
// t.Cleanup hook that calls Verify when the test ends.
//
// T must be an interface type. Of panics if T is a concrete type or if T is
// the predeclared any type (which is not satisfiably mockable).
//
// Preconditions:
//   - T is an interface type with at least one exported method.
//   - t is non-nil.
//   - A concrete adapter has been registered via RegisterAdapter[T] (or
//     pre-registered by init() for standard library interfaces).
//
// Postconditions:
//   - The returned Mock[T].Interface() value satisfies T.
//   - t.Cleanup is registered; Verify runs automatically when t ends.
//
// Concurrency: Of itself is not called concurrently (it is a setup call).
// The returned Mock[T] is safe for concurrent use once constructed.
//
// Panics if T is not an interface type. The panic message is in library-register
// format: "mock.Of[T]: T must be an interface type, got <kind>".
func Of[T any](t assert.TB, opts ...Option) Mock[T] {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	if ifaceType.Kind() != reflect.Interface {
		panic(fmt.Sprintf(
			"mock.Of[T]: T must be an interface type, got %s",
			ifaceType.Kind(),
		))
	}
	if ifaceType.NumMethod() == 0 {
		panic("mock.Of[T]: T must be an interface type with at least one exported method")
	}

	options := applyMockOptions(opts)

	s := &mockState{
		tb:          t,
		mode:        options.mode,
		retTypes:    make(map[string][]reflect.Type),
		methodSet:   make(map[string]bool),
		methodTypes: make(map[string]reflect.Method),
	}

	// Build method set and return-type map using the reflectx cache.
	info := reflectx.GetOrBuild(ifaceType)
	for _, mi := range info.Methods {
		s.methodSet[mi.Name] = true
		s.methodTypes[mi.Name] = ifaceType.Method(mi.Index)
		s.retTypes[mi.Name] = info.ReturnTypes[mi.Name]
	}

	m := Mock[T]{state: s}

	// Build the synthesized interface value.
	m.ifaceT = m.buildInterfaceValue(ifaceType)

	// Register cleanup to call Verify automatically.
	t.Cleanup(func() {
		s.closedMu.Lock()
		alreadyClosed := s.closed
		s.closedMu.Unlock()
		if !alreadyClosed {
			m.Verify()
		}
	})

	return m
}

// buildInterfaceValue creates a value of type T that routes all method calls
// through m.dispatch. It uses the registered adapter factory for T.
func (m *Mock[T]) buildInterfaceValue(ifaceType reflect.Type) T {
	factory, ok := getAdapterFactory(ifaceType)
	if !ok {
		panic(fmt.Sprintf(
			"mock.Of[%v]: no adapter registered; call mock.RegisterAdapter[%v] in TestMain or init(), "+
				"or annotate the interface with //+glacier:mock and run glaciergen",
			ifaceType, ifaceType,
		))
	}
	dispatch := func(method string, args []reflect.Value) []reflect.Value {
		return m.dispatch(method, args)
	}
	impl := factory(dispatch)
	return impl.(T)
}

// dispatch is called by every synthesized method implementation. It acquires
// the write lock, finds the first matching expectation, increments its counter,
// captures the return values, and returns them :  all under a single critical
// section (§23.14).
func (m *Mock[T]) dispatch(methodName string, args []reflect.Value) []reflect.Value {
	s := m.state

	// Convert reflect.Value args to []any for matcher evaluation.
	rawArgs := make([]any, len(args))
	for i, a := range args {
		if a.IsValid() && a.CanInterface() {
			rawArgs[i] = a.Interface()
		}
	}

	retTypes := s.retTypes[methodName]

	s.mu.Lock()
	defer s.mu.Unlock()

	// Record the call (defensive copy of args).
	argsCopy := make([]any, len(rawArgs))
	copy(argsCopy, rawArgs)
	call := &Call{
		Method: methodName,
		Args:   argsCopy,
		At:     time.Now(),
	}

	// Scan expectations in registration order; first match wins.
	for _, exp := range s.expectations {
		if !exp.matches(methodName, rawArgs) {
			continue
		}
		// Atomic match-AND-increment-AND-respond (§23.14).
		exp.count++
		call.Matched = true
		s.calls = append(s.calls, call)

		var retVals []any
		switch {
		case exp.doFn.IsValid():
			results := exp.doFn.Call(args)
			retVals = make([]any, len(results))
			for i, r := range results {
				if r.IsValid() && r.CanInterface() {
					retVals[i] = r.Interface()
				}
			}
		case exp.retSeq != nil:
			retVals = exp.retSeq.next(s.tb, methodName, retTypes)
		case exp.retFn != nil:
			retVals = exp.retFn(rawArgs)
		default:
			retVals = zeroValues(retTypes)
		}

		return toReflectValues(retVals, retTypes)
	}

	// No matching expectation found.
	call.Matched = false
	s.calls = append(s.calls, call)

	switch s.mode {
	case strictFatalf:
		s.tb.Helper()
		s.tb.Fatalf("%s", m.formatUnmatchedCall(methodName, rawArgs))
	case strictErrorf:
		s.tb.Helper()
		s.tb.Errorf("%s", m.formatUnmatchedCall(methodName, rawArgs))
	case lenient:
		s.unmatched = append(s.unmatched, call)
	}

	return toReflectValues(zeroValues(retTypes), retTypes)
}

// formatUnmatchedCall formats a structured failure message for an unmatched call.
func (m *Mock[T]) formatUnmatchedCall(method string, args []any) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Unexpected call to %s.", method)
	b.WriteString("\nReceived args:")
	for i, a := range args {
		fmt.Fprintf(&b, "\n  [%d] %s", i, formatLogValue(a))
	}
	if len(m.state.expectations) == 0 {
		b.WriteString("\nNo expectations registered.")
	} else {
		b.WriteString("\nRegistered expectations:")
		for _, exp := range m.state.expectations {
			if exp.method == method {
				fmt.Fprintf(&b, "\n  %s", exp.String())
			}
		}
	}
	return b.String()
}

// formatLogValue renders v for use in failure messages, honouring
// slog.LogValuer so that log.Redact(secret) renders as [REDACTED].
func formatLogValue(v any) string {
	if v == nil {
		return "<nil>"
	}
	// slog.Value from slog.LogValuer :  check both the direct interface and
	// a fallback string rendering.
	type logValuerWithAny interface {
		LogValue() interface{ Any() any }
	}
	if lv, ok := v.(logValuerWithAny); ok {
		return fmt.Sprintf("%v", lv.LogValue().Any())
	}
	type simpleLogValuer interface {
		LogValue() any
	}
	if lv, ok := v.(simpleLogValuer); ok {
		return fmt.Sprintf("%v", lv.LogValue())
	}
	return fmt.Sprintf("%v", v)
}

// toReflectValues converts a []any to []reflect.Value using retTypes for
// zero-value substitution when a slot is nil.
func toReflectValues(vals []any, retTypes []reflect.Type) []reflect.Value {
	out := make([]reflect.Value, len(vals))
	for i, v := range vals {
		t := retTypes[i]
		if v == nil {
			out[i] = reflect.Zero(t)
			continue
		}
		rv := reflect.ValueOf(v)
		if t.Kind() == reflect.Interface {
			// Box the concrete value into the interface type.
			iface := reflect.New(t).Elem()
			if rv.Type().Implements(t) || rv.Type().AssignableTo(t) {
				iface.Set(rv)
			} else {
				iface = rv
			}
			out[i] = iface
		} else {
			out[i] = rv
		}
	}
	return out
}

// Interface returns the synthesized value that satisfies T.
// All method calls on the returned value route through the expectation engine.
//
// Concurrency: safe to call from multiple goroutines. Calls on the returned
// value are individually atomic (each call acquires Mock.mu for its duration).
func (m *Mock[T]) Interface() T {
	return m.ifaceT
}

// OnCall starts building an expectation for the method named by method.
//
// Preconditions:
//   - method matches ^[A-Z][A-Za-z0-9_]*$ and is ≤64 bytes (§23.9 row 20).
//   - method names a method in the interface T.
//
// Panics if either precondition is violated; the panic message is in
// library-register format.
//
// Concurrency: OnCall is a setup call; call it before the code under test runs.
// Concurrent OnCall and method dispatch is not supported.
func (m *Mock[T]) OnCall(method string) *Expectation {
	s := m.state
	if len(method) > maxMethodNameBytes {
		panic(fmt.Sprintf(
			"mock.OnCall: method name %q is invalid: exceeds %d bytes",
			method, maxMethodNameBytes,
		))
	}
	if !methodNameRe.MatchString(method) {
		panic(fmt.Sprintf(
			"mock.OnCall: method name %q is invalid: must match ^[A-Z][A-Za-z0-9_]*$",
			method,
		))
	}
	if !s.methodSet[method] {
		panic(fmt.Sprintf(
			"mock.OnCall: method %q not found in interface",
			method,
		))
	}

	mt := s.methodTypes[method]
	exp := &Expectation{
		method:     method,
		minCalls:   1, // default: Times(1)
		maxCalls:   1, // default: Times(1)
		methodType: mt,
		tb:         s.tb,
		retTypes:   s.retTypes[method],
	}
	s.mu.Lock()
	s.expectations = append(s.expectations, exp)
	s.mu.Unlock()
	return exp
}

// CallsTo returns all recorded calls to the named method in arrival order.
// Returns nil if no calls were recorded for that method.
//
// Concurrency: safe to call after the code under test has finished.
func (m *Mock[T]) CallsTo(method string) []*Call {
	s := m.state
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*Call
	for _, c := range s.calls {
		if c.Method == method {
			out = append(out, c)
		}
	}
	return out
}

// UnmatchedCalls returns all calls for which no expectation matched,
// in arrival order. Always empty in strict modes (those calls fail the test
// instead of being recorded here).
//
// Concurrency: safe to call after the code under test has finished.
func (m *Mock[T]) UnmatchedCalls() []*Call {
	s := m.state
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Call, len(s.unmatched))
	copy(out, s.unmatched)
	return out
}

// Verify checks that every registered expectation's call count falls within
// its stated bounds. It reports all violations in a single t.Errorf call.
// Verify is idempotent: subsequent calls after the first are no-ops.
//
// Verify is called automatically at t.Cleanup. Call it manually for
// mid-test checkpoints.
//
// Concurrency: safe to call from any goroutine.
func (m *Mock[T]) Verify() {
	s := m.state
	s.closedMu.Lock()
	if s.closed {
		s.closedMu.Unlock()
		return
	}
	s.closed = true
	s.closedMu.Unlock()

	s.mu.RLock()
	defer s.mu.RUnlock()

	var violations []string
	for _, exp := range s.expectations {
		if exp.count < exp.minCalls {
			violations = append(violations, fmt.Sprintf(
				"  %s: expected at least %d call(s), got %d",
				exp.String(), exp.minCalls, exp.count,
			))
		}
		if exp.maxCalls >= 0 && exp.count > exp.maxCalls {
			violations = append(violations, fmt.Sprintf(
				"  %s: expected at most %d call(s), got %d",
				exp.String(), exp.maxCalls, exp.count,
			))
		}
	}

	if len(violations) > 0 {
		s.tb.Helper()
		s.tb.Errorf("Mock verify failed:\n%s", strings.Join(violations, "\n"))
	}
}

// Close is an alias for Verify that satisfies the io.Closer contract.
// It is idempotent: the second and subsequent calls are no-ops.
// Always returns nil.
//
// Use Close when wiring mock teardown alongside other io.Closer-shaped
// resources in a defer or errs.Join chain.
//
// Concurrency: safe to call from any goroutine.
func (m *Mock[T]) Close() error {
	m.Verify()
	return nil
}
