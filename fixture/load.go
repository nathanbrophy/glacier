// SPDX-License-Identifier: Apache-2.0

package fixture

import (
	"encoding/json"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/internal/safefile"
)

// Load reads testdata/<name> and returns its bytes. Calls t.Fatal on any
// error (file not found, permission denied, safefile rejection). Registers
// t.Helper.
//
// name must be a relative path; ".." components and absolute paths are
// rejected by safefile.
func Load(t assert.TB, name string, opts ...GoldenOption) []byte {
	t.Helper()
	cfg, err := applyGoldenOptions(opts)
	if err != nil {
		t.Fatalf("fixture: Load: %v", err)
		return nil
	}
	root, err := resolveRoot(cfg, 1)
	if err != nil {
		t.Fatalf("fixture: Load: %v", err)
		return nil
	}
	data, err := safefile.ReadFile(root, name)
	if err != nil {
		t.Fatalf("fixture: Load: read testdata/%s: %v", name, err)
		return nil
	}
	return data
}

// LoadJSON[T] reads testdata/<name> and unmarshals it as JSON into T.
// Calls t.Fatal on read or unmarshal error. Registers t.Helper.
//
// name must be a relative path; ".." components are rejected by safefile.
// The file content is developer-committed, not untrusted input; no depth cap
// is applied (see spec 0010 §Schema row 14).
func LoadJSON[T any](t assert.TB, name string, opts ...GoldenOption) T {
	t.Helper()
	var zero T
	data := Load(t, name, opts...)
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("fixture: LoadJSON: unmarshal testdata/%s into %T: %v", name, zero, err)
		return zero
	}
	return v
}
