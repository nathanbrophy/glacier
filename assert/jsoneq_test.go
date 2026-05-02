// SPDX-License-Identifier: Apache-2.0

package assert

import "testing"

// §21.4 F7, E17, E18

func TestJSONEqIdentical(t *testing.T) {
	mt := &mockTB{}
	got := []byte(`{"name":"Ada","age":36}`)
	want := []byte(`{"name":"Ada","age":36}`)
	True(t, JSONEq(mt, got, want), "JSONEq: identical JSON")
}

func TestJSONEqKeyOrderInvariant(t *testing.T) {
	mt := &mockTB{}
	got := []byte(`{"age":36,"name":"Ada"}`)
	want := []byte(`{"name":"Ada","age":36}`)
	True(t, JSONEq(mt, got, want), "JSONEq: different key order → equal")
}

func TestJSONEqArrayOrderMatters(t *testing.T) {
	mt := &mockTB{}
	got := []byte(`{"tags":["math","logic"]}`)
	want := []byte(`{"tags":["logic","math"]}`)
	False(t, JSONEq(mt, got, want), "JSONEq: array order matters by default")
}

func TestJSONEqArrayIgnoreOrder(t *testing.T) {
	mt := &mockTB{}
	got := []byte(`{"tags":["math","logic"]}`)
	want := []byte(`{"tags":["logic","math"]}`)
	True(t, JSONEq(mt, got, want, IgnoreOrder()), "JSONEq: IgnoreOrder ignores array order")
}

func TestJSONEqIgnoreCaseValues(t *testing.T) {
	mt := &mockTB{}
	got := []byte(`{"name":"ADA"}`)
	want := []byte(`{"name":"ada"}`)
	True(t, JSONEq(mt, got, want, IgnoreCase()), "JSONEq: IgnoreCase folds string values")
}

func TestJSONEqMalformedGot(t *testing.T) {
	mt := &mockTB{}
	False(t, JSONEq(mt, []byte(`{invalid`), []byte(`{}`)), "JSONEq: malformed got")
	Equal(t, mt.errorfCalls, 1)
}

func TestJSONEqMalformedWant(t *testing.T) {
	mt := &mockTB{}
	False(t, JSONEq(mt, []byte(`{}`), []byte(`{invalid`)), "JSONEq: malformed want")
	Equal(t, mt.errorfCalls, 1)
}

// L-add-13: JSONEq with embedded null values vs missing keys.
func TestJSONEqNullVsMissingKey(t *testing.T) {
	mt := &mockTB{}
	// {"a": null} != {} — they are not equal.
	False(t, JSONEq(mt, []byte(`{"a":null}`), []byte(`{}`)), "JSONEq: null value != missing key")
}

func TestJSONEqNilTreatedAsNull(t *testing.T) {
	mt := &mockTB{}
	True(t, JSONEq(mt, nil, []byte(`null`)), "JSONEq: nil == JSON null")
}
