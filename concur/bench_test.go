// SPDX-License-Identifier: Apache-2.0

package concur_test

import (
	"context"
	"sync"
	"testing"

	"github.com/nathanbrophy/glacier/concur"
)

// BenchmarkMutexLockUnlock measures concur.Mutex against sync.Mutex baseline.
func BenchmarkMutexLockUnlock(b *testing.B) {
	b.ReportAllocs()
	var mu concur.Mutex
	for b.Loop() {
		mu.Lock()
		mu.Unlock()
	}
}

func BenchmarkSyncMutexLockUnlock(b *testing.B) {
	b.ReportAllocs()
	var mu sync.Mutex
	for b.Loop() {
		mu.Lock()
		mu.Unlock()
	}
}

// BenchmarkSemaphoreAcquireRelease — fast-path benchmark (should be <= 50 ns/op).
func BenchmarkSemaphoreAcquireRelease(b *testing.B) {
	b.ReportAllocs()
	s, _ := concur.NewSemaphore(1000)
	ctx := context.Background()
	for b.Loop() {
		_ = s.Acquire(ctx, 1)
		_ = s.Release(1)
	}
}
