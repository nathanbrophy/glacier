// SPDX-License-Identifier: Apache-2.0

package fixture_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/nathanbrophy/glacier/fixture"
)

// TestCaptureRoundTrip: Capture(t, fn) returns fn's stdout/stderr. (#46)
func TestCaptureRoundTrip(t *testing.T) {
	stdout, stderr := fixture.Capture(t, func() {
		fmt.Fprint(os.Stdout, "stdout output")
		fmt.Fprint(os.Stderr, "stderr output")
	})
	if stdout != "stdout output" {
		t.Fatalf("stdout = %q; want %q", stdout, "stdout output")
	}
	if stderr != "stderr output" {
		t.Fatalf("stderr = %q; want %q", stderr, "stderr output")
	}
}

// TestCaptureRestoresStreams: After Capture returns, os.Stdout/Stderr restored.
// (#47)
func TestCaptureRestoresStreams(t *testing.T) {
	origOut := os.Stdout
	origErr := os.Stderr

	fixture.Capture(t, func() {
		// Inside: streams redirected.
		if os.Stdout == origOut {
			t.Error("os.Stdout not redirected inside Capture")
		}
	})

	// Outside: streams restored.
	if os.Stdout != origOut {
		t.Fatal("os.Stdout not restored after Capture")
	}
	if os.Stderr != origErr {
		t.Fatal("os.Stderr not restored after Capture")
	}
}

// TestCapturePanicRestoresStreams: If fn panics, streams are still restored.
// (#50, EX2)
func TestCapturePanicRestoresStreams(t *testing.T) {
	origOut := os.Stdout
	origErr := os.Stderr

	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		fixture.Capture(t, func() {
			panic("intentional panic in capture")
		})
	}()

	if !panicked {
		t.Fatal("expected panic to propagate through Capture")
	}
	if os.Stdout != origOut {
		t.Fatal("os.Stdout not restored after panic in Capture")
	}
	if os.Stderr != origErr {
		t.Fatal("os.Stderr not restored after panic in Capture")
	}
}

// TestCaptureEmptyOutput: Capture of fn that writes nothing returns empty strings.
func TestCaptureEmptyOutput(t *testing.T) {
	stdout, stderr := fixture.Capture(t, func() {})
	if stdout != "" {
		t.Fatalf("stdout = %q; want empty", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q; want empty", stderr)
	}
}

// TestCaptureMultiLine: Capture returns multi-line output.
func TestCaptureMultiLine(t *testing.T) {
	stdout, _ := fixture.Capture(t, func() {
		fmt.Fprintln(os.Stdout, "line 1")
		fmt.Fprintln(os.Stdout, "line 2")
		fmt.Fprintln(os.Stdout, "line 3")
	})
	if stdout != "line 1\nline 2\nline 3\n" {
		t.Fatalf("stdout = %q; want multi-line", stdout)
	}
}

// TestCaptureFromGoroutine: Capture lock holds even when fn writes from
// goroutine. (#49)
func TestCaptureFromGoroutine(t *testing.T) {
	// NOTE: Do NOT call t.Parallel() in tests that use Capture.
	stdout, _ := fixture.Capture(t, func() {
		done := make(chan struct{})
		go func() {
			fmt.Fprint(os.Stdout, "goroutine output")
			close(done)
		}()
		<-done
	})
	if !containsStr(stdout, "goroutine output") {
		t.Fatalf("stdout %q does not contain 'goroutine output'", stdout)
	}
}

// TestCaptureProcessWideLockSerializes: Two serial Capture calls don't
// interleave (we can't do truly parallel Capture since it uses a mutex,
// so we verify the lock by checking sequential isolation). (#48)
func TestCaptureProcessWideLockSerializes(t *testing.T) {
	// NOTE: Do NOT call t.Parallel() in Capture tests.
	stdout1, _ := fixture.Capture(t, func() { fmt.Fprint(os.Stdout, "first") })
	stdout2, _ := fixture.Capture(t, func() { fmt.Fprint(os.Stdout, "second") })
	if stdout1 != "first" {
		t.Fatalf("first capture = %q; want 'first'", stdout1)
	}
	if stdout2 != "second" {
		t.Fatalf("second capture = %q; want 'second'", stdout2)
	}
}

func containsStr(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	if len(s) < len(sub) {
		return false
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
