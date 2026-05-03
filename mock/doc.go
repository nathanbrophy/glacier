// SPDX-License-Identifier: Apache-2.0

// Package mock provides interface mocking for tests in two complementary
// modes. Runtime reflect is the default: mock.Of[T any](t) produces a
// programmable mock for any interface T at runtime, with stringly-keyed
// expectations and generic typed matchers :  no codegen step, no workflow
// friction. Codegen is the opt-in path: a +glacier:mock marker on an interface
// causes glaciergen to emit a typed <Interface>Mock struct with method-named
// expectation builders, giving full IDE autocomplete and compile-time safety.
// Both paths share the same expectation engine, matchers, and return-value
// programming. Strict-by-default: any unexpected call fails the test loudly.
// No unsafe, no on-disk emission at runtime. Full API in specs/0012-mock.md.
package mock
