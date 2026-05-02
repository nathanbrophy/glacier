// SPDX-License-Identifier: Apache-2.0

package mock_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/nathanbrophy/glacier/mock"
)

// --- Interfaces used across all test files ---

// Greeter is a simple single-method interface for basic tests.
type Greeter interface {
	Greet(name string) string
}

// greeterAdapter implements Greeter by routing calls through a dispatch function.
type greeterAdapter struct {
	dispatch func(string, []reflect.Value) []reflect.Value
}

func (a *greeterAdapter) Greet(name string) string {
	res := a.dispatch("Greet", []reflect.Value{reflect.ValueOf(name)})
	if len(res) == 0 || !res[0].IsValid() {
		return ""
	}
	return res[0].Interface().(string)
}

// Calculator is a multi-method interface.
type Calculator interface {
	Add(a, b int) int
	Sub(a, b int) int
}

// calculatorAdapter implements Calculator via dispatch.
type calculatorAdapter struct {
	dispatch func(string, []reflect.Value) []reflect.Value
}

func (a *calculatorAdapter) Add(x, y int) int {
	res := a.dispatch("Add", []reflect.Value{reflect.ValueOf(x), reflect.ValueOf(y)})
	if len(res) == 0 || !res[0].IsValid() {
		return 0
	}
	return int(res[0].Int())
}

func (a *calculatorAdapter) Sub(x, y int) int {
	res := a.dispatch("Sub", []reflect.Value{reflect.ValueOf(x), reflect.ValueOf(y)})
	if len(res) == 0 || !res[0].IsValid() {
		return 0
	}
	return int(res[0].Int())
}

// Repo is a multi-method interface that returns errors.
type Repo interface {
	FindUser(ctx context.Context, id string) (User, error)
	SaveUser(ctx context.Context, u User) error
}

// User is a simple value type used in Repo tests.
type User struct {
	ID   string
	Name string
}

// repoAdapter implements Repo via dispatch.
type repoAdapter struct {
	dispatch func(string, []reflect.Value) []reflect.Value
}

func (a *repoAdapter) FindUser(ctx context.Context, id string) (User, error) {
	res := a.dispatch("FindUser", []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(id)})
	var u User
	var err error
	if len(res) > 0 && res[0].IsValid() && !res[0].IsZero() {
		u = res[0].Interface().(User)
	}
	if len(res) > 1 && res[1].IsValid() && !res[1].IsNil() {
		err = res[1].Interface().(error)
	}
	return u, err
}

func (a *repoAdapter) SaveUser(ctx context.Context, u User) error {
	res := a.dispatch("SaveUser", []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(u)})
	if len(res) > 0 && res[0].IsValid() && !res[0].IsNil() {
		return res[0].Interface().(error)
	}
	return nil
}

// Queue has a variadic-like return (multiple returns of basic types).
type Queue interface {
	Pop() (int, bool)
}

// queueAdapter implements Queue via dispatch.
type queueAdapter struct {
	dispatch func(string, []reflect.Value) []reflect.Value
}

func (a *queueAdapter) Pop() (int, bool) {
	res := a.dispatch("Pop", nil)
	var v int
	var ok bool
	if len(res) > 0 && res[0].IsValid() {
		v = int(res[0].Int())
	}
	if len(res) > 1 && res[1].IsValid() {
		ok = res[1].Bool()
	}
	return v, ok
}

// Stringer wraps a simple interface for testing fmt.Stringer-like patterns.
type SimpleStringer interface {
	String() string
}

// stringerAdapter implements SimpleStringer via dispatch.
type stringerAdapter struct {
	dispatch func(string, []reflect.Value) []reflect.Value
}

func (a *stringerAdapter) String() string {
	res := a.dispatch("String", nil)
	if len(res) == 0 || !res[0].IsValid() {
		return ""
	}
	return res[0].Interface().(string)
}

// TestMain registers all adapters before tests run.
func TestMain(m *testing.M) {
	mock.RegisterAdapter[Greeter](func(dispatch func(string, []reflect.Value) []reflect.Value) Greeter {
		return &greeterAdapter{dispatch: dispatch}
	})
	mock.RegisterAdapter[Calculator](func(dispatch func(string, []reflect.Value) []reflect.Value) Calculator {
		return &calculatorAdapter{dispatch: dispatch}
	})
	mock.RegisterAdapter[Repo](func(dispatch func(string, []reflect.Value) []reflect.Value) Repo {
		return &repoAdapter{dispatch: dispatch}
	})
	mock.RegisterAdapter[Queue](func(dispatch func(string, []reflect.Value) []reflect.Value) Queue {
		return &queueAdapter{dispatch: dispatch}
	})
	mock.RegisterAdapter[SimpleStringer](func(dispatch func(string, []reflect.Value) []reflect.Value) SimpleStringer {
		return &stringerAdapter{dispatch: dispatch}
	})
	m.Run()
}
