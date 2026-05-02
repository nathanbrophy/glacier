// SPDX-License-Identifier: Apache-2.0

package mock_test

import (
	"context"
	"fmt"

	"github.com/nathanbrophy/glacier/mock"
)

// ExampleOf_basic demonstrates creating a runtime mock and verifying expectations.
func ExampleOf_basic() {
	// In a real test, use *testing.T. Here we use a no-op stand-in.
	t := &fakeT{}
	t.cleanup = nil

	m := mock.Of[Greeter](t)

	m.OnCall("Greet").
		With(mock.Eq[string]("world")).
		Return("hello, world").
		Times(1)

	g := m.Interface()
	fmt.Println(g.Greet("world"))
	// Output: hello, world
}

// ExampleEq_typed demonstrates compile-time-typed argument matching.
func ExampleEq_typed() {
	t := &fakeT{}

	m := mock.Of[Greeter](t)
	m.OnCall("Greet").
		With(mock.Eq[string]("alice")).
		Return("hi alice").
		AnyTimes()

	g := m.Interface()
	fmt.Println(g.Greet("alice"))
	// Output: hi alice
}

// ExampleExpectation_ReturnSeq demonstrates sequenced return values.
func ExampleExpectation_ReturnSeq() {
	t := &fakeT{}

	m := mock.Of[Queue](t)
	m.OnCall("Pop").
		ReturnSeq([][]any{
			{1, true},
			{2, true},
		}, mock.SeqExhaust).
		AnyTimes()

	q := m.Interface()
	v, ok := q.Pop()
	fmt.Println(v, ok)
	v, ok = q.Pop()
	fmt.Println(v, ok)
	// Output:
	// 1 true
	// 2 true
}

// ExampleOf_repo demonstrates a multi-method interface mock.
func ExampleOf_repo() {
	t := &fakeT{}

	m := mock.Of[Repo](t)
	m.OnCall("FindUser").
		With(mock.Any[context.Context](), mock.Eq[string]("u-42")).
		Return(User{ID: "u-42", Name: "Alice"}, nil).
		Times(1)

	repo := m.Interface()
	got, err := repo.FindUser(context.Background(), "u-42")
	fmt.Println(got.Name, err)
	// Output: Alice <nil>
}
