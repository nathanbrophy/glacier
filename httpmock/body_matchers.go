// SPDX-License-Identifier: Apache-2.0

package httpmock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/nathanbrophy/glacier/assert"
)

// BodyMatcher matches a request body. Implementations must be goroutine-safe.
type BodyMatcher interface {
	Match(body []byte, contentType string) bool
	String() string
}

// BodyExact returns a BodyMatcher that requires byte-for-byte equality.
func BodyExact(body []byte) BodyMatcher {
	cp := make([]byte, len(body))
	copy(cp, body)
	return &exactMatcher{want: cp}
}

type exactMatcher struct{ want []byte }

// Match implements BodyMatcher.
func (m *exactMatcher) Match(body []byte, _ string) bool { return bytes.Equal(body, m.want) }

// String implements BodyMatcher.
func (m *exactMatcher) String() string {
	prefix := m.want
	if len(prefix) > 16 {
		prefix = prefix[:16]
	}
	return fmt.Sprintf("body exact: %x", prefix)
}

// BodyJSON returns a BodyMatcher that unmarshals the request body as JSON into
// a zero-value T and compares it to want.
func BodyJSON[T any](want T, opts ...assert.EqualOption) BodyMatcher {
	return &jsonMatcher[T]{want: want, opts: opts}
}

type jsonMatcher[T any] struct {
	want T
	opts []assert.EqualOption
}

// Match implements BodyMatcher.
func (m *jsonMatcher[T]) Match(body []byte, _ string) bool {
	var got T
	if err := json.Unmarshal(body, &got); err != nil {
		return false
	}
	if len(m.opts) == 0 {
		return reflect.DeepEqual(got, m.want)
	}
	tb := &noopTB{}
	return assert.Equal(tb, got, m.want, m.opts...)
}

// String implements BodyMatcher.
func (m *jsonMatcher[T]) String() string { return fmt.Sprintf("body JSON: %T", m.want) }

// noopTB satisfies assert.TB for use inside BodyJSON.Match where no *testing.T is available.
type noopTB struct{ failed bool }

// Helper implements assert.TB; no-op.
func (n *noopTB) Helper() {}

// Errorf implements assert.TB by recording the failure flag.
func (n *noopTB) Errorf(_ string, _ ...any) { n.failed = true }

// Fatalf implements assert.TB by recording the failure flag.
func (n *noopTB) Fatalf(_ string, _ ...any) { n.failed = true }

// FailNow implements assert.TB by recording the failure flag.
func (n *noopTB) FailNow() { n.failed = true }

// Cleanup implements assert.TB; no-op.
func (n *noopTB) Cleanup(_ func()) {}

// Name implements assert.TB and returns an empty name.
func (n *noopTB) Name() string { return "" }

// BodyContains returns a BodyMatcher that checks for a substring.
func BodyContains(s string) BodyMatcher { return &containsMatcher{sub: s} }

type containsMatcher struct{ sub string }

// Match implements BodyMatcher.
func (m *containsMatcher) Match(body []byte, _ string) bool {
	return strings.Contains(string(body), m.sub)
}

// String implements BodyMatcher.
func (m *containsMatcher) String() string { return fmt.Sprintf("body contains: %q", m.sub) }

// BodyMatchFn returns a BodyMatcher that delegates to f.
func BodyMatchFn(f func([]byte) bool) BodyMatcher { return &fnMatcher{f: f} }

type fnMatcher struct{ f func([]byte) bool }

// Match implements BodyMatcher.
func (m *fnMatcher) Match(body []byte, _ string) bool { return m.f(body) }

// String implements BodyMatcher.
func (m *fnMatcher) String() string { return fmt.Sprintf("body fn: %p", m.f) }
