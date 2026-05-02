// SPDX-License-Identifier: Apache-2.0

package mock_test

import (
	"testing"

	"github.com/nathanbrophy/glacier/mock"
)

func BenchmarkMockCallNoArgs(b *testing.B) {
	m := mock.Of[SimpleStringer](b)
	m.OnCall("String").Return("hello").AnyTimes()
	s := m.Interface()
	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_ = s.String()
	}
}

func BenchmarkMockCall3Args(b *testing.B) {
	m := mock.Of[Calculator](b)
	m.OnCall("Add").With(mock.Any[int](), mock.Any[int]()).Return(0).AnyTimes()
	c := m.Interface()
	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_ = c.Add(1, 2)
	}
}

func BenchmarkMockCallWithMatchers(b *testing.B) {
	m := mock.Of[Greeter](b)
	m.OnCall("Greet").
		With(mock.Eq[string]("world")).
		Return("hi").
		AnyTimes()
	g := m.Interface()
	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_ = g.Greet("world")
	}
}

func BenchmarkMockOf(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		m := mock.Of[Greeter](b)
		m.OnCall("Greet").Return("hi").AnyTimes()
		_ = m.Interface()
	}
}
