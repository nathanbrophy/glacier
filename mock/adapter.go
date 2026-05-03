// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"fmt"
	"reflect"
	"sync"
)

// adapterFactory is the concrete function type stored in adapterRegistry.
// It takes a dispatch function and returns a value satisfying T (as any).
// The dispatch function maps method name + reflect.Value args → reflect.Value returns.
type adapterFactory func(dispatch func(method string, args []reflect.Value) []reflect.Value) any

// adapterRegistry maps reflect.Type (interface type) → adapterFactory.
// Entries are written once (at RegisterAdapter call time) and read-only thereafter.
var adapterRegistry sync.Map

// RegisterAdapter registers a factory for the interface type T. The factory
// receives a dispatch function and must return a value that satisfies T. All
// method calls on the returned value must route through dispatch.
//
// RegisterAdapter is typically called once per interface type in TestMain or
// an init() function. The glaciergen code generator emits RegisterAdapter calls
// automatically when it processes a //+glacier:mock-annotated interface.
//
// Panics if T is not an interface type or if a factory is already registered
// for T (first-registration-wins is not the contract; double-registration is
// a programmer error).
//
// Example (user-written or glaciergen-generated):
//
//	func init() {
//	    mock.RegisterAdapter[Repo](func(dispatch func(string, []reflect.Value) []reflect.Value) any {
//	        return &repoMockAdapter{dispatch: dispatch}
//	    })
//	}
func RegisterAdapter[T any](factory func(dispatch func(string, []reflect.Value) []reflect.Value) T) {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	if ifaceType.Kind() != reflect.Interface {
		//glacier:nolint=panic-in-library programmer error: non-interface T surfaces at TestMain registration.
		panic(fmt.Sprintf("mock.RegisterAdapter: T must be an interface type, got %s", ifaceType.Kind()))
	}
	wrapped := func(dispatch func(string, []reflect.Value) []reflect.Value) any {
		return factory(dispatch)
	}
	if _, loaded := adapterRegistry.LoadOrStore(ifaceType, adapterFactory(wrapped)); loaded {
		//glacier:nolint=panic-in-library programmer error: duplicate registration surfaces at TestMain.
		panic(fmt.Sprintf("mock.RegisterAdapter: adapter already registered for %v", ifaceType))
	}
}

// getAdapterFactory retrieves the registered factory for interface type T.
// Returns (factory, true) if found, (nil, false) otherwise.
func getAdapterFactory(ifaceType reflect.Type) (adapterFactory, bool) {
	v, ok := adapterRegistry.Load(ifaceType)
	if !ok {
		return nil, false
	}
	return v.(adapterFactory), true
}
