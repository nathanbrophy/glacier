// SPDX-License-Identifier: Apache-2.0

package fixture_test

import (
	"testing"

	"github.com/nathanbrophy/glacier/fixture"
)

// TestPathTraversalRejectedGolden: Golden refuses ".." paths. (#19)
func TestPathTraversalRejectedGolden(t *testing.T) {
	cases := []string{
		"../oops",
		"a/../../b",
		"../secret.txt",
	}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			m := newMockTB()
			ok := fixture.Golden(m, name, []byte("data"))
			if ok {
				t.Fatalf("Golden(%q) returned true; expected rejection", name)
			}
			if !m.Failed() {
				t.Fatal("mockTB not marked failed on path traversal")
			}
		})
	}
}

// TestPathTraversalRejectedLoad: Load refuses ".." paths. (#21)
func TestPathTraversalRejectedLoad(t *testing.T) {
	m := newMockTB()
	panicked := callAndRecover(func() {
		fixture.Load(m, "../etc/passwd")
	})
	if !m.Failed() && !panicked {
		t.Fatal("Load did not report failure on path traversal")
	}
}

// TestPathTraversalRejectedWithRoot: WithRoot refuses ".." root paths. (#22)
func TestPathTraversalRejectedWithRoot(t *testing.T) {
	m := newMockTB()
	ok := fixture.Golden(m, "file.txt", []byte("data"), fixture.WithRoot("../../etc"))
	if ok {
		t.Fatal("Golden with traversal WithRoot returned true; expected rejection")
	}
	if !m.Failed() {
		t.Fatal("mockTB not marked failed on traversal WithRoot")
	}
}

// TestAbsolutePathRejected: Golden refuses absolute paths. (#23)
func TestAbsolutePathRejected(t *testing.T) {
	m := newMockTB()
	// Try an absolute path.
	ok := fixture.Golden(m, "/etc/passwd", []byte("data"))
	if ok {
		t.Fatal("Golden with absolute path returned true; expected rejection")
	}
	if !m.Failed() {
		t.Fatal("mockTB not marked failed on absolute path")
	}
}

// TestPathTraversalRejectedSnapshot: Snapshot refuses ".." paths. (#20)
func TestPathTraversalRejectedSnapshot(t *testing.T) {
	m := newMockTB()
	ok := fixture.Snapshot(m, "../oops", struct{ V int }{V: 1})
	if ok {
		t.Fatal("Snapshot with traversal path returned true; expected rejection")
	}
	if !m.Failed() {
		t.Fatal("mockTB not marked failed on traversal snapshot")
	}
}

// TestSnapshotPathRejectsTraversal: #69 in matrix.
func TestSnapshotPathRejectsTraversal(t *testing.T) {
	names := []string{"../oops", "../../secret"}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			m := newMockTB()
			ok := fixture.Snapshot(m, name, 42)
			if ok {
				t.Fatalf("Snapshot(%q) returned true; expected rejection", name)
			}
			if !m.Failed() {
				t.Fatal("mockTB not marked failed")
			}
		})
	}
}
