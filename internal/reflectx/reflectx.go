// SPDX-License-Identifier: Apache-2.0

// Package reflectx provides reflection helpers shared by mock and other
// Glacier packages. It centralizes per-type reflection work (method enumeration,
// type caching) so that reflection overhead is paid once at construction rather
// than on every call.
package reflectx

import (
	"reflect"
	"sync"
)

// MethodInfo describes one method of an interface type.
type MethodInfo struct {
	Name  string
	Type  reflect.Type // the func type (no receiver)
	Index int          // position in reflect.Type.Method() enumeration
}

// InterfaceInfo is the cached descriptor for one interface type.
// Built once and stored in typeCache on first use.
type InterfaceInfo struct {
	// Type is the interface reflect.Type.
	Type reflect.Type
	// Methods is the ordered method list (by reflect.Type.Method index).
	Methods []MethodInfo
	// ReturnTypes maps method name → ordered list of reflect.Type for returns.
	ReturnTypes map[string][]reflect.Type
}

// typeCache caches InterfaceInfo by reflect.Type (interface type).
var typeCache sync.Map // map[reflect.Type]*InterfaceInfo

// GetOrBuild returns the cached InterfaceInfo for ifaceType, building and
// caching it if not yet present. Thread-safe; the build is idempotent.
func GetOrBuild(ifaceType reflect.Type) *InterfaceInfo {
	if v, ok := typeCache.Load(ifaceType); ok {
		return v.(*InterfaceInfo)
	}
	info := build(ifaceType)
	actual, _ := typeCache.LoadOrStore(ifaceType, info)
	return actual.(*InterfaceInfo)
}

// build constructs an InterfaceInfo for ifaceType.
func build(ifaceType reflect.Type) *InterfaceInfo {
	n := ifaceType.NumMethod()
	methods := make([]MethodInfo, n)
	retTypes := make(map[string][]reflect.Type, n)
	for i := range n {
		m := ifaceType.Method(i)
		methods[i] = MethodInfo{
			Name:  m.Name,
			Type:  m.Type,
			Index: i,
		}
		numOut := m.Type.NumOut()
		rt := make([]reflect.Type, numOut)
		for j := range numOut {
			rt[j] = m.Type.Out(j)
		}
		retTypes[m.Name] = rt
	}
	return &InterfaceInfo{
		Type:        ifaceType,
		Methods:     methods,
		ReturnTypes: retTypes,
	}
}

// ResetForTest clears the type cache. Intended for test isolation; callable
// from both test and production code (build constraints applied at call site).
func ResetForTest() {
	typeCache.Range(func(k, _ any) bool {
		typeCache.Delete(k)
		return true
	})
}
