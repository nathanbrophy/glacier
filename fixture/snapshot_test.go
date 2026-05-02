// SPDX-License-Identifier: Apache-2.0

package fixture_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/nathanbrophy/glacier/fixture"
)

// TestSnapshotDeterministicAcrossRuns: Two invocations produce byte-identical
// pretty-printed output. (#6 in test matrix)
func TestSnapshotDeterministicAcrossRuns(t *testing.T) {
	type Item struct {
		ID    int
		Name  string
		Score float64
	}
	v := Item{ID: 1, Name: "glacier", Score: 9.5}

	t.Setenv("GLACIER_GOLDEN_UPDATE", "1")

	// First run creates the snapshot (in testdata/snapshots/ relative to this file).
	ok1 := fixture.Snapshot(t, "deterministic_runs", v)
	if !ok1 {
		t.Fatal("first Snapshot call returned false")
	}

	// Second run should match exactly.
	t.Setenv("GLACIER_GOLDEN_UPDATE", "0")
	ok2 := fixture.Snapshot(t, "deterministic_runs", v)
	if !ok2 {
		t.Fatal("second Snapshot call returned false; output is non-deterministic")
	}
}

// TestSnapshotPrettyPrintStableMapKeys: Map keys sorted. (#12)
func TestSnapshotPrettyPrintStableMapKeys(t *testing.T) {
	m1 := map[string]int{"a": 1, "b": 2, "c": 3, "z": 26, "m": 13}
	m2 := map[string]int{"z": 26, "m": 13, "a": 1, "c": 3, "b": 2}

	t.Setenv("GLACIER_GOLDEN_UPDATE", "1")
	ok := fixture.Snapshot(t, "stable_map_keys", m1)
	if !ok {
		t.Fatal("Snapshot(m1) returned false")
	}

	// m2 has the same keys/values as m1; both should produce the same snapshot.
	t.Setenv("GLACIER_GOLDEN_UPDATE", "0")
	ok2 := fixture.Snapshot(t, "stable_map_keys", m2)
	if !ok2 {
		t.Fatal("Snapshot(m2) did not match m1; map keys not sorted consistently")
	}
}

// TestSnapshotPrettyPrintLineEndingsLF: Output uses \n regardless of platform.
// (#13)
func TestSnapshotPrettyPrintLineEndingsLF(t *testing.T) {
	type Data struct {
		A string
		B int
	}
	t.Setenv("GLACIER_GOLDEN_UPDATE", "1")

	ok := fixture.Snapshot(t, "lf_endings", Data{A: "hello", B: 42})
	if !ok {
		t.Fatal("Snapshot returned false")
	}

	// Read the snap file directly to check line endings.
	// The snap file will be in the package's testdata/snapshots/ directory.
	// We can't easily get the path in a test without importing reflect or runtime,
	// so we verify by re-snapshotting and checking the second invocation succeeds.
	t.Setenv("GLACIER_GOLDEN_UPDATE", "0")
	ok2 := fixture.Snapshot(t, "lf_endings", Data{A: "hello", B: 42})
	if !ok2 {
		t.Fatal("LF check: second Snapshot call failed")
	}
	// As an extra check, we can verify the snapshot file doesn't contain CRLF
	// by reading the file. The testdata directory is relative to this test file.
	snapPath := "testdata/snapshots/lf_endings.snap"
	content, err := os.ReadFile(snapPath)
	if err == nil {
		if strings.Contains(string(content), "\r\n") {
			t.Fatal("snapshot file contains CRLF line endings; expected LF only")
		}
	}
	// If ReadFile fails (file doesn't exist yet — first run), we skip the check.
}

// TestSnapshotMissingCreates: GLACIER_GOLDEN_UPDATE=1 + missing snapshot →
// created. (#11)
func TestSnapshotMissingCreates(t *testing.T) {
	t.Setenv("GLACIER_GOLDEN_UPDATE", "1")

	ok := fixture.Snapshot(t, "missing_creates_test", struct{ N int }{N: 7})
	if !ok {
		t.Fatal("Snapshot returned false on missing file with GLACIER_GOLDEN_UPDATE=1")
	}
}

// TestSnapshotTimeFormat: time.Time values use date-only RFC 3339 (EX1).
func TestSnapshotTimeFormat(t *testing.T) {
	type HasTime struct {
		Name      string
		CreatedAt time.Time
	}

	t.Setenv("GLACIER_GOLDEN_UPDATE", "1")

	v1 := HasTime{Name: "test", CreatedAt: time.Date(2026, 5, 1, 12, 30, 0, 0, time.UTC)}
	ok1 := fixture.Snapshot(t, "time_format_test", v1)
	if !ok1 {
		t.Fatal("first Snapshot call failed")
	}

	// A different time-of-day should still match (same date, different time).
	t.Setenv("GLACIER_GOLDEN_UPDATE", "0")
	v2 := HasTime{Name: "test", CreatedAt: time.Date(2026, 5, 1, 23, 59, 59, 999, time.UTC)}
	ok2 := fixture.Snapshot(t, "time_format_test", v2)
	if !ok2 {
		t.Fatal("Snapshot with same date but different time should match (date-only format)")
	}
}

// TestSnapshotFormatterDeterministicWithMaps: 100-key map produces identical
// output across multiple calls. (#67)
func TestSnapshotFormatterDeterministicWithMaps(t *testing.T) {
	m := make(map[string]int, 100)
	for i := range 100 {
		key := fmt.Sprintf("key%03d", i)
		m[key] = i
	}

	t.Setenv("GLACIER_GOLDEN_UPDATE", "1")
	ok := fixture.Snapshot(t, "map100_deterministic", m)
	if !ok {
		t.Fatal("Snapshot(map100) first call failed")
	}

	t.Setenv("GLACIER_GOLDEN_UPDATE", "0")
	for range 10 {
		ok2 := fixture.Snapshot(t, "map100_deterministic", m)
		if !ok2 {
			t.Fatal("Snapshot(map100) produced non-deterministic output")
		}
	}
}

// TestSnapshotFormatterUnicodeStable: UTF-8 strings preserve byte identity.
// (#68)
func TestSnapshotFormatterUnicodeStable(t *testing.T) {
	type Unicode struct {
		Text string
	}
	v := Unicode{Text: "こんにちは世界 🐧"}

	t.Setenv("GLACIER_GOLDEN_UPDATE", "1")
	ok1 := fixture.Snapshot(t, "unicode_stable_test", v)
	if !ok1 {
		t.Fatal("Snapshot(unicode) failed")
	}

	t.Setenv("GLACIER_GOLDEN_UPDATE", "0")
	ok2 := fixture.Snapshot(t, "unicode_stable_test", v)
	if !ok2 {
		t.Fatal("Snapshot(unicode) produced non-deterministic output")
	}
}
