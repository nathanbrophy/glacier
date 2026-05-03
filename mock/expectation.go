// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"fmt"
	"reflect"
	"strings"
)

// seqExhaustion controls behavior when a ReturnSeq is exhausted.
type seqExhaustion int

const (
	// SeqCycle wraps the position back to 0 when the sequence is exhausted.
	SeqCycle seqExhaustion = iota
	// SeqExhaust causes the mock to call t.Errorf and return zero values
	// when the sequence is exhausted.
	SeqExhaust
)

// strictMode controls the unmatched-call policy.
type strictMode int

const (
	strictErrorf strictMode = iota // default; t.Errorf, continue
	strictFatalf                   // t.Fatalf, stop test
	lenient                        // record in unmatched, continue silently
)

// returnSeq holds values for ReturnSeq.
//
// invariant: values is non-empty; validated at ReturnSeq call time.
// invariant: pos is only mutated under Mock.mu.
type returnSeq struct {
	values [][]any // invariant: len >= 1
	pos    int     // invariant: 0 <= pos < len(values)
	mode   seqExhaustion
}

// next advances the sequence and returns the next value set.
// tb is used for SeqExhaust failure reporting; methodName is for error messages.
// Must be called under Mock.mu.
func (rs *returnSeq) next(tb interface {
	Errorf(string, ...any)
	Helper()
}, methodName string, retTypes []reflect.Type) []any {
	if rs.pos >= len(rs.values) {
		// Already exhausted.
		switch rs.mode {
		case SeqExhaust:
			tb.Helper()
			tb.Errorf("Mock.%s: ReturnSeq exhausted (SeqExhaust mode); no more return values.", methodName)
			return zeroValues(retTypes)
		default: // SeqCycle
			rs.pos = 0
		}
	}
	vals := rs.values[rs.pos]
	rs.pos++
	return vals
}

// Expectation captures a single registered expectation for one method.
//
// invariant: method is a valid exported identifier matching ^[A-Z][A-Za-z0-9_]*$
//
//	and ≤64 bytes (§23.9 row 20); validated at OnCall time.
//
// invariant: matchers length == number of method parameters or zero (match-any shorthand).
// invariant: count is only mutated while Mock.mu is write-locked (§23.14).
// invariant: retFn, retSeq, and doFn are mutually exclusive; at most one is non-nil.
//
//	Validation happens at Return/ReturnSeq/Do call time (panics).
//
// invariant: minCalls <= maxCalls unless maxCalls == -1 (AnyTimes sentinel).
type Expectation struct {
	method   string       // invariant: validated identifier
	matchers []anyMatcher // invariant: may be nil (match any args)
	count    int          // invariant: guarded by Mock.mu
	minCalls int          // invariant: >= 0
	maxCalls int          // invariant: >= minCalls, or -1 for AnyTimes

	retFn  func([]any) []any // invariant: nil unless Return was called
	retSeq *returnSeq        // invariant: nil unless ReturnSeq was called
	doFn   reflect.Value     // invariant: zero unless Do was called

	// methodType is the reflect.Method for arity and type checking.
	// Set at OnCall time; never mutated after.
	methodType reflect.Method

	// back-ref to the owning mock's TB for seq exhaustion reporting.
	tb interface {
		Errorf(string, ...any)
		Helper()
	}
	// retTypes caches the method's return types for zero-value generation.
	retTypes []reflect.Type
}

// With sets the argument matchers for this expectation. The number of matchers
// must equal the number of parameters of the method, or With may be omitted to
// match any arguments.
//
// The Matcher[T] type parameter must correspond to the parameter type at the
// same position in the method signature; a mismatch causes a panic at
// registration time (library-register format).
//
// With returns the expectation for chaining.
//
//glacier:nolint=panic-in-library test-helper programmer error: matcher-arity mismatch surfaces at expectation setup.
func (e *Expectation) With(matchers ...anyMatcher) *Expectation {
	mt := e.methodType.Type
	// mt is a func type: skip receiver (already stripped for interface methods).
	numIn := mt.NumIn()
	if len(matchers) != numIn {
		panic(fmt.Sprintf(
			"mock.OnCall(%q).With: got %d matcher(s), method has %d parameter(s)",
			e.method, len(matchers), numIn,
		))
	}
	e.matchers = matchers
	return e
}

// Return sets the fixed return values for this expectation.
// The number and types of values must match the method's return signature.
//
// Return, ReturnSeq, and Do are mutually exclusive; calling more than one
// panics (library-register format).
//
// Return returns the expectation for chaining.
//
//glacier:nolint=panic-in-library test-helper programmer error: return arity/type mismatches surface at expectation setup.
func (e *Expectation) Return(vals ...any) *Expectation {
	e.checkReturnExclusive("Return")
	mt := e.methodType.Type
	numOut := mt.NumOut()
	if len(vals) != numOut {
		panic(fmt.Sprintf(
			"mock.OnCall(%q).Return: got %d value(s), method returns %d value(s)",
			e.method, len(vals), numOut,
		))
	}
	// Type-check each return value against the method's return type.
	for i, v := range vals {
		wantType := mt.Out(i)
		if err := checkReturnValue(v, wantType, e.method, i); err != nil {
			panic(err.Error())
		}
	}
	snapshot := make([]any, len(vals))
	copy(snapshot, vals)
	e.retFn = func([]any) []any { return snapshot }
	return e
}

// ReturnSeq sets a sequence of return value sets. On the first call the first
// set is used, on the second the second, and so on.
//
// The default exhaustion mode is SeqCycle: after the last set, the sequence
// wraps back to the first. Pass SeqExhaust to fail the test instead.
//
// vals must be non-empty; each inner slice must match the method's return
// signature in length and type.
//
// Return, ReturnSeq, and Do are mutually exclusive.
// ReturnSeq returns the expectation for chaining.
//
//glacier:nolint=panic-in-library test-helper programmer error: row-arity/type mismatches surface at expectation setup.
func (e *Expectation) ReturnSeq(vals [][]any, mode ...seqExhaustion) *Expectation {
	e.checkReturnExclusive("ReturnSeq")
	if len(vals) == 0 {
		panic(fmt.Sprintf("mock.OnCall(%q).ReturnSeq: vals must be non-empty", e.method))
	}
	mt := e.methodType.Type
	numOut := mt.NumOut()
	for i, row := range vals {
		if len(row) != numOut {
			panic(fmt.Sprintf(
				"mock.OnCall(%q).ReturnSeq: row %d has %d value(s), method returns %d value(s)",
				e.method, i, len(row), numOut,
			))
		}
		for j, v := range row {
			wantType := mt.Out(j)
			if err := checkReturnValue(v, wantType, e.method, j); err != nil {
				panic(fmt.Sprintf("mock.OnCall(%q).ReturnSeq: row %d: %s", e.method, i, err.Error()))
			}
		}
	}
	// Deep-copy vals.
	snapshot := make([][]any, len(vals))
	for i, row := range vals {
		snapshot[i] = make([]any, len(row))
		copy(snapshot[i], row)
	}
	exhaustMode := SeqCycle
	if len(mode) > 0 {
		exhaustMode = mode[len(mode)-1]
	}
	e.retSeq = &returnSeq{values: snapshot, mode: exhaustMode}
	return e
}

// Do sets a function to call when this expectation matches.
// fn must be a function whose parameter types match the method's parameter
// types and whose return types match the method's return types.
// A signature mismatch panics at registration time (library-register format).
//
// Return, ReturnSeq, and Do are mutually exclusive.
// Do returns the expectation for chaining.
//
//glacier:nolint=panic-in-library test-helper programmer error: fn signature mismatches surface at expectation setup.
func (e *Expectation) Do(fn any) *Expectation {
	e.checkReturnExclusive("Do")
	if fn == nil {
		panic(fmt.Sprintf("mock.OnCall(%q).Do: fn must not be nil", e.method))
	}
	fnVal := reflect.ValueOf(fn)
	if fnVal.Kind() != reflect.Func {
		panic(fmt.Sprintf("mock.OnCall(%q).Do: fn must be a function, got %T", e.method, fn))
	}
	// Check signature compatibility.
	mt := e.methodType.Type
	fnType := fnVal.Type()
	numIn := mt.NumIn()
	numOut := mt.NumOut()
	if fnType.NumIn() != numIn {
		panic(fmt.Sprintf(
			"mock.OnCall(%q).Do: fn has %d parameter(s), method has %d",
			e.method, fnType.NumIn(), numIn,
		))
	}
	for i := range numIn {
		want := mt.In(i)
		got := fnType.In(i)
		if !got.AssignableTo(want) && !want.AssignableTo(got) {
			panic(fmt.Sprintf(
				"mock.OnCall(%q).Do: fn parameter %d is %v, want %v",
				e.method, i, got, want,
			))
		}
	}
	if fnType.NumOut() != numOut {
		panic(fmt.Sprintf(
			"mock.OnCall(%q).Do: fn returns %d value(s), method returns %d",
			e.method, fnType.NumOut(), numOut,
		))
	}
	for i := range numOut {
		want := mt.Out(i)
		got := fnType.Out(i)
		if !got.AssignableTo(want) && !want.AssignableTo(got) {
			panic(fmt.Sprintf(
				"mock.OnCall(%q).Do: fn return %d is %v, want %v",
				e.method, i, got, want,
			))
		}
	}
	e.doFn = fnVal
	return e
}

// Times requires the method to be called exactly n times.
// n must be >= 1.
//
// Times returns the expectation for chaining.
//
//glacier:nolint=panic-in-library test-helper programmer error: invalid n is documented as a panic precondition.
func (e *Expectation) Times(n int) *Expectation {
	if n < 1 {
		panic(fmt.Sprintf("mock.OnCall(%q).Times: n must be >= 1, got %d", e.method, n))
	}
	e.minCalls = n
	e.maxCalls = n
	return e
}

// AtLeast requires the method to be called at least n times.
// n must be >= 1.
//
// AtLeast returns the expectation for chaining.
//
//glacier:nolint=panic-in-library test-helper programmer error: invalid n is documented as a panic precondition.
func (e *Expectation) AtLeast(n int) *Expectation {
	if n < 1 {
		panic(fmt.Sprintf("mock.OnCall(%q).AtLeast: n must be >= 1, got %d", e.method, n))
	}
	e.minCalls = n
	e.maxCalls = -1 // AnyTimes upper bound
	return e
}

// AtMost requires the method to be called at most n times.
// n must be >= 0. AtMost(0) is equivalent to Never.
//
// AtMost returns the expectation for chaining.
//
//glacier:nolint=panic-in-library test-helper programmer error: invalid n is documented as a panic precondition.
func (e *Expectation) AtMost(n int) *Expectation {
	if n < 0 {
		panic(fmt.Sprintf("mock.OnCall(%q).AtMost: n must be >= 0, got %d", e.method, n))
	}
	e.minCalls = 0
	e.maxCalls = n
	return e
}

// AnyTimes allows the method to be called zero or more times.
// This expectation never causes a Verify failure regardless of call count.
//
// AnyTimes returns the expectation for chaining.
func (e *Expectation) AnyTimes() *Expectation {
	e.minCalls = 0
	e.maxCalls = -1
	return e
}

// Never asserts the method is never called with matching arguments.
// If the method is called and matches, Verify reports a violation.
//
// Never returns the expectation for chaining.
func (e *Expectation) Never() *Expectation {
	e.minCalls = 0
	e.maxCalls = 0
	return e
}

// checkReturnExclusive panics if another return programming method has already
// been set on this Expectation.
//
//glacier:nolint=panic-in-library test-helper programmer error: Return/ReturnSeq/Do conflict surfaces at expectation setup.
func (e *Expectation) checkReturnExclusive(caller string) {
	if e.retFn != nil {
		panic(fmt.Sprintf("mock.OnCall(%q).%s: Return has already been called; Return/ReturnSeq/Do are mutually exclusive", e.method, caller))
	}
	if e.retSeq != nil {
		panic(fmt.Sprintf("mock.OnCall(%q).%s: ReturnSeq has already been called; Return/ReturnSeq/Do are mutually exclusive", e.method, caller))
	}
	if e.doFn.IsValid() {
		panic(fmt.Sprintf("mock.OnCall(%q).%s: Do has already been called; Return/ReturnSeq/Do are mutually exclusive", e.method, caller))
	}
}

// isExhausted reports whether this expectation has reached its maximum
// allowed call count. Must be called under Mock.mu (read is safe here since
// this is a pure read with count guarded by the write-lock callers hold).
func (e *Expectation) isExhausted() bool {
	if e.maxCalls == -1 {
		return false
	}
	return e.count >= e.maxCalls
}

// matches reports whether this expectation matches the given method name
// and arguments. Does not acquire any lock; caller must hold Mock.mu.
func (e *Expectation) matches(method string, args []any) bool {
	if e.method != method {
		return false
	}
	if e.isExhausted() {
		return false
	}
	if e.matchers == nil {
		// Match-any shorthand.
		return true
	}
	if len(args) != len(e.matchers) {
		return false
	}
	for i, m := range e.matchers {
		if !m.matchAny(args[i]) {
			return false
		}
	}
	return true
}

// String returns a human-readable description of this expectation for
// use in failure messages.
func (e *Expectation) String() string {
	var b strings.Builder
	b.WriteString(e.method)
	b.WriteString("(")
	if e.matchers == nil {
		b.WriteString("<any args>")
	} else {
		for i, m := range e.matchers {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(m.String())
		}
	}
	b.WriteString(")")
	switch e.maxCalls {
	case -1:
		if e.minCalls == 0 {
			b.WriteString(" [AnyTimes]")
		} else {
			fmt.Fprintf(&b, " [AtLeast(%d), called=%d]", e.minCalls, e.count)
		}
	case 0:
		fmt.Fprintf(&b, " [Never, called=%d]", e.count)
	default:
		if e.minCalls == e.maxCalls {
			fmt.Fprintf(&b, " [Times(%d), called=%d]", e.minCalls, e.count)
		} else {
			fmt.Fprintf(&b, " [AtMost(%d), called=%d]", e.maxCalls, e.count)
		}
	}
	return b.String()
}

// checkReturnValue verifies that v can be assigned to the expected return type.
// Returns an error describing the mismatch, or nil if the value is valid.
func checkReturnValue(v any, want reflect.Type, method string, pos int) error {
	if v == nil {
		// nil is valid for any nilable type.
		switch want.Kind() {
		case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map,
			reflect.Func, reflect.Chan:
			return nil
		}
		return fmt.Errorf(
			"mock.OnCall(%q).Return: return value %d is nil but type %v is not nilable",
			method, pos, want,
		)
	}
	got := reflect.TypeOf(v)
	if got.AssignableTo(want) {
		return nil
	}
	// Check if the value is a concrete type satisfying an interface.
	if want.Kind() == reflect.Interface && got.Implements(want) {
		return nil
	}
	return fmt.Errorf(
		"mock.OnCall(%q).Return: return value %d is %v, want %v",
		method, pos, got, want,
	)
}

// zeroValues returns zero reflect.Values for each type in retTypes,
// as a []any slice.
func zeroValues(retTypes []reflect.Type) []any {
	out := make([]any, len(retTypes))
	for i, t := range retTypes {
		out[i] = reflect.Zero(t).Interface()
	}
	return out
}
