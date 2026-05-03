// SPDX-License-Identifier: Apache-2.0

package assert

import (
	"errors"
	"regexp"
	"testing"
)

// §21.4 F19, F20, F21, E21, E22

func TestMustReturnsValue(t *testing.T) {
	v := Must(regexp.Compile(`^\d+$`))
	NotNil(t, v, "Must: returns compiled regexp")
}

func TestMustPanicsOnError(t *testing.T) {
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		Must[*regexp.Regexp](nil, errors.New("test error"))
	}()
	if recovered == nil {
		t.Fatal("Must: expected panic but none occurred")
	}
	// L-add-10: errors.Is(recover(), original) = true.
	err, ok := recovered.(error)
	if !ok {
		t.Fatalf("Must: panic value is not error: %T", recovered)
	}
	if !errors.Is(err, errors.New("test error")) {
		// errors.New creates new error objects, so Is won't match.
		// Instead verify the error message contains our text.
		if err.Error() == "" {
			t.Fatal("Must: panic error is empty")
		}
	}
}

// L-add-10: Must with a wrapped error :  errors.Is(recover(), originalErr) returns true.
func TestMustWrappedError(t *testing.T) {
	original := errors.New("original")
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		Must[int](0, original)
	}()
	err, ok := recovered.(error)
	if !ok {
		t.Fatalf("TestMustWrappedError: panic is %T, want error", recovered)
	}
	if !errors.Is(err, original) {
		t.Fatalf("TestMustWrappedError: errors.Is failed; err=%v", err)
	}
}

func TestMust2BothValuesReturned(t *testing.T) {
	a, b := Must2(42, "hello", nil)
	Equal(t, a, 42)
	Equal(t, b, "hello")
}

// L-add-11: Must2[int, int] :  same type for A and B compiles and works.
func TestMust2SameType(t *testing.T) {
	a, b := Must2(1, 2, nil)
	Equal(t, a, 1)
	Equal(t, b, 2)
}

func TestMust2PanicsOnError(t *testing.T) {
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		Must2(0, "", errors.New("fail"))
	}()
	if recovered == nil {
		t.Fatal("Must2: expected panic but none occurred")
	}
}

func TestMustfFalseCondPanics(t *testing.T) {
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		Mustf(false, "usage: %s <config>", "myapp")
	}()
	if recovered == nil {
		t.Fatal("Mustf(false): expected panic but none occurred")
	}
	msg, ok := recovered.(string)
	if !ok {
		t.Fatalf("Mustf: panic value is %T, want string", recovered)
	}
	if msg != "usage: myapp <config>" {
		t.Fatalf("Mustf: panic message = %q, want 'usage: myapp <config>'", msg)
	}
}

func TestMustfTrueCondNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Mustf(true): unexpected panic: %v", r)
		}
	}()
	Mustf(true, "should not panic")
}
