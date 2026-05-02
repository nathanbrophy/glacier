// SPDX-License-Identifier: Apache-2.0

package log_test

import (
	"bytes"
	"log/slog"
	"sync"
	"testing"

	"github.com/nathanbrophy/glacier/log"
)

// T#L27 Multiple goroutines logging concurrently through the same handler: no race.
func TestConcurrentLogging(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	h := log.NewHandler(&buf, log.WithColor(log.ColorNever))
	l := slog.New(h)

	const goroutines = 10
	const records = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			for range records {
				l.Info("concurrent log", "goroutine", "worker")
			}
		}()
	}
	wg.Wait()
}
