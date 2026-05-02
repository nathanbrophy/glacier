// SPDX-License-Identifier: Apache-2.0

package fixture

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/nathanbrophy/glacier/assert"
)

// leakConfig holds the resolved configuration for GuardLeaks.
type leakConfig struct {
	// invariant: at least one watch* field is true, or all are false (no-op).
	watchTempDirs   bool
	watchGoroutines bool
	watchEnv        bool
	watchFDs        bool
	// invariant: strict == false means t.Errorf; strict == true means t.Fatalf.
	strict bool
	// invariant: drainTimeout >= 0; zero means "use default (100 ms)".
	drainTimeout time.Duration
}

// LeakOption configures GuardLeaks behaviour.
type LeakOption interface{ applyLeak(*leakConfig) error }

type leakOptFunc func(*leakConfig) error

func (f leakOptFunc) applyLeak(c *leakConfig) error { return f(c) }

// WatchTempDirs enables monitoring of os.TempDir() for directories matching
// the "glacier-" prefix created during the test. Any such directory present
// after the test but not before is reported as a leak.
func WatchTempDirs() LeakOption {
	return leakOptFunc(func(c *leakConfig) error {
		c.watchTempDirs = true
		return nil
	})
}

// WatchGoroutines enables goroutine-count monitoring. Goroutines present after
// the test that were not present before (excluding well-known runtime
// goroutines: GC, finalizer, signal handler, runtime/cgo, test infrastructure)
// are reported with their stack traces.
func WatchGoroutines() LeakOption {
	return leakOptFunc(func(c *leakConfig) error {
		c.watchGoroutines = true
		return nil
	})
}

// WatchEnv enables environment-variable monitoring. Any env var added or
// changed during the test (after GuardLeaks is called) is reported as a leak.
// Env vars that existed before GuardLeaks was called are always ignored.
func WatchEnv() LeakOption {
	return leakOptFunc(func(c *leakConfig) error {
		c.watchEnv = true
		return nil
	})
}

// WatchFDs enables file-descriptor monitoring. On Linux and macOS, file
// descriptors open after the test but not before are reported. On Windows,
// WatchFDs is a no-op that emits a debug log message via fmt (slog.LevelDebug
// equivalent); no failure is reported.
func WatchFDs() LeakOption {
	return leakOptFunc(func(c *leakConfig) error {
		c.watchFDs = true
		return nil
	})
}

// WatchAll enables all four watchers: WatchTempDirs, WatchGoroutines,
// WatchEnv, and WatchFDs.
func WatchAll() LeakOption {
	return leakOptFunc(func(c *leakConfig) error {
		c.watchTempDirs = true
		c.watchGoroutines = true
		c.watchEnv = true
		c.watchFDs = true
		return nil
	})
}

// StrictLeaks causes GuardLeaks to call t.Fatalf instead of t.Errorf when a
// leak is detected. Default is t.Errorf (non-halt).
func StrictLeaks() LeakOption {
	return leakOptFunc(func(c *leakConfig) error {
		c.strict = true
		return nil
	})
}

// Millisecond is re-exported for use in WithDrainTimeout calls from example
// and test code within this package. It mirrors time.Millisecond.
const Millisecond = time.Millisecond

// WithDrainTimeout sets the window that WatchGoroutines waits for transient
// goroutines (e.g., goroutines finishing async work) to terminate before
// declaring a leak. Default is 100 ms. d must be positive.
func WithDrainTimeout(d time.Duration) LeakOption {
	return leakOptFunc(func(c *leakConfig) error {
		if d <= 0 {
			return fmt.Errorf("fixture: WithDrainTimeout: d must be positive, got %v", d)
		}
		c.drainTimeout = d
		return nil
	})
}

// applyLeakOptions applies opts to a zero leakConfig and returns it.
func applyLeakOptions(opts []LeakOption) (leakConfig, error) {
	var c leakConfig
	for _, o := range opts {
		if o != nil {
			if err := o.applyLeak(&c); err != nil {
				return leakConfig{}, err
			}
		}
	}
	return c, nil
}

// GuardLeaks records baseline state for each enabled watcher and registers a
// t.Cleanup that reports any new leaks detected after the test completes.
// With no options, GuardLeaks is a no-op (all watchers disabled by default;
// use WatchAll() or specific Watch* options). With StrictLeaks(), leaks call
// t.Fatalf instead of t.Errorf.
func GuardLeaks(t assert.TB, opts ...LeakOption) {
	t.Helper()
	cfg, err := applyLeakOptions(opts)
	if err != nil {
		t.Errorf("fixture: GuardLeaks: %v", err)
		return
	}

	// Baselines — recorded synchronously at call time.
	var (
		baseEnv      map[string]string
		baseTempDirs map[string]struct{}
		baseGoros    string
		baseFDCount  int
	)

	if cfg.watchEnv {
		baseEnv = snapshotEnv()
	}
	if cfg.watchTempDirs {
		baseTempDirs = snapshotTempDirs()
	}
	if cfg.watchGoroutines {
		baseGoros = snapshotGoroutines()
	}
	if cfg.watchFDs {
		baseFDCount = countFDs()
	}

	// report calls t.Errorf or t.Fatalf based on Strict mode.
	report := func(format string, args ...any) {
		t.Helper()
		if cfg.strict {
			t.Fatalf(format, args...)
		} else {
			t.Errorf(format, args...)
		}
	}

	drain := cfg.drainTimeout
	if drain <= 0 {
		drain = 100 * time.Millisecond
	}

	t.Cleanup(func() {
		if cfg.watchEnv {
			checkEnvLeaks(t, report, baseEnv)
		}
		if cfg.watchTempDirs {
			checkTempDirLeaks(t, report, baseTempDirs)
		}
		if cfg.watchGoroutines {
			checkGoroutineLeaks(t, report, baseGoros, drain)
		}
		if cfg.watchFDs {
			checkFDLeaks(t, report, baseFDCount)
		}
	})
}

// ── Environment watcher ───────────────────────────────────────────────────────

func snapshotEnv() map[string]string {
	env := os.Environ()
	m := make(map[string]string, len(env))
	for _, kv := range env {
		idx := strings.IndexByte(kv, '=')
		if idx < 0 {
			m[kv] = ""
		} else {
			m[kv[:idx]] = kv[idx+1:]
		}
	}
	return m
}

func checkEnvLeaks(t assert.TB, report func(string, ...any), base map[string]string) {
	t.Helper()
	current := snapshotEnv()
	var leaked []string
	for k, v := range current {
		if bv, ok := base[k]; !ok {
			leaked = append(leaked, fmt.Sprintf("added:   %s=%s", k, v))
		} else if bv != v {
			leaked = append(leaked, fmt.Sprintf("changed: %s=%s (was %s)", k, v, bv))
		}
	}
	if len(leaked) > 0 {
		sort.Strings(leaked)
		report("fixture: GuardLeaks: WatchEnv detected env leaks:\n  %s", strings.Join(leaked, "\n  "))
	}
}

// ── Temp-dir watcher ──────────────────────────────────────────────────────────

func snapshotTempDirs() map[string]struct{} {
	entries, err := os.ReadDir(os.TempDir())
	if err != nil {
		return nil
	}
	m := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "glacier-") {
			m[e.Name()] = struct{}{}
		}
	}
	return m
}

func checkTempDirLeaks(t assert.TB, report func(string, ...any), base map[string]struct{}) {
	t.Helper()
	current := snapshotTempDirs()
	var leaked []string
	for name := range current {
		if _, ok := base[name]; !ok {
			leaked = append(leaked, filepath.Join(os.TempDir(), name))
		}
	}
	if len(leaked) > 0 {
		sort.Strings(leaked)
		report("fixture: GuardLeaks: WatchTempDirs detected leaked temp dirs:\n  %s", strings.Join(leaked, "\n  "))
	}
}

// ── Goroutine watcher ─────────────────────────────────────────────────────────

const maxStackBuf = 1 << 20 // 1 MiB

func snapshotGoroutines() string {
	buf := make([]byte, maxStackBuf)
	n := runtime.Stack(buf, true)
	return string(buf[:n])
}

// goroutineNoisePrefixes are function name prefixes considered framework noise.
// Goroutines starting with any of these are filtered out as false positives.
var goroutineNoisePrefixes = []string{
	"runtime.goexit",
	"runtime.gcBgMarkWorker",
	"runtime.forcegchelper",
	"runtime.bgsweep",
	"runtime.bgscavenge",
	"runtime.runfinq",
	"runtime.timerproc",
	"signal.signal_recv",
	"os/signal",
	"runtime/cgo",
	"testing.(*M).Run",
	"testing.runTests",
	"testing.(*T).Run",
	"testing.tRunner",
	"testing.(*T).Cleanup",
	"net/http.(*Transport)",
	"net.(*Resolver)",
	"created by testing.(*T)",
	"created by testing.runTests",
	"created by runtime/trace",
	"created by net/http",
}

// isNoiseGoroutine returns true if the goroutine stack line belongs to a
// well-known runtime/test-infrastructure goroutine that should be filtered.
func isNoiseGoroutine(stack string) bool {
	for _, prefix := range goroutineNoisePrefixes {
		if strings.Contains(stack, prefix) {
			return true
		}
	}
	return false
}

// parseGoroutines splits a full stack dump into individual goroutine blocks.
func parseGoroutines(dump string) []string {
	var result []string
	// Goroutine blocks are separated by blank lines.
	blocks := strings.Split(strings.TrimSpace(dump), "\n\n")
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block != "" {
			result = append(result, block)
		}
	}
	return result
}

// goroutineID extracts the numeric goroutine ID from the first line of a
// goroutine block. Returns "" on failure.
func goroutineID(block string) string {
	// Format: "goroutine 42 [running]:"
	line, _, _ := strings.Cut(block, "\n")
	parts := strings.Fields(line)
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

func checkGoroutineLeaks(t assert.TB, report func(string, ...any), base string, drain time.Duration) {
	t.Helper()

	baseIDs := make(map[string]struct{})
	for _, block := range parseGoroutines(base) {
		if id := goroutineID(block); id != "" {
			baseIDs[id] = struct{}{}
		}
	}

	deadline := time.Now().Add(drain)
	var leaked []string
	for {
		current := snapshotGoroutines()
		leaked = leaked[:0]
		for _, block := range parseGoroutines(current) {
			id := goroutineID(block)
			if _, ok := baseIDs[id]; ok {
				continue // existed before the test
			}
			if isNoiseGoroutine(block) {
				continue // well-known false positive
			}
			leaked = append(leaked, block)
		}
		if len(leaked) == 0 || time.Now().After(deadline) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if len(leaked) > 0 {
		report("fixture: GuardLeaks: WatchGoroutines detected %d leaked goroutine(s):\n\n%s",
			len(leaked), strings.Join(leaked, "\n\n"))
	}
}
