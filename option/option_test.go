// SPDX-License-Identifier: Apache-2.0

package option_test

import (
	"errors"
	"fmt"
	"go/token"
	"go/types"
	"regexp"
	"testing"

	"golang.org/x/tools/go/packages"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	"github.com/nathanbrophy/glacier/option"
)

// testConfig is a simple struct used as T across most tests.
type testConfig struct {
	a int
	b string
	c bool
}

// sentinel errors used across tests.
var (
	errA = errors.New("option: apply: errA")
	errB = errors.New("option: apply: errB")
	errC = errors.New("option: apply: errC")
)

// helpers

func withA(v int) option.Option[testConfig] {
	return option.OptionFunc[testConfig](func(c *testConfig) error {
		c.a = v
		return nil
	})
}

func withB(v string) option.Option[testConfig] {
	return option.OptionFunc[testConfig](func(c *testConfig) error {
		c.b = v
		return nil
	})
}

func withC(v bool) option.Option[testConfig] {
	return option.OptionFunc[testConfig](func(c *testConfig) error {
		c.c = v
		return nil
	})
}

func withErr(err error) option.Option[testConfig] {
	return option.OptionFunc[testConfig](func(_ *testConfig) error {
		return err
	})
}

// withMutateThenErr mutates first, then returns an error (E9).
func withMutateThenErr(v int, err error) option.Option[testConfig] {
	return option.OptionFunc[testConfig](func(c *testConfig) error {
		c.a = v
		return err
	})
}

// ---- T#1–T#3, T#8–T#10 TestApplySuccessCases ----
// Apply with various option slices that all succeed.

func TestApplySuccessCases(t *testing.T) {
	t.Parallel()
	type tc struct {
		name    string
		opts    []option.Option[testConfig]
		wantCfg testConfig
	}
	cases := []tc{
		{
			name:    "empty opts returns zero value",
			opts:    []option.Option[testConfig]{},
			wantCfg: testConfig{},
		},
		{
			name:    "single option applied",
			opts:    []option.Option[testConfig]{withA(42)},
			wantCfg: testConfig{a: 42},
		},
		{
			name:    "multiple options applied in order",
			opts:    []option.Option[testConfig]{withA(1), withB("hello"), withC(true)},
			wantCfg: testConfig{a: 1, b: "hello", c: true},
		},
		{
			name:    "nil option skipped, non-nil applied",
			opts:    []option.Option[testConfig]{nil, withA(3), nil},
			wantCfg: testConfig{a: 3},
		},
		{
			name:    "all nil options returns zero value",
			opts:    []option.Option[testConfig]{nil, nil, nil},
			wantCfg: testConfig{},
		},
		{
			name:    "duplicate field :  last wins",
			opts:    []option.Option[testConfig]{withA(1), withA(2), withA(3)},
			wantCfg: testConfig{a: 3},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, err := option.Apply(c.opts)
			require.NoError(t, err)
			assert.Equal(t, got, c.wantCfg)
		})
	}
}

// ---- T#4–T#5 TestApplyDefaultShortCircuit ----
// Apply in default (short-circuit) mode stops at the first error.

func TestApplyDefaultShortCircuit(t *testing.T) {
	t.Parallel()
	applied := 0
	counter := option.OptionFunc[testConfig](func(_ *testConfig) error {
		applied++
		return nil
	})
	// first option errors; counter is second :  should not run.
	_, err := option.Apply([]option.Option[testConfig]{withErr(errA), counter})
	require.Error(t, err)
	assert.ErrorIs(t, err, errA)
	assert.Equal(t, applied, 0)
}

// ---- T#5 TestApplyDefaultSecondErrors ----
// First option applies successfully; second errors and stops iteration.

func TestApplyDefaultSecondErrors(t *testing.T) {
	t.Parallel()
	thirdRan := false
	third := option.OptionFunc[testConfig](func(_ *testConfig) error {
		thirdRan = true
		return nil
	})
	got, err := option.Apply([]option.Option[testConfig]{withA(7), withErr(errB), third})
	require.Error(t, err)
	assert.ErrorIs(t, err, errB)
	assert.Equal(t, got.a, 7)
	assert.False(t, thirdRan)
}

// ---- T#6–T#7, T#11–T#12 TestApplyModes ----
// Apply mode selection: default short-circuits; Strict accumulates all errors.

func TestApplyModes(t *testing.T) {
	t.Parallel()
	type tc struct {
		name      string
		opts      []option.Option[testConfig]
		modes     []option.Mode
		wantErrA  bool
		wantErrB  bool
		wantNilEr bool
		wantA     int
	}
	cases := []tc{
		{
			name:     "strict :  accumulates all errors, applies successes",
			opts:     []option.Option[testConfig]{withErr(errA), withA(99), withErr(errB)},
			modes:    []option.Mode{option.Strict()},
			wantErrA: true,
			wantErrB: true,
			wantA:    99,
		},
		{
			name:      "strict :  no errors returns nil",
			opts:      []option.Option[testConfig]{withA(5), withB("ok"), withC(true)},
			modes:     []option.Mode{option.Strict()},
			wantNilEr: true,
			wantA:     5,
		},
		{
			name:     "multiple modes :  last wins (Mode{} then Strict → strict)",
			opts:     []option.Option[testConfig]{withErr(errA), withA(7), withErr(errB)},
			modes:    []option.Mode{option.Mode{}, option.Strict()},
			wantErrA: true,
			wantErrB: true,
			wantA:    7,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, err := option.Apply(c.opts, c.modes...)
			if c.wantNilEr {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				if c.wantErrA {
					assert.ErrorIs(t, err, errA)
				}
				if c.wantErrB {
					assert.ErrorIs(t, err, errB)
				}
			}
			if c.wantA != 0 {
				assert.Equal(t, got.a, c.wantA)
			}
		})
	}
}

// ---- T#12 TestApplyZeroModesIsDefault ----
// No mode option → default short-circuit: second option should not run.

func TestApplyZeroModesIsDefault(t *testing.T) {
	t.Parallel()
	secondRan := false
	second := option.OptionFunc[testConfig](func(_ *testConfig) error {
		secondRan = true
		return nil
	})
	_, err := option.Apply([]option.Option[testConfig]{withErr(errA), second})
	require.Error(t, err)
	assert.ErrorIs(t, err, errA)
	assert.False(t, secondRan)
}

// ---- T#13–T#14 TestApplyGenericT ----
// Apply works on non-struct T (primitive int, slice of string).

func TestApplyGenericT(t *testing.T) {
	t.Parallel()

	t.Run("primitive int T", func(t *testing.T) {
		t.Parallel()
		setTo42 := option.OptionFunc[int](func(n *int) error {
			*n = 42
			return nil
		})
		got, err := option.Apply[int]([]option.Option[int]{setTo42})
		require.NoError(t, err)
		assert.Equal(t, got, 42)
	})

	t.Run("slice of string T", func(t *testing.T) {
		t.Parallel()
		appendHello := option.OptionFunc[[]string](func(s *[]string) error {
			*s = append(*s, "hello")
			return nil
		})
		appendWorld := option.OptionFunc[[]string](func(s *[]string) error {
			*s = append(*s, "world")
			return nil
		})
		got, err := option.Apply[[]string]([]option.Option[[]string]{appendHello, appendWorld})
		require.NoError(t, err)
		assert.Len(t, got, 2)
		assert.Equal(t, got[0], "hello")
		assert.Equal(t, got[1], "world")
	})
}

// ---- T#15 TestApplyOptionPanicsPropagates ----

func TestApplyOptionPanicsPropagates(t *testing.T) {
	panicking := option.OptionFunc[testConfig](func(_ *testConfig) error {
		panic("deliberate panic in option")
	})
	defer func() {
		r := recover()
		if r == nil {
			require.True(t, false, "expected panic to propagate from Apply, but it did not")
		}
	}()
	//nolint:errcheck // panic expected :  return never reached.
	_, _ = option.Apply([]option.Option[testConfig]{panicking})
}

// ---- T#16 TestApplyOptionMutateThenError ----

func TestApplyOptionMutateThenError(t *testing.T) {
	t.Parallel()
	got, err := option.Apply([]option.Option[testConfig]{withMutateThenErr(99, errA)})
	require.Error(t, err)
	// Partial state is visible: a was set before the error was returned.
	var zero testConfig
	assert.NotEqual(t, got, zero)
	assert.Equal(t, got.a, 99)
}

// ---- T#17 TestOptionFuncSatisfiesOption ----

func TestOptionFuncSatisfiesOption(t *testing.T) {
	// Compile-time check: OptionFunc[testConfig] satisfies Option[testConfig].
	var _ option.Option[testConfig] = option.OptionFunc[testConfig](nil)
	// If we got here the assignment compiled; the test passes.
}

// ---- T#18 TestOptionFuncTypedNilApplyPanics ----

func TestOptionFuncTypedNilApplyPanics(t *testing.T) {
	// A typed nil OptionFunc panics when invoked through Apply because
	// the underlying func is nil. This is a caller error.
	nilFunc := option.OptionFunc[testConfig](nil)
	defer func() {
		r := recover()
		if r == nil {
			require.True(t, false, "expected panic from nil OptionFunc, but Apply did not panic")
		}
	}()
	//nolint:errcheck
	_, _ = option.Apply([]option.Option[testConfig]{nilFunc})
}

// ---- T#19 TestStrictReturnsStrictMode ----
// Verified indirectly: if Strict() mode is not active, Apply would short-circuit
// and we'd only see the first error. Getting both errors proves Strict().strict == true.

func TestStrictReturnsStrictMode(t *testing.T) {
	t.Parallel()
	_, err := option.Apply(
		[]option.Option[testConfig]{withErr(errA), withErr(errB)},
		option.Strict(),
	)
	assert.ErrorIs(t, err, errA)
	assert.ErrorIs(t, err, errB)
}

// ---- T#20–T#26 TestValidateCases ----
// Validate with zero, passing, failing, nil-target, and nil-validator scenarios.

func TestValidateCases(t *testing.T) {
	t.Parallel()
	alwaysPass := option.Validator[testConfig](func(_ *testConfig) error { return nil })
	failA := option.Validator[testConfig](func(_ *testConfig) error { return errA })
	failB := option.Validator[testConfig](func(_ *testConfig) error { return errB })

	type tc struct {
		name       string
		target     *testConfig
		validators []option.Validator[testConfig]
		wantNil    bool
		wantErrA   bool
		wantErrB   bool
		wantErrMsg string
	}
	cfg := testConfig{a: 1}
	cases := []tc{
		{
			name:       "no validators returns nil",
			target:     &cfg,
			validators: nil,
			wantNil:    true,
		},
		{
			name:       "all validators pass returns nil",
			target:     &cfg,
			validators: []option.Validator[testConfig]{alwaysPass, alwaysPass, alwaysPass},
			wantNil:    true,
		},
		{
			name:       "multiple failures joined",
			target:     &testConfig{},
			validators: []option.Validator[testConfig]{failA, alwaysPass, failB},
			wantErrA:   true,
			wantErrB:   true,
		},
		{
			name:       "nil target returns sentinel error",
			target:     nil,
			validators: nil,
			wantErrMsg: "option: validate: target is nil",
		},
		{
			name:       "nil validator skipped",
			target:     &cfg,
			validators: []option.Validator[testConfig]{nil, alwaysPass},
			wantNil:    true,
		},
		{
			name:       "all nil validators returns nil",
			target:     &testConfig{},
			validators: []option.Validator[testConfig]{nil, nil, nil},
			wantNil:    true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			err := option.Validate(c.target, c.validators...)
			if c.wantNil {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			if c.wantErrMsg != "" {
				assert.Equal(t, err.Error(), c.wantErrMsg)
			}
			if c.wantErrA {
				assert.ErrorIs(t, err, errA)
			}
			if c.wantErrB {
				assert.ErrorIs(t, err, errB)
			}
		})
	}
}

// ---- T#26 TestValidateValidatorPanicsPropagates ----

func TestValidateValidatorPanicsPropagates(t *testing.T) {
	cfg := testConfig{}
	panicking := option.Validator[testConfig](func(_ *testConfig) error {
		panic("deliberate panic in validator")
	})
	defer func() {
		r := recover()
		if r == nil {
			require.True(t, false, "expected panic to propagate from Validate, but it did not")
		}
	}()
	//nolint:errcheck
	_ = option.Validate(&cfg, panicking)
}

// ---- T#27–T#29, L-add-4 TestRequiredErrorMessages ----
// Required validator error message format across various field-name inputs.

func TestRequiredErrorMessages(t *testing.T) {
	t.Parallel()
	type tc struct {
		name      string
		fieldName string
		present   bool // whether getter returns non-nil
		wantErr   bool
		wantMsg   string
	}
	cases := []tc{
		{
			name:      "field present :  no error",
			fieldName: "val",
			present:   true,
			wantErr:   false,
		},
		{
			name:      "field absent :  error with quoted name",
			fieldName: "val",
			present:   false,
			wantErr:   true,
			wantMsg:   `option: required: field "val" not set`,
		},
		{
			name:      "field name with embedded quotes :  %q escaping",
			fieldName: `my "field"`,
			present:   false,
			wantErr:   true,
			wantMsg:   `option: required: field "my \"field\"" not set`,
		},
		{
			name:      "empty field name :  error with empty quotes",
			fieldName: "",
			present:   false,
			wantErr:   true,
			wantMsg:   `option: required: field "" not set`,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			type cfg struct{ v *int }
			var val = 42
			target := cfg{}
			if c.present {
				target.v = &val
			}
			vtor := option.Required[cfg](c.fieldName, func(cc *cfg) any {
				if cc.v == nil {
					return nil
				}
				return cc.v
			})
			err := vtor(&target)
			if c.wantErr {
				require.Error(t, err)
				assert.Equal(t, err.Error(), c.wantMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---- T#30 TestRequiredGenericTLoadBearing ----
// The §23.17 amendment: Required[T] getter is typed to *T, giving compile-time safety
// that the getter navigates the correct config type. Callers must return explicit nil
// for pointer fields (typed nil pointer in any is not == nil).

func TestRequiredGenericTLoadBearing(t *testing.T) {
	t.Parallel()
	type cfgA struct{ logger *int }
	type cfgB struct{ handler *string }
	valA := 1
	a := cfgA{logger: &valA}
	b := cfgB{handler: nil}

	// For pointer fields: return explicit nil when the pointer is nil.
	vtorA := option.Required[cfgA]("logger", func(c *cfgA) any {
		if c.logger == nil {
			return nil
		}
		return c.logger
	})
	vtorB := option.Required[cfgB]("handler", func(c *cfgB) any {
		if c.handler == nil {
			return nil
		}
		return c.handler
	})

	assert.NoError(t, vtorA(&a))
	err := vtorB(&b)
	require.Error(t, err)
	const want = `option: required: field "handler" not set`
	assert.Equal(t, err.Error(), want)
}

// ---- T#31 TestErrorRegisterConformanceOption ----
// Every error string produced by this package must match:
//
//	^option: [a-z][^A-Z.]*$

func TestErrorRegisterConformanceOption(t *testing.T) {
	re := regexp.MustCompile(`^option: [a-z][^A-Z.]*$`)

	type cfgV struct{ v *int }
	var (
		nilCfgV cfgV
		ptrInt  = new(int)
	)
	popCfgV := cfgV{v: ptrInt}
	_ = popCfgV

	// Collect all error strings produced by the package under all code paths.
	var allErrors []error

	// Validate: nil target.
	allErrors = append(allErrors, option.Validate[testConfig](nil))

	// Required: field not set. Return explicit nil for pointer fields.
	vtor := option.Required[cfgV]("v", func(c *cfgV) any {
		if c.v == nil {
			return nil
		}
		return c.v
	})
	allErrors = append(allErrors, vtor(&nilCfgV))

	// Apply: single error (from OptionFunc).
	_, err := option.Apply([]option.Option[testConfig]{withErr(errA)})
	// errA itself comes from the caller :  not from the package. Skip that.
	// We only check errors the package itself constructs.
	_ = err

	// Apply strict: joined errors also come from caller :  skip.

	for _, e := range allErrors {
		if e == nil {
			assert.True(t, false, "unexpected nil in error list (test setup bug)")
			continue
		}
		assert.True(t, re.MatchString(e.Error()), "error string %q does not match register pattern %s", e.Error(), re)
	}
}

// ---- T#48 TestSurfaceClosed_OptionPackage ----
// Verify the package exports exactly 8 symbols using go/types (via golang.org/x/tools/go/packages).

func TestSurfaceClosed_OptionPackage(t *testing.T) {
	fset := token.NewFileSet()
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedImports,
		Fset: fset,
	}
	pkgs, err := packages.Load(cfg, "github.com/nathanbrophy/glacier/option")
	require.NoError(t, err, "packages.Load failed")
	require.Len(t, pkgs, 1)
	pkg := pkgs[0]
	require.True(t, len(pkg.Errors) == 0, fmt.Sprintf("package load errors: %v", pkg.Errors))

	scope := pkg.Types.Scope()
	names := scope.Names()

	const wantCount = 8
	assert.Len(t, names, wantCount, "exported symbol count mismatch: %v", names)

	want := map[string]bool{
		"Option":     true,
		"OptionFunc": true,
		"Apply":      true,
		"Mode":       true,
		"Strict":     true,
		"Validator":  true,
		"Validate":   true,
		"Required":   true,
	}
	for _, name := range names {
		assert.True(t, want[name], "unexpected export %q", name)
	}
	for name := range want {
		obj := scope.Lookup(name)
		assert.NotNil(t, obj, "expected export %q not found", name)
		if obj != nil {
			assert.True(t, obj.Exported(), "%q is not exported", name)
		}
	}
	_ = types.Universe // ensure go/types is referenced
}
