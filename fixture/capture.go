// SPDX-License-Identifier: Apache-2.0

package fixture

import (
	"bytes"
	"io"
	"os"
	"sync"

	"github.com/nathanbrophy/glacier/assert"
)

// captureMu is the process-wide mutex that serializes Capture calls.
// os.Stdout and os.Stderr are process-global file descriptors; concurrent
// capture would interleave output. Tests using Capture must NOT call
// t.Parallel().
var captureMu sync.Mutex

// Capture calls fn while redirecting os.Stdout and os.Stderr to in-memory
// buffers, then returns the captured output as (stdout, stderr string).
// os.Stdout and os.Stderr are restored before Capture returns, even if fn
// panics. Capture holds a process-wide mutex for the duration of fn, so two
// concurrent Capture calls will serialize. Tests using Capture must not call
// t.Parallel.
func Capture(t assert.TB, fn func()) (stdout, stderr string) {
	t.Helper()
	captureMu.Lock()
	defer captureMu.Unlock()

	// Save originals.
	origOut := os.Stdout
	origErr := os.Stderr

	// Create pipes.
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Errorf("fixture: Capture: create stdout pipe: %v", err)
		return "", ""
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		wOut.Close()
		rOut.Close()
		t.Errorf("fixture: Capture: create stderr pipe: %v", err)
		return "", ""
	}

	// Redirect.
	os.Stdout = wOut
	os.Stderr = wErr

	// Restore on exit, including on panic.
	defer func() {
		os.Stdout = origOut
		os.Stderr = origErr
	}()

	// Run fn; recover any panic to ensure restore.
	var panicVal any
	var panicked bool
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicVal = r
				panicked = true
			}
		}()
		fn()
	}()

	// Close write ends to flush pipe buffers.
	wOut.Close()
	wErr.Close()

	// Read captured output.
	var outBuf, errBuf bytes.Buffer
	io.Copy(&outBuf, rOut) //nolint:errcheck // read from pipe; errors ignored after write end closed
	io.Copy(&errBuf, rErr) //nolint:errcheck
	rOut.Close()
	rErr.Close()

	if panicked {
		panic(panicVal)
	}

	return outBuf.String(), errBuf.String()
}
