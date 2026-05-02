// SPDX-License-Identifier: Apache-2.0

package fixture_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/nathanbrophy/glacier/fixture"
)

// TestLoadFile: Load returns correct bytes from testdata/. (#14)
func TestLoadFile(t *testing.T) {
	dir := t.TempDir()
	content := []byte("fixture load test content")
	if err := os.WriteFile(filepath.Join(dir, "data.txt"), content, 0o644); err != nil {
		t.Fatal(err)
	}
	got := fixture.Load(t, "data.txt", fixture.WithRoot(dir))
	if string(got) != string(content) {
		t.Fatalf("Load returned %q; want %q", got, content)
	}
}

// TestLoadFileMissingFatals: Load on missing file calls t.Fatal. (#15)
func TestLoadFileMissingFatals(t *testing.T) {
	dir := t.TempDir()
	m := newMockTB()
	panicked := callAndRecover(func() {
		fixture.Load(m, "nonexistent.txt", fixture.WithRoot(dir))
	})
	if !m.Failed() && !panicked {
		t.Fatal("Load did not fatal on missing file")
	}
}

// TestLoadJSONUnmarshals: LoadJSON[T] returns correctly populated struct. (#16)
func TestLoadJSONUnmarshals(t *testing.T) {
	type Point struct {
		X int
		Y int
	}
	dir := t.TempDir()
	data, _ := json.Marshal(Point{X: 10, Y: 20})
	if err := os.WriteFile(filepath.Join(dir, "point.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	got := fixture.LoadJSON[Point](t, "point.json", fixture.WithRoot(dir))
	if got.X != 10 || got.Y != 20 {
		t.Fatalf("LoadJSON returned %+v; want {X:10 Y:20}", got)
	}
}

// TestLoadJSONBadJSONFatals: LoadJSON on malformed JSON calls t.Fatal. (#17)
func TestLoadJSONBadJSONFatals(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bad.json"), []byte("{invalid json}"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := newMockTB()
	panicked := callAndRecover(func() {
		fixture.LoadJSON[map[string]any](m, "bad.json", fixture.WithRoot(dir))
	})
	if !m.Failed() && !panicked {
		t.Fatal("LoadJSON did not fatal on bad JSON")
	}
}

// TestLoadJSONWrongTypeFatals: LoadJSON[T] on mismatched schema calls t.Fatal. (#18)
func TestLoadJSONWrongTypeFatals(t *testing.T) {
	type MyStruct struct {
		Value int
	}
	dir := t.TempDir()
	// JSON for a string, not an int.
	if err := os.WriteFile(filepath.Join(dir, "wrong.json"), []byte(`{"Value":"not-an-int"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	m := newMockTB()
	panicked := callAndRecover(func() {
		fixture.LoadJSON[MyStruct](m, "wrong.json", fixture.WithRoot(dir))
	})
	if !m.Failed() && !panicked {
		t.Fatal("LoadJSON did not fatal on type mismatch")
	}
}

// Table-driven Load tests.
func TestLoadTableDriven(t *testing.T) {
	cases := []struct {
		name    string
		content []byte
	}{
		{"text", []byte("hello world")},
		{"binary", []byte{0x00, 0xFF, 0xAB, 0xCD}},
		{"empty", []byte{}},
		{"unicode", []byte("こんにちは世界")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "file"), tc.content, 0o644); err != nil {
				t.Fatal(err)
			}
			got := fixture.Load(t, "file", fixture.WithRoot(dir))
			if string(got) != string(tc.content) {
				t.Fatalf("Load mismatch: got %q, want %q", got, tc.content)
			}
		})
	}
}
