// SPDX-License-Identifier: Apache-2.0

package fixture_test

import (
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/nathanbrophy/glacier/fixture"
)

// ExampleGolden demonstrates golden-file comparison for an HTTP response body.
func ExampleGolden() {
	// (In real tests, t is the *testing.T from the test runner.)
	// Here we fabricate a *testing.T for the example.
	var t *testing.T // placeholder

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","version":"1.2.3"}`))
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	// GLACIER_GOLDEN_UPDATE=1 creates the file on first run.
	fixture.Golden(t, "health_response.json", rec.Body.Bytes())
}

// ExampleSnapshot demonstrates typed snapshot comparison.
func ExampleSnapshot() {
	var t *testing.T // placeholder

	type UserProfile struct {
		ID        int
		Name      string
		Email     string
		CreatedAt time.Time
	}
	got := UserProfile{
		ID:    42,
		Name:  "Alice",
		Email: "alice@example.com",
		// time.Time will be stored as date-only in the snapshot.
		CreatedAt: time.Now(),
	}
	// Snapshot persists to testdata/snapshots/user_profile.snap.
	fixture.Snapshot(t, "user_profile", got)
}

// ExampleNewFS demonstrates the in-memory filesystem.
func ExampleNewFS() {
	fsys := fixture.NewFS(map[string][]byte{
		"config.json":    []byte(`{"env":"test"}`),
		"data/input.txt": []byte("input data"),
	})

	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
	// Output: {"env":"test"}
}

// ExampleCapture demonstrates stdout/stderr capture.
func ExampleCapture() {
	var t *testing.T // placeholder

	printBanner := func() {
		fmt.Fprintln(os.Stdout, "Glacier v0.1.0")
		fmt.Fprintln(os.Stderr, "debug: banner rendered")
	}

	// NOTE: do not call t.Parallel() in tests that use Capture.
	stdout, stderr := fixture.Capture(t, printBanner)
	fmt.Println("captured stdout:", stdout)
	fmt.Println("captured stderr:", stderr)
}

// ExampleReal demonstrates the real clock.
func ExampleReal() {
	clk := fixture.Real()
	_ = clk.Now() // returns current wall time
}

// ExampleNewClock demonstrates the deterministic fake clock.
func ExampleNewClock() {
	var t *testing.T // placeholder
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	clk := fixture.NewClock(t, start)
	clk.Advance(5 * time.Second)
	// clk.Now() == start + 5s
}

// ExampleGuardLeaks demonstrates the lifecycle invariant guard.
func ExampleGuardLeaks() {
	var t *testing.T // placeholder

	fixture.GuardLeaks(t,
		fixture.WatchGoroutines(),
		fixture.WatchTempDirs(),
		fixture.WatchEnv(),
		fixture.StrictLeaks(),
		fixture.WithDrainTimeout(200*fixture.Millisecond),
	)
	// Any env vars, temp dirs, or goroutines added during the test but not
	// cleaned up will be reported at t.Cleanup time.
}
