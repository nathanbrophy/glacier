// SPDX-License-Identifier: Apache-2.0

// Package profile provides a one-call helper for enabling pprof CPU, heap,
// and goroutine profiling. It is wired by the --profile global flag.
package profile

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
)

// Start begins CPU profiling to <file>.cpu and returns a stop function.
// When stop is called it flushes the CPU profile and writes <file>.heap and
// <file>.goroutine. If file is empty, Start returns a no-op stop and a nil
// error.
//
// The returned stop function is idempotent and safe to call more than once.
func Start(file string) (stop func(), err error) {
	if file == "" {
		return func() {}, nil
	}

	cpuPath := file + ".cpu"
	cpuFile, err := os.Create(cpuPath)
	if err != nil {
		return func() {}, fmt.Errorf("profile: create cpu file: %w", err)
	}
	if err := pprof.StartCPUProfile(cpuFile); err != nil {
		_ = cpuFile.Close()
		return func() {}, fmt.Errorf("profile: start cpu profile: %w", err)
	}

	var once sync.Once
	stop = func() {
		once.Do(func() {
			pprof.StopCPUProfile()
			_ = cpuFile.Close()

			writeProfile("heap", file+".heap")
			writeGoroutines(file + ".goroutine")
		})
	}
	return stop, nil
}

// writeProfile writes the named pprof profile to path.
func writeProfile(name, path string) {
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()
	runtime.GC() // flush allocations before heap snapshot
	if p := pprof.Lookup(name); p != nil {
		_ = p.WriteTo(f, 0)
	}
}

// writeGoroutines writes a goroutine profile to path.
func writeGoroutines(path string) {
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()
	if p := pprof.Lookup("goroutine"); p != nil {
		_ = p.WriteTo(f, 1)
	}
}
