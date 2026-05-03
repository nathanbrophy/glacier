// SPDX-License-Identifier: Apache-2.0

package conf

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// globalRegistry is the process-wide store for all Register calls.
var globalRegistry = &registry{}

// registry holds the set of named registrations.
type registry struct {
	mu   sync.Mutex
	regs map[string]*registration
}

// registration describes a single registered configuration section.
type registration struct {
	path     string
	defaults any
	store    func(any)  // atomically stores a new *T
	load     func() any // atomically loads the current *T as any
}

// Register declares a configuration section of type T at the given
// dot-separated path and returns a snapshot accessor. Before Load is called,
// the accessor returns a pointer to a copy of defaults.
//
// Panics if path has already been registered in this process.
//
// Typical usage at package level:
//
//	var serverCfg = conf.Register("server", ServerConfig{Port: 8080})
//
// After conf.Load, serverCfg() returns the fully populated *ServerConfig.
func Register[T any](path string, defaults T) func() *T {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	if globalRegistry.regs == nil {
		globalRegistry.regs = make(map[string]*registration)
	}
	if _, exists := globalRegistry.regs[path]; exists {
		//glacier:nolint=panic-in-library programmer error: duplicate Register paths surface at package init.
		panic(fmt.Sprintf("conf: register: path %q already registered", path))
	}

	// Seed the atomic pointer with a copy of defaults so the accessor is
	// immediately usable before any Load call.
	ptr := &atomic.Pointer[T]{}
	initial := new(T)
	*initial = defaults
	ptr.Store(initial)

	reg := &registration{
		path:     path,
		defaults: defaults,
		store: func(v any) {
			if typed, ok := v.(*T); ok {
				ptr.Store(typed)
			}
		},
		load: func() any {
			return ptr.Load()
		},
	}
	globalRegistry.regs[path] = reg

	return func() *T {
		return ptr.Load()
	}
}
