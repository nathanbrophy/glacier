// SPDX-License-Identifier: Apache-2.0

package fixture_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nathanbrophy/glacier/fixture"
)

// TestGoldenCreateOnMissing: GLACIER_GOLDEN_UPDATE=1 + missing file → file
// created with bytes; returns true. (#1 in test matrix)
func TestGoldenCreateOnMissing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GLACIER_GOLDEN_UPDATE", "1")

	got := []byte("hello, glacier!")
	ok := fixture.Golden(t, "create_on_missing.txt", got, fixture.WithRoot(dir))
	if !ok {
		t.Fatal("Golden returned false; expected true on create")
	}
	written, err := os.ReadFile(filepath.Join(dir, "create_on_missing.txt"))
	if err != nil {
		t.Fatalf("golden file not created: %v", err)
	}
	if string(written) != string(got) {
		t.Fatalf("golden file content mismatch: got %q, want %q", written, got)
	}
}

// TestGoldenMissingNoUpdateErrors: Missing file + no env → t.Errorf with hint
// message; returns false. (#2 in test matrix)
func TestGoldenMissingNoUpdateErrors(t *testing.T) {
	dir := t.TempDir()
	m := newMockTB()
	ok := fixture.Golden(m, "nonexistent.txt", []byte("data"), fixture.WithRoot(dir))
	if ok {
		t.Fatal("Golden returned true; expected false on missing file without env")
	}
	if !m.Failed() {
		t.Fatal("mockTB not marked failed")
	}
	if !m.containsError("GLACIER_GOLDEN_UPDATE") {
		t.Fatalf("expected hint about GLACIER_GOLDEN_UPDATE in error; got: %v", m.allErrors())
	}
}

// TestGoldenMatchPassesSilently: Matching bytes → no error, returns true.
// (#3 in test matrix)
func TestGoldenMatchPassesSilently(t *testing.T) {
	dir := t.TempDir()
	content := []byte("exact match content")
	if err := os.WriteFile(filepath.Join(dir, "match.txt"), content, 0o644); err != nil {
		t.Fatal(err)
	}
	m := newMockTB()
	ok := fixture.Golden(m, "match.txt", content, fixture.WithRoot(dir))
	if !ok {
		t.Fatal("Golden returned false on exact match")
	}
	if m.Failed() {
		t.Fatalf("Golden reported errors on match: %v", m.allErrors())
	}
}

// TestGoldenMismatchReportsDiff: Mismatch → t.Errorf with diff. (#4 in test matrix)
func TestGoldenMismatchReportsDiff(t *testing.T) {
	dir := t.TempDir()
	want := []byte("line one\nline two\n")
	got := []byte("line one\nline THREE\n")
	if err := os.WriteFile(filepath.Join(dir, "mismatch.txt"), want, 0o644); err != nil {
		t.Fatal(err)
	}
	m := newMockTB()
	ok := fixture.Golden(m, "mismatch.txt", got, fixture.WithRoot(dir))
	if ok {
		t.Fatal("Golden returned true on mismatch")
	}
	if !m.Failed() {
		t.Fatal("mockTB not marked failed on mismatch")
	}
	// The error message should mention a diff.
	if !m.containsError("mismatch") {
		t.Fatalf("expected diff in error; got: %v", m.allErrors())
	}
}

// TestGoldenWithRoot: WithRoot redirects to alternate testdata dir. (#5 in test matrix)
func TestGoldenWithRoot(t *testing.T) {
	dir := t.TempDir()
	content := []byte("alternative root content")
	if err := os.WriteFile(filepath.Join(dir, "alt.txt"), content, 0o644); err != nil {
		t.Fatal(err)
	}
	ok := fixture.Golden(t, "alt.txt", content, fixture.WithRoot(dir))
	if !ok {
		t.Fatal("Golden returned false with WithRoot on exact match")
	}
}

// TestGoldenUpdateOverwritesExisting: GLACIER_GOLDEN_UPDATE=1 + changed got →
// file overwritten; returns true. (#24 in test matrix)
func TestGoldenUpdateOverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GLACIER_GOLDEN_UPDATE", "1")

	// Write original.
	original := []byte("original content")
	if err := os.WriteFile(filepath.Join(dir, "overwrite.txt"), original, 0o644); err != nil {
		t.Fatal(err)
	}

	updated := []byte("updated content")
	ok := fixture.Golden(t, "overwrite.txt", updated, fixture.WithRoot(dir))
	if !ok {
		t.Fatal("Golden returned false on update with GLACIER_GOLDEN_UPDATE=1")
	}

	written, err := os.ReadFile(filepath.Join(dir, "overwrite.txt"))
	if err != nil {
		t.Fatalf("read after update: %v", err)
	}
	if string(written) != string(updated) {
		t.Fatalf("file not updated: got %q, want %q", written, updated)
	}
}

// TestEnvVarRenameToGlacier: GLACIER_GOLDEN_UPDATE=1 activates update mode.
// MONGOOSE_GOLDEN_UPDATE must not exist. (#66 in test matrix)
func TestEnvVarRenameToGlacier(t *testing.T) {
	dir := t.TempDir()
	// Only the GLACIER_ var should work.
	t.Setenv("GLACIER_GOLDEN_UPDATE", "1")
	os.Unsetenv("MONGOOSE_GOLDEN_UPDATE") // ensure old name doesn't exist

	got := []byte("glacier brand")
	ok := fixture.Golden(t, "glacier_brand.txt", got, fixture.WithRoot(dir))
	if !ok {
		t.Fatal("Golden returned false with GLACIER_GOLDEN_UPDATE=1")
	}
	// Verify MONGOOSE variant is not recognized (set it, unset GLACIER, expect error).
	t.Setenv("GLACIER_GOLDEN_UPDATE", "0")
	t.Setenv("MONGOOSE_GOLDEN_UPDATE", "1")
	m := newMockTB()
	ok2 := fixture.Golden(m, "mongoose_brand.txt", got, fixture.WithRoot(dir))
	if ok2 {
		t.Fatal("Golden must not accept MONGOOSE_GOLDEN_UPDATE=1; old name must not work")
	}
}

// TestGoldenBinaryMismatchHexDiff: Binary mismatch shows hex header, not line diff.
func TestGoldenBinaryMismatchHexDiff(t *testing.T) {
	dir := t.TempDir()
	want := []byte{0x00, 0x01, 0x02, 0xFE, 0xFF}
	got := []byte{0x00, 0x01, 0x02, 0xAA, 0xBB}
	if err := os.WriteFile(filepath.Join(dir, "binary.bin"), want, 0o644); err != nil {
		t.Fatal(err)
	}
	m := newMockTB()
	ok := fixture.Golden(m, "binary.bin", got, fixture.WithRoot(dir))
	if ok {
		t.Fatal("Golden returned true on binary mismatch")
	}
	if !m.containsError("binary mismatch") {
		t.Fatalf("expected 'binary mismatch' in error; got: %v", m.allErrors())
	}
}

// TestGoldenEmptyGot: Empty got slice is distinct from missing file. (EX5)
func TestGoldenEmptyGot(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "empty.txt"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	ok := fixture.Golden(t, "empty.txt", []byte{}, fixture.WithRoot(dir))
	if !ok {
		t.Fatal("Golden returned false on empty match")
	}
}

// Table-driven test for Golden with various content types.
func TestGoldenTableDriven(t *testing.T) {
	cases := []struct {
		name    string
		want    []byte
		got     []byte
		matches bool
	}{
		{"text_match", []byte("foo\nbar\n"), []byte("foo\nbar\n"), true},
		{"text_mismatch", []byte("foo\n"), []byte("bar\n"), false},
		{"empty_match", []byte{}, []byte{}, true},
		{"binary_match", []byte{0x01, 0x02, 0x03}, []byte{0x01, 0x02, 0x03}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "file"), tc.want, 0o644); err != nil {
				t.Fatal(err)
			}
			m := newMockTB()
			ok := fixture.Golden(m, "file", tc.got, fixture.WithRoot(dir))
			if ok != tc.matches {
				t.Fatalf("Golden(%q) = %v, want %v; errors: %v", tc.name, ok, tc.matches, m.allErrors())
			}
		})
	}
}
