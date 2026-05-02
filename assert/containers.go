// SPDX-License-Identifier: Apache-2.0

package assert

import (
	"fmt"
	"reflect"
	"strings"
)

// Contains reports whether haystack contains needle. haystack may be a
// string, []T, or map[K]V; needle is matched against elements or keys using
// the smart-equal engine (and any supplied opts). On failure reports via
// t.Errorf.
//
// Supported types for haystack:
//   - string: needle must be a string; reports strings.Contains.
//   - []T: needle is an element; smart-equal comparison.
//   - map[K]V: needle is a key K; smart-equal key lookup.
//
// On unsupported haystack type, reports via t.Errorf and returns false.
//
// Preconditions: t is non-nil.
// Concurrency: goroutine-safe.
//
// §21.4 F3
func Contains(t TB, haystack, needle any, opts ...EqualOption) bool {
	t.Helper()
	cfg := applyEqualOptions(opts)
	rv := reflect.ValueOf(haystack)

	switch {
	case haystack == nil:
		t.Errorf("Contains failed: haystack is nil.")
		return false

	case rv.Kind() == reflect.String:
		ns, ok := needle.(string)
		if !ok {
			t.Errorf("Contains failed: haystack is string but needle is %T (must be string).", needle)
			return false
		}
		hs := rv.String()
		if cfg.ignoreCase {
			if strings.Contains(strings.ToLower(hs), strings.ToLower(ns)) {
				return true
			}
		} else if strings.Contains(hs, ns) {
			return true
		}
		t.Errorf("Contains failed: %q does not contain %q.", hs, ns)
		return false

	case rv.Kind() == reflect.Slice:
		nv := reflect.ValueOf(needle)
		for i := range rv.Len() {
			if smartEqual(rv.Index(i), nv, &cfg, nil) {
				return true
			}
		}
		t.Errorf("Contains failed: slice does not contain element %s.", formatValue(needle))
		return false

	case rv.Kind() == reflect.Map:
		nv := reflect.ValueOf(needle)
		for _, k := range rv.MapKeys() {
			if smartEqual(k, nv, &cfg, nil) {
				return true
			}
		}
		t.Errorf("Contains failed: map does not contain key %s.", formatValue(needle))
		return false

	default:
		t.Errorf("Contains failed: unsupported haystack type %T (want string, slice, or map).", haystack)
		return false
	}
}

// Len reports whether the length of container equals want. container may be
// a slice, array, map, string, or channel. On failure reports the actual
// length via t.Errorf. On unsupported type, reports a type error and returns
// false.
//
// Len is non-generic: it accepts any and uses reflect to determine the
// length, so it works uniformly across slices, maps, strings, and channels
// without requiring the caller to specify a type parameter.
//
// Preconditions: t is non-nil; want >= 0.
// Concurrency: goroutine-safe.
//
// §21.4 F3; §23.17
func Len(t TB, container any, want int, msg ...any) bool {
	t.Helper()
	if container == nil {
		t.Errorf("Len failed: container is nil.")
		return false
	}
	rv := reflect.ValueOf(container)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String, reflect.Chan:
		got := rv.Len()
		if got == want {
			return true
		}
		suffix := fmtMsg(msg)
		t.Errorf("Len failed: length=%d, want %d.%s", got, want, suffix)
		return false
	default:
		t.Errorf("Len failed: unsupported container type %T (want slice, array, map, string, or chan).",
			container)
		return false
	}
}

// Subset reports whether every element of want appears in got using the
// smart-equal engine. T is constrained to any, giving compile-time type
// match. An empty want slice always returns true.
//
// Preconditions: t is non-nil.
// Concurrency: goroutine-safe.
//
// §21.4 F9, E19
func Subset[T any](t TB, got, want []T, opts ...EqualOption) bool {
	t.Helper()
	cfg := applyEqualOptions(opts)
	for _, w := range want {
		wv := reflect.ValueOf(w)
		found := false
		for _, g := range got {
			gv := reflect.ValueOf(g)
			if smartEqual(gv, wv, &cfg, nil) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Subset failed: element %s not found in got.", fmt.Sprintf("%#v", w))
			return false
		}
	}
	return true
}
