// SPDX-License-Identifier: Apache-2.0

package fixture_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nathanbrophy/glacier/fixture"
)

// BenchmarkGoldenBytes: Compare-and-pass path for 4 KiB goldens; zero allocs
// on match. (#70)
func BenchmarkGoldenBytes(b *testing.B) {
	dir := b.TempDir()
	content := make([]byte, 4*1024)
	for i := range content {
		content[i] = byte(i % 256)
	}
	if err := os.WriteFile(filepath.Join(dir, "bench.bin"), content, 0o644); err != nil {
		b.Fatal(err)
	}

	// Avoid re-allocating content inside the loop.
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		fixture.Golden(b, "bench.bin", content, fixture.WithRoot(dir))
	}
}

// BenchmarkSnapshotStruct: Snapshot of a moderately sized struct. (#71)
func BenchmarkSnapshotStruct(b *testing.B) {
	type BigStruct struct {
		A, B, C, D, E int
		F, G, H, I, J string
		K, L, M, N, O float64
		P, Q, R, S, T bool
		U             time.Time
	}
	v := BigStruct{
		A: 1, B: 2, C: 3, D: 4, E: 5,
		F: "f", G: "g", H: "h", I: "i", J: "j",
		K: 1.1, L: 2.2, M: 3.3, N: 4.4, O: 5.5,
		P: true, Q: false, R: true, S: false, T: true,
		U: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	// Write the snapshot to testdata/snapshots first.
	b.Setenv("GLACIER_GOLDEN_UPDATE", "1")
	fixture.Snapshot(b, "bench_big_struct", v)
	b.Setenv("GLACIER_GOLDEN_UPDATE", "0")

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		fixture.Snapshot(b, "bench_big_struct", v)
	}
}

// BenchmarkSnapshotMap100Keys: Deterministic-formatter perf with 100-key map.
// (#72)
func BenchmarkSnapshotMap100Keys(b *testing.B) {
	m := make(map[string]int, 100)
	for i := range 100 {
		m[fmt.Sprintf("key%03d", i)] = i
	}

	b.Setenv("GLACIER_GOLDEN_UPDATE", "1")
	fixture.Snapshot(b, "bench_map100", m)
	b.Setenv("GLACIER_GOLDEN_UPDATE", "0")

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		fixture.Snapshot(b, "bench_map100", m)
	}
}

// BenchmarkCaptureSmallOutput: Capture of small fn (lock + redirect). (#73)
func BenchmarkCaptureSmallOutput(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		fixture.Capture(b, func() {
			fmt.Fprint(os.Stdout, "x")
		})
	}
}
