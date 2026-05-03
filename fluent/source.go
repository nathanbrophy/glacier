// SPDX-License-Identifier: Apache-2.0

package fluent

import (
	"bufio"
	"io"
	"iter"
	"strings"
)

// From returns an iterator that yields the elements of s in order.
//
// Preconditions: none. From(nil) yields nothing.
// Concurrency: safe; each call creates an independent closure over the slice header.
func From[T any](s []T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range s {
			if !yield(v) {
				return
			}
		}
	}
}

// FromMap returns an iterator that yields every key/value pair in m.
// Iteration order is undefined (Go map semantics).
//
// Preconditions: none. FromMap(nil) yields nothing.
// Concurrency: safe for concurrent calls; callers must not modify m during iteration.
func FromMap[K comparable, V any](m map[K]V) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range m {
			if !yield(k, v) {
				return
			}
		}
	}
}

// FromChan returns an iterator that yields values received from ch until ch
// is closed. The iterating goroutine blocks on each receive.
//
// Preconditions: ch must eventually be closed; otherwise the iterator never terminates.
// Concurrency: each call returns an independent iterator; two goroutines must not
// share the same iterator value. Multiple goroutines may each call FromChan on the
// same channel (they will race for values; this is channel semantics, not fluent's).
func FromChan[T any](ch <-chan T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range ch {
			if !yield(v) {
				return
			}
		}
	}
}

// Range returns a lazy half-open arithmetic sequence [start, stop) with the
// given step. The sequence is empty when start == stop, or when the step
// direction mismatches (e.g., start < stop with step < 0).
//
// Preconditions: step must not be zero; Range panics with
// "fluent: Range: step must be non-zero" if step == 0.
// Concurrency: safe.
func Range(start, stop, step int) iter.Seq[int] {
	if step == 0 {
		//glacier:nolint=panic-in-library programmer error: zero step is documented as a panic precondition.
		panic("fluent: Range: step must be non-zero")
	}
	return func(yield func(int) bool) {
		for i := start; (step > 0 && i < stop) || (step < 0 && i > stop); i += step {
			if !yield(i) {
				return
			}
		}
	}
}

// Repeat returns an iterator that yields v exactly n times.
//
// Preconditions: n >= 0. Repeat(v, 0) yields nothing; Repeat(v, n) for n < 0
// yields nothing.
// Concurrency: safe.
func Repeat[T any](v T, n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		for i := 0; i < n; i++ {
			if !yield(v) {
				return
			}
		}
	}
}

// Generate returns a lazy iterator driven by fn. On each pull, fn is called;
// if fn returns (v, true), v is yielded. If fn returns (_, false), iteration
// ends.
//
// Generate(fn) can produce an infinite sequence if fn never returns false;
// callers must use Take or break to bound iteration.
//
// Preconditions: fn must not be nil.
// Concurrency: safe if fn is goroutine-safe; Generate does not add synchronization.
func Generate[T any](fn func() (T, bool)) iter.Seq[T] {
	return func(yield func(T) bool) {
		for {
			v, ok := fn()
			if !ok {
				return
			}
			if !yield(v) {
				return
			}
		}
	}
}

// Lines returns a lazy iterator that yields newline-delimited lines from r.
// Each yielded string has the trailing newline (and any \r) stripped.
// Lines uses bufio.Scanner internally; lines exceeding bufio.MaxScanTokenSize
// (64 KiB by default) cause the iterator to stop and the error to be silently
// discarded. Callers that need to detect scan errors should wrap r.
//
// Security: Lines does not enforce an overall reader size cap. Callers reading
// from untrusted sources must wrap r with io.LimitReader before calling Lines
// (see Glacier's untrusted-input register, row 14).
//
// Preconditions: r must not be nil.
// Concurrency: safe; the iterating goroutine owns r during iteration.
func Lines(r io.Reader) iter.Seq[string] {
	return func(yield func(string) bool) {
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			if !yield(sc.Text()) {
				return
			}
		}
	}
}

// Words returns a lazy iterator that yields whitespace-separated tokens from r.
// Uses bufio.Scanner with ScanWords.
//
// Security: same size-cap note as Lines (untrusted-input register row 15).
//
// Preconditions: r must not be nil.
// Concurrency: safe.
func Words(r io.Reader) iter.Seq[string] {
	return func(yield func(string) bool) {
		sc := bufio.NewScanner(r)
		sc.Split(bufio.ScanWords)
		for sc.Scan() {
			if !yield(sc.Text()) {
				return
			}
		}
	}
}

// Split returns a lazy iterator that yields the parts of s split on sep via
// successive strings.Index walks. No intermediate slice is allocated.
//
// Split("a,b,c", ",") yields "a", "b", "c".
// Split("abc", ",") yields "abc".
// Split("", ",") yields one empty string.
//
// Preconditions: sep must not be empty; Split panics with
// "fluent: Split: separator must not be empty" if sep == "".
// Concurrency: safe.
func Split(s, sep string) iter.Seq[string] {
	if sep == "" {
		//glacier:nolint=panic-in-library programmer error: empty separator is documented as a panic precondition.
		panic("fluent: Split: separator must not be empty")
	}
	return func(yield func(string) bool) {
		for {
			idx := strings.Index(s, sep)
			if idx < 0 {
				yield(s)
				return
			}
			if !yield(s[:idx]) {
				return
			}
			s = s[idx+len(sep):]
		}
	}
}

// Pairs converts a Seq[KV[K, V]] into a Seq2[K, V].
//
// Pairs is the inverse of Entries: Entries(Pairs(seq)) round-trips.
//
// Preconditions: none.
// Concurrency: safe.
func Pairs[K, V any](s iter.Seq[KV[K, V]]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for kv := range s {
			if !yield(kv.K, kv.V) {
				return
			}
		}
	}
}
