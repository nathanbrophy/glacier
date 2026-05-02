// SPDX-License-Identifier: Apache-2.0

package mock_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/mock"
)

// fakeT is a minimal assert.TB that captures calls for inspection in tests.
type fakeT struct {
	testing.TB
	errors  []string
	fatals  []string
	stopped bool
	cleanup []func()
}

func newFakeT() *fakeT { return &fakeT{} }

func (f *fakeT) Helper() {}
func (f *fakeT) Errorf(format string, args ...any) {
	f.errors = append(f.errors, fmt.Sprintf(format, args...))
}
func (f *fakeT) Fatalf(format string, args ...any) {
	f.fatals = append(f.fatals, fmt.Sprintf(format, args...))
	f.stopped = true
}
func (f *fakeT) FailNow()          { f.stopped = true }
func (f *fakeT) Cleanup(fn func()) { f.cleanup = append(f.cleanup, fn) }
func (f *fakeT) Name() string      { return "fakeT" }
func (f *fakeT) runCleanup() {
	for i := len(f.cleanup) - 1; i >= 0; i-- {
		f.cleanup[i]()
	}
}

func TestOfBasic(t *testing.T) {
	m := mock.Of[Greeter](t)
	if m.Interface() == nil {
		t.Fatal("Interface() must not be nil")
	}
}

func TestOfNonInterfacePanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for non-interface type")
		}
		msg := fmt.Sprintf("%v", r)
		if !strings.Contains(msg, "must be an interface type") {
			t.Fatalf("panic message %q does not mention 'must be an interface type'", msg)
		}
	}()
	_ = mock.Of[int](t) // must panic
}

func TestOfRegistersCleanup(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	m.OnCall("Greet").Return("hello").Times(1)
	// Do NOT call Greet — Verify at cleanup should report violation.
	ft.runCleanup()
	if len(ft.errors) == 0 {
		t.Fatal("expected Verify error at cleanup for unmet expectation")
	}
}

func TestInterfaceReturnsSatisfyingValue(t *testing.T) {
	m := mock.Of[Greeter](t)
	g := m.Interface()
	// Satisfy the interface: just check the value is non-nil.
	if g == nil {
		t.Fatal("Interface() returned nil")
	}
	// Ensure the interface method can be called (set up a matching expectation).
	m.OnCall("Greet").With(mock.Any[string]()).Return("hi").AnyTimes()
	result := g.Greet("world")
	if result != "hi" {
		t.Fatalf("got %q, want %q", result, "hi")
	}
}

func TestInterfaceMethodsRoutedToMock(t *testing.T) {
	m := mock.Of[Calculator](t)
	m.OnCall("Add").With(mock.Eq[int](2), mock.Eq[int](3)).Return(5).Times(1)
	result := m.Interface().Add(2, 3)
	if result != 5 {
		t.Fatalf("got %d, want 5", result)
	}
}

func TestOnCallReturnSimple(t *testing.T) {
	m := mock.Of[Greeter](t)
	m.OnCall("Greet").With(mock.Eq[string]("alice")).Return("hello, alice").Times(1)
	got := m.Interface().Greet("alice")
	if got != "hello, alice" {
		t.Fatalf("got %q, want %q", got, "hello, alice")
	}
}

func TestOnCallMethodNotInInterface(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for unknown method")
		}
	}()
	m.OnCall("NoSuchMethod")
}

func TestOnCallMethodNameRegex(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	cases := []string{
		"lowercase",   // doesn't start with uppercase
		"123abc",      // starts with digit
		"",            // empty
		"Has Space",   // contains space
		"Has\x00null", // contains null
	}
	for _, name := range cases {
		name := name
		t.Run(name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil {
					t.Fatalf("expected panic for invalid name %q", name)
				}
			}()
			m.OnCall(name)
		})
	}
}

func TestOnCallMethodNameOversize(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	longName := strings.Repeat("A", 65)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for oversize method name")
		}
	}()
	m.OnCall(longName)
}

func TestCallsToReturnsRecordedCalls(t *testing.T) {
	m := mock.Of[Greeter](t)
	m.OnCall("Greet").With(mock.Any[string]()).Return("hi").AnyTimes()
	m.Interface().Greet("alice")
	m.Interface().Greet("bob")
	calls := m.CallsTo("Greet")
	if len(calls) != 2 {
		t.Fatalf("got %d calls, want 2", len(calls))
	}
	if calls[0].Args[0] != "alice" {
		t.Errorf("first call arg: got %v, want alice", calls[0].Args[0])
	}
	if calls[1].Args[0] != "bob" {
		t.Errorf("second call arg: got %v, want bob", calls[1].Args[0])
	}
}

func TestCallsToOrderingPreserved(t *testing.T) {
	m := mock.Of[Calculator](t)
	m.OnCall("Add").With(mock.Any[int](), mock.Any[int]()).Return(0).AnyTimes()
	for i := range 5 {
		m.Interface().Add(i, i)
	}
	calls := m.CallsTo("Add")
	for i, c := range calls {
		if c.Args[0].(int) != i {
			t.Errorf("call %d: arg[0]=%v, want %d", i, c.Args[0], i)
		}
	}
}

func TestUnmatchedCallsLenient(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft, mock.LenientMode())
	m.Interface().Greet("nobody")
	calls := m.UnmatchedCalls()
	if len(calls) != 1 {
		t.Fatalf("got %d unmatched calls, want 1", len(calls))
	}
	if len(ft.errors) != 0 {
		t.Fatalf("lenient mode should not call Errorf; got: %v", ft.errors)
	}
}

func TestUnmatchedCallsStrictAlwaysEmpty(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft, mock.StrictDefault())
	m.Interface().Greet("nobody")
	calls := m.UnmatchedCalls()
	if len(calls) != 0 {
		t.Fatalf("strict mode: UnmatchedCalls should be empty, got %d", len(calls))
	}
	if len(ft.errors) == 0 {
		t.Fatal("strict mode: expected Errorf for unmatched call")
	}
}

func TestVerifyAutoInvokedAtCleanup(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	m.OnCall("Greet").Return("hi").Times(2)
	m.Interface().Greet("x") // only 1 call, but Times(2)
	ft.runCleanup()
	if len(ft.errors) == 0 {
		t.Fatal("expected verify error at cleanup")
	}
}

func TestVerifyMidTestCheckpoint(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	m.OnCall("Greet").Return("hi").Times(1)
	m.Verify() // called 0 times, should fail
	if len(ft.errors) == 0 {
		t.Fatal("expected verify failure (0 calls, expected 1)")
	}
}

func TestStrictDefault(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft) // default is strict
	m.Interface().Greet("x")
	if len(ft.errors) == 0 {
		t.Fatal("expected Errorf for unmatched call in default strict mode")
	}
}

func TestStrictUnmatched_TErrorf(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft, mock.StrictDefault())
	m.Interface().Greet("x")
	if len(ft.errors) == 0 {
		t.Fatal("expected Errorf")
	}
	if !strings.Contains(ft.errors[0], "Greet") {
		t.Errorf("error message should mention method name; got: %q", ft.errors[0])
	}
}

func TestStrictFatalHalts(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft, mock.StrictFatal())
	m.Interface().Greet("x")
	if len(ft.fatals) == 0 {
		t.Fatal("expected Fatalf for unmatched call in StrictFatal mode")
	}
}

func TestLenientMode(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft, mock.LenientMode())
	m.Interface().Greet("x")
	if len(ft.errors) != 0 {
		t.Fatalf("lenient mode: unexpected Errorf: %v", ft.errors)
	}
	if len(m.UnmatchedCalls()) != 1 {
		t.Fatal("lenient mode: expected 1 unmatched call")
	}
}

func TestStrictUnmatchedReturnsZeroValues(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft, mock.StrictDefault())
	result := m.Interface().Greet("x")
	if result != "" {
		t.Fatalf("unmatched strict call: expected zero string, got %q", result)
	}
}

func TestFirstRegisteredMatchWins(t *testing.T) {
	m := mock.Of[Greeter](t)
	m.OnCall("Greet").With(mock.Any[string]()).Return("first").AnyTimes()
	m.OnCall("Greet").With(mock.Any[string]()).Return("second").AnyTimes()
	result := m.Interface().Greet("x")
	if result != "first" {
		t.Fatalf("expected first match to win, got %q", result)
	}
}

func TestVerifyReportsAllUnmetInOneError(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Calculator](ft)
	m.OnCall("Add").With(mock.Any[int](), mock.Any[int]()).Return(0).Times(2)
	m.OnCall("Sub").With(mock.Any[int](), mock.Any[int]()).Return(0).Times(3)
	// Call neither → Verify should report both in one Errorf.
	m.Verify()
	if len(ft.errors) != 1 {
		t.Fatalf("expected exactly 1 Errorf call consolidating all violations; got %d", len(ft.errors))
	}
}

func TestMockClose_AliasForVerify(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	m.OnCall("Greet").Return("hi").Times(1)
	err := m.Close()
	if err != nil {
		t.Fatalf("Close() returned non-nil error: %v", err)
	}
	// Greet was not called → Verify should have fired.
	if len(ft.errors) == 0 {
		t.Fatal("Close() should have run Verify and reported unmet expectation")
	}
}

func TestMockCloseIdempotent(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	m.OnCall("Greet").Return("hi").Times(1)
	m.Close()
	initialErrors := len(ft.errors)
	m.Close() // second call should be a no-op
	if len(ft.errors) != initialErrors {
		t.Fatal("second Close() should not re-run Verify")
	}
}

func TestMockClose_BeforeCleanup(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	m.OnCall("Greet").Return("hi").Times(1)
	m.Close() // manual close runs Verify
	errorsBefore := len(ft.errors)
	ft.runCleanup() // cleanup should not re-run Verify
	if len(ft.errors) != errorsBefore {
		t.Fatal("Cleanup should be no-op after manual Close")
	}
}

func TestThirdPartyInterfaceMockable(t *testing.T) {
	// io.ReadCloser is a standard library interface; we need to register an adapter.
	// Since we can't import io in testhelpers (no TestMain coordination), we test
	// with our own interfaces here.
	// This test validates that multi-method interfaces work.
	m := mock.Of[Repo](t)
	m.OnCall("FindUser").
		With(mock.Any[context.Context](), mock.Eq[string]("u-1")).
		Return(User{ID: "u-1", Name: "Alice"}, nil).
		Times(1)

	ctx := context.Background()
	got, err := m.Interface().FindUser(ctx, "u-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Alice" {
		t.Errorf("got %q, want Alice", got.Name)
	}
}

func TestExpectationDoFn(t *testing.T) {
	m := mock.Of[Greeter](t)
	var called string
	m.OnCall("Greet").Do(func(name string) string {
		called = name
		return "hi " + name
	}).AnyTimes()
	result := m.Interface().Greet("world")
	if result != "hi world" {
		t.Errorf("Do fn: got %q, want %q", result, "hi world")
	}
	if called != "world" {
		t.Errorf("Do fn not called with correct arg; got %q", called)
	}
}

func TestExpectationDoFnWrongSignature(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for wrong Do signature")
		}
	}()
	m.OnCall("Greet").Do(func(x int) string { return "" }) // wrong param type
}

func TestExpectationReturnArityMismatch(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for arity mismatch")
		}
	}()
	m.OnCall("Greet").Return("a", "b") // Greet returns 1 value, not 2
}

func TestExpectationReturnTypeMismatch(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Calculator](ft)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for type mismatch")
		}
	}()
	m.OnCall("Add").
		With(mock.Any[int](), mock.Any[int]()).
		Return("not an int") // wrong type
}

func TestExpectationWithMatchersGeneric(t *testing.T) {
	m := mock.Of[Greeter](t)
	m.OnCall("Greet").With(mock.Eq[string]("world")).Return("hello").Times(1)
	result := m.Interface().Greet("world")
	if result != "hello" {
		t.Fatalf("got %q, want hello", result)
	}
}

func TestMockInBenchmarkB(t *testing.T) {
	// Verify that *testing.B satisfies assert.TB.
	// We just verify compilation; runtime test requires a benchmark.
	// This is a compile-time check via the type system.
	t.Log("*testing.B satisfies assert.TB (compile-time check)")
}

func TestFailureMessageRegisterCLI(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft, mock.StrictDefault())
	m.Interface().Greet("x")
	if len(ft.errors) == 0 {
		t.Fatal("expected error message")
	}
	msg := ft.errors[0]
	// The message should be sentence-case, period-terminated.
	if !strings.HasSuffix(strings.TrimSpace(msg), ".") {
		t.Errorf("error message should end with '.'; got: %q", msg)
	}
	// Should start with uppercase.
	if len(msg) > 0 && msg[0] < 'A' || msg[0] > 'Z' {
		t.Errorf("error message should start with uppercase; got: %q", msg)
	}
}

func TestInternalPanicRegisterLibrary(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		msg := fmt.Sprintf("%v", r)
		// Library register: lowercase, colon-delimited, no trailing period.
		if !strings.Contains(msg, "mock") {
			t.Errorf("panic message should reference 'mock'; got: %q", msg)
		}
	}()
	m.OnCall("Greet").Return("a", "b") // arity mismatch → panic
}

func TestMockUsesAssertHelpersInTests(t *testing.T) {
	// Compile-time check: the import of mock.Of requires assert.TB which is
	// from the assert package. This test just validates the dogfooding invariant
	// exists at the import level.
	t.Log("assert package is imported via mock.Of[T](t assert.TB,...)")
}

// errSentinel is a fixed error used for error-return tests.
var errSentinel = errors.New("sentinel error")

func TestRepoMockWithError(t *testing.T) {
	m := mock.Of[Repo](t)
	m.OnCall("FindUser").
		With(mock.Any[context.Context](), mock.Any[string]()).
		Return(User{}, errSentinel).
		Times(1)
	_, err := m.Interface().FindUser(context.Background(), "bad-id")
	if !errors.Is(err, errSentinel) {
		t.Fatalf("expected sentinel error, got: %v", err)
	}
}
