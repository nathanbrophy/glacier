// SPDX-License-Identifier: Apache-2.0

package assert

import (
	"bytes"
	"encoding/json"
	"reflect"
)

// JSONEq parses got and want as JSON and reports deep equality of the
// resulting values. JSON object key order is always ignored (maps are
// unordered). IgnoreOrder() additionally ignores JSON array element order.
// IgnoreCase() applies case folding to string values and object keys.
//
// On JSON parse failure, reports the parse error via t.Errorf and returns
// false without comparing.
//
// Preconditions: t is non-nil; got and want are valid JSON bytes (nil is
// treated as the JSON null).
// Concurrency: goroutine-safe.
//
// §21.4 F7, E17, E18
func JSONEq(t TB, got, want []byte, opts ...EqualOption) bool {
	t.Helper()
	var gv, wv any
	if err := json.Unmarshal(normalizeJSON(got), &gv); err != nil {
		t.Errorf("JSONEq failed: cannot parse got as JSON: %s.", err.Error())
		return false
	}
	if err := json.Unmarshal(normalizeJSON(want), &wv); err != nil {
		t.Errorf("JSONEq failed: cannot parse want as JSON: %s.", err.Error())
		return false
	}
	cfg := applyEqualOptions(opts)
	if smartEqual(reflect.ValueOf(gv), reflect.ValueOf(wv), &cfg, nil) {
		return true
	}
	t.Errorf("JSONEq failed:\n%s", renderDiff(gv, wv))
	return false
}

// normalizeJSON treats nil as the JSON null literal.
func normalizeJSON(b []byte) []byte {
	if b == nil {
		return []byte("null")
	}
	return b
}

// BytesEq reports whether got and want are byte-for-byte equal via
// bytes.Equal. nil and empty slices are considered equal (Go semantics for
// bytes.Equal). On failure reports lengths and a hex diff prefix via
// t.Errorf.
//
// Concurrency: goroutine-safe.
//
// §21.4 F8
func BytesEq(t TB, got, want []byte, msg ...any) bool {
	t.Helper()
	if bytes.Equal(got, want) {
		return true
	}
	suffix := fmtMsg(msg)
	// Show a short hex prefix of the difference.
	limit := 16
	gHex := hexPrefix(got, limit)
	wHex := hexPrefix(want, limit)
	t.Errorf("BytesEq failed: byte slices differ (len got=%d, len want=%d).\n  got:  %s\n  want: %s.%s",
		len(got), len(want), gHex, wHex, suffix)
	return false
}

// hexPrefix returns a hex string of up to n bytes from b.
func hexPrefix(b []byte, n int) string {
	if len(b) > n {
		b = b[:n]
	}
	const hexChars = "0123456789abcdef"
	buf := make([]byte, 0, len(b)*3)
	for i, v := range b {
		if i > 0 {
			buf = append(buf, ' ')
		}
		buf = append(buf, hexChars[v>>4], hexChars[v&0xf])
	}
	return string(buf)
}
