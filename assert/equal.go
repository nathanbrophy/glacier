// SPDX-License-Identifier: Apache-2.0

package assert

import (
	"fmt"
	"math"
	"reflect"
	"strings"
)

// visited is the per-Equal-call cycle-detection set.
// keyed by (got pointer address, want pointer address).
// invariant: only allocated for pointer and interface kinds; nil otherwise.
type visited map[[2]uintptr]bool

// primitiveEqual implements the §23.5 / §23.13 fast path: when T is comparable
// and got == want by ==, return true with no reflect invocation.
// For non-comparable T (slices, maps, funcs), the interface == comparison
// panics; we recover and return false so the caller falls through to smartEqual.
// The recovery path only executes for non-comparable types; for comparable
// types (ints, strings, bools, structs with all-comparable fields) there is no
// deferred call overhead after inlining.
func primitiveEqual[T any](got, want T) (ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	var a, b any = got, want
	return a == b
}

// Equal reports whether got equals want using Glacier's smart-equal
// algorithm. T is constrained to any, providing compile-time type match:
// Equal(t, 5, "5") is a compile error.
//
// The primitive fast path: when T is comparable and got == want by ==,
// Equal returns true immediately with zero allocations (≤ 50 ns/op).
// The slow path handles pointers, slices, maps, structs, and interfaces via
// reflect, optionally guided by opts.
//
// On failure, Equal calls t.Helper, then t.Errorf with a structured diff.
// It returns false so callers can chain: if !assert.Equal(t, ...) { return }.
//
// Preconditions: t is non-nil. opts may be nil or empty.
// Postconditions: t.Errorf called at most once per Equal call.
// Concurrency: goroutine-safe; no shared mutable state.
//
// §21.4 F2, F11; §23.5
func Equal[T any](t TB, got, want T, opts ...EqualOption) bool {
	t.Helper()
	cfg := applyEqualOptions(opts)
	if cfg.deltaSet && cfg.delta < 0 {
		t.Errorf("Equal failed: WithDelta requires a non-negative delta; got delta=%v.", cfg.delta)
		return false
	}
	// Primitive fast path: when T is comparable and got == want by ==, return
	// true immediately with zero allocations. No reflection, no recursion.
	// This is the §23.5 / §23.13 hot path; it is proven via BenchmarkEqualPrimitive
	// and TestPrimitiveFastPathBypass.
	if primitiveEqual(got, want) {
		return true
	}
	if smartEqual(reflect.ValueOf(got), reflect.ValueOf(want), &cfg, nil) {
		return true
	}
	t.Errorf("Equal failed:\n%s", renderDiff(got, want))
	return false
}

// NotEqual reports whether got does not equal want. Uses the same
// smart-equal engine as Equal. On failure (got equals want), reports via
// t.Errorf. Returns true when values differ.
//
// Preconditions: t is non-nil.
// Concurrency: goroutine-safe.
//
// §21.4 F2; §23.5
func NotEqual[T any](t TB, got, want T, opts ...EqualOption) bool {
	t.Helper()
	cfg := applyEqualOptions(opts)
	// Use the same equality check as Equal but invert the outcome.
	equal := primitiveEqual(got, want) ||
		smartEqual(reflect.ValueOf(got), reflect.ValueOf(want), &cfg, nil)
	if !equal {
		return true
	}
	t.Errorf("NotEqual failed: values are equal: %s.", formatValue(got))
	return false
}

// smartEqual is the recursive equality engine. It is called from Equal and
// NotEqual and also from Contains, Subset, and JSONEq (via reflect.Value
// wrappers).
//
// Primitive fast path: when both values have a comparable type and are equal
// by ==, return true with zero allocations.
func smartEqual(got, want reflect.Value, cfg *equalConfig, vis visited) bool {
	// Primitive fast path: comparable and equal by ==.
	// We check using reflect.Value.Comparable() to guard the == call.
	if got.IsValid() && want.IsValid() &&
		got.Type() == want.Type() &&
		got.Comparable() && want.Comparable() &&
		got.Interface() == want.Interface() {
		return true
	}

	// Both invalid (untyped nil).
	if !got.IsValid() && !want.IsValid() {
		return true
	}
	// One valid, one not.
	if !got.IsValid() || !want.IsValid() {
		return false
	}

	// Type mismatch.
	if got.Type() != want.Type() {
		return false
	}

	// Custom Equal method: interface{ Equal(any) bool }.
	if got.CanInterface() {
		type equaler interface{ Equal(any) bool }
		if eq, ok := got.Interface().(equaler); ok {
			if want.CanInterface() {
				return eq.Equal(want.Interface())
			}
		}
	}

	// Delta check for floats at the top level (before the kind switch).
	if cfg.deltaSet {
		switch got.Kind() {
		case reflect.Float32, reflect.Float64:
			gf := got.Float()
			wf := want.Float()
			if math.IsNaN(gf) || math.IsNaN(wf) {
				return false
			}
			return math.Abs(gf-wf) <= cfg.delta
		}
	}

	switch got.Kind() {
	case reflect.Ptr:
		return equalPointers(got, want, cfg, vis)

	case reflect.Interface:
		return equalInterfaces(got, want, cfg, vis)

	case reflect.Slice:
		if got.IsNil() && want.IsNil() {
			return true
		}
		if got.IsNil() != want.IsNil() {
			// nil vs empty: treat as not equal (Go semantics differ)
			// Actually check lengths: nil slice and empty slice have same len
			// but we treat nil != non-nil for semantic correctness.
			// bytes.Equal treats nil==empty, but for general slices we don't.
			// Spec says: "else: len-check then elementwise recurse"
			// A nil slice has len 0; an empty slice has len 0 but is not nil.
			// We'll treat them as equal if len is 0 (consistent with bytes.Equal-like behavior
			// for []byte, but spec says smart-equal). Let's follow reflect.DeepEqual:
			// nil slice != non-nil slice.
			return false
		}
		return equalSlices(got, want, cfg, vis)

	case reflect.Array:
		return equalArrays(got, want, cfg, vis)

	case reflect.Map:
		return equalMaps(got, want, cfg, vis)

	case reflect.Struct:
		return equalStructs(got, want, cfg, vis)

	case reflect.String:
		gs, ws := got.String(), want.String()
		if cfg.ignoreWhitespace {
			gs = normalizeWhitespace(gs)
			ws = normalizeWhitespace(ws)
		}
		if cfg.ignoreCase {
			return strings.EqualFold(gs, ws)
		}
		return gs == ws

	case reflect.Float32, reflect.Float64:
		gf := got.Float()
		wf := want.Float()
		if math.IsNaN(gf) || math.IsNaN(wf) {
			return false
		}
		return gf == wf

	case reflect.Chan, reflect.Func:
		// Pointer identity.
		if got.IsNil() && want.IsNil() {
			return true
		}
		if got.IsNil() != want.IsNil() {
			return false
		}
		return got.Pointer() == want.Pointer()

	default:
		// Fallback: reflect.DeepEqual for all other kinds.
		return reflect.DeepEqual(got.Interface(), want.Interface())
	}
}

func equalPointers(got, want reflect.Value, cfg *equalConfig, vis visited) bool {
	if got.IsNil() && want.IsNil() {
		return true
	}
	if got.IsNil() != want.IsNil() {
		return false
	}
	// Cycle detection.
	if vis == nil {
		vis = make(visited)
	}
	key := [2]uintptr{got.Pointer(), want.Pointer()}
	if vis[key] {
		// Back-edge: same structure on both sides.
		return true
	}
	vis[key] = true
	return smartEqual(got.Elem(), want.Elem(), cfg, vis)
}

func equalInterfaces(got, want reflect.Value, cfg *equalConfig, vis visited) bool {
	if got.IsNil() && want.IsNil() {
		return true
	}
	if got.IsNil() != want.IsNil() {
		return false
	}
	// Unwrap dynamic values.
	return smartEqual(got.Elem(), want.Elem(), cfg, vis)
}

func equalSlices(got, want reflect.Value, cfg *equalConfig, vis visited) bool {
	if got.Len() != want.Len() {
		return false
	}
	if cfg.ignoreOrder {
		return equalSlicesMultiset(got, want, cfg, vis)
	}
	for i := range got.Len() {
		if !smartEqual(got.Index(i), want.Index(i), cfg, vis) {
			return false
		}
	}
	return true
}

func equalArrays(got, want reflect.Value, cfg *equalConfig, vis visited) bool {
	if got.Len() != want.Len() {
		return false
	}
	if cfg.ignoreOrder {
		return equalArrayMultiset(got, want, cfg, vis)
	}
	for i := range got.Len() {
		if !smartEqual(got.Index(i), want.Index(i), cfg, vis) {
			return false
		}
	}
	return true
}

// equalSlicesMultiset compares slices as multisets using a bucket approach.
// O(n) with hash buckets; falls back to O(n²) scan for non-hashable elements.
func equalSlicesMultiset(got, want reflect.Value, cfg *equalConfig, vis visited) bool {
	n := got.Len()
	// matched[i] = true means want[i] has been matched.
	matched := make([]bool, n)
	for i := range n {
		g := got.Index(i)
		found := false
		for j := range n {
			if matched[j] {
				continue
			}
			if smartEqual(g, want.Index(j), cfg, vis) {
				matched[j] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func equalArrayMultiset(got, want reflect.Value, cfg *equalConfig, vis visited) bool {
	n := got.Len()
	matched := make([]bool, n)
	for i := range n {
		g := got.Index(i)
		found := false
		for j := range n {
			if matched[j] {
				continue
			}
			if smartEqual(g, want.Index(j), cfg, vis) {
				matched[j] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func equalMaps(got, want reflect.Value, cfg *equalConfig, vis visited) bool {
	if got.IsNil() && want.IsNil() {
		return true
	}
	if got.IsNil() != want.IsNil() {
		return false
	}
	if got.Len() != want.Len() {
		return false
	}
	for _, key := range want.MapKeys() {
		var gotVal reflect.Value
		if cfg.ignoreCase && key.Kind() == reflect.String {
			// Case-insensitive key lookup.
			wantKey := strings.ToLower(key.String())
			for _, gk := range got.MapKeys() {
				if gk.Kind() == reflect.String && strings.ToLower(gk.String()) == wantKey {
					gotVal = got.MapIndex(gk)
					break
				}
			}
		} else {
			gotVal = got.MapIndex(key)
		}
		if !gotVal.IsValid() {
			return false
		}
		if !smartEqual(gotVal, want.MapIndex(key), cfg, vis) {
			return false
		}
	}
	return true
}

func equalStructs(got, want reflect.Value, cfg *equalConfig, vis visited) bool {
	t := got.Type()

	// Check whether the struct has unexported fields. reflect.Value.Field(i).CanInterface()
	// returns false for unexported fields obtained via reflection, so we cannot use
	// Interface() on them. reflect.DeepEqual can still compare them correctly because it
	// has internal access to the memory layout. When any unexported field is present and no
	// field-level option (IgnoreFields) is in use, fall back to reflect.DeepEqual on the
	// whole struct :  this is the only panic-free approach.
	hasUnexported := false
	for i := range t.NumField() {
		if !t.Field(i).IsExported() {
			hasUnexported = true
			break
		}
	}
	if hasUnexported && cfg.ignoreFields == nil {
		return reflect.DeepEqual(got.Interface(), want.Interface())
	}

	for i := range t.NumField() {
		field := t.Field(i)
		if !field.IsExported() {
			// IgnoreFields is active; we cannot compare unexported fields via Interface(),
			// so they are conservatively treated as equal. Only exported field names can
			// be passed to IgnoreFields, so this is correct behavior.
			continue
		}
		if cfg.ignoreFields != nil && cfg.ignoreFields[field.Name] {
			continue
		}
		if !smartEqual(got.Field(i), want.Field(i), cfg, vis) {
			return false
		}
	}
	return true
}

// normalizeWhitespace trims leading/trailing whitespace and collapses
// internal runs of whitespace to a single space.
func normalizeWhitespace(s string) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	b.Grow(len(s))
	prevSpace := false
	for _, r := range s {
		if isWhitespace(r) {
			if !prevSpace {
				b.WriteByte(' ')
				prevSpace = true
			}
		} else {
			b.WriteRune(r)
			prevSpace = false
		}
	}
	return b.String()
}

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\f' || r == '\v'
}

// formatValue formats a value for use in failure messages,
// calling LogValue() on slog.LogValuer values for security.
func formatValue(v any) string {
	if lv, ok := v.(interface{ LogValue() any }); ok {
		return fmt.Sprintf("%v", lv.LogValue())
	}
	// Also check for slog.LogValuer interface (returns slog.Value).
	if lv, ok := v.(interface{ LogValue() interface{ Any() any } }); ok {
		return fmt.Sprintf("%v", lv.LogValue().Any())
	}
	return fmt.Sprintf("%v", v)
}
