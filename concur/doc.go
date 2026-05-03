// SPDX-License-Identifier: Apache-2.0

// Package concur is the Glacier answer to "I need a Mutex / errgroup /
// semaphore / sync.Pool / sync.Once / sync.WaitGroup :  but ctx-aware,
// observable, and friendly to Glacier's error story." Production builds are
// byte-equivalent to stdlib equivalents where possible (zero overhead is the
// rule, not an aspiration). Debug builds (-tags glacier_debug) add
// hold-timeout diagnostics and caller stack capture for catching deadlocks and
// held-too-long locks. Group adds ctx-aware Wait, optional concurrency cap,
// panic recovery, and TryGo non-blocking scheduling. Semaphore is
// atomic-counter-backed for high throughput. Pool, Once, and WaitGroup are
// generic or ctx-aware wrappers over their stdlib counterparts.
package concur
