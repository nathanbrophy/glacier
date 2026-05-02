// SPDX-License-Identifier: Apache-2.0

// Bootstrap discipline: stdlib testing + math/rand only; no assert/ or fixture/ packages.

package option_test

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/nathanbrophy/glacier/option"
)

// ---- T#41 PropertyApplyIdempotent ----
// For pure idempotent options, Apply(Apply(opts)) == Apply(opts).
// "Idempotent" here means applying the same options twice produces the same T.

func PropertyApplyIdempotent(t *testing.T) {
	t.Helper()
	type cfg struct{ x, y int }

	setX := func(v int) option.Option[cfg] {
		return option.OptionFunc[cfg](func(c *cfg) error { c.x = v; return nil })
	}
	setY := func(v int) option.Option[cfg] {
		return option.OptionFunc[cfg](func(c *cfg) error { c.y = v; return nil })
	}

	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 100; i++ {
		x := rng.Intn(1000)
		y := rng.Intn(1000)
		opts := []option.Option[cfg]{setX(x), setY(y)}
		got1, err1 := option.Apply(opts)
		if err1 != nil {
			t.Fatalf("iter %d: unexpected error on first Apply: %v", i, err1)
		}
		got2, err2 := option.Apply(opts)
		if err2 != nil {
			t.Fatalf("iter %d: unexpected error on second Apply: %v", i, err2)
		}
		if got1 != got2 {
			t.Errorf("iter %d: idempotency broken: first=%+v second=%+v", i, got1, got2)
		}
	}
}

func TestPropertyApplyIdempotent(t *testing.T) { PropertyApplyIdempotent(t) }

// ---- T#42 PropertyApplyNilSkipPermutation ----
// Inserting nils anywhere in opts produces same result as opts without nils.

func TestPropertyApplyNilSkipPermutation(t *testing.T) {
	type cfg struct{ v int }
	setV := func(v int) option.Option[cfg] {
		return option.OptionFunc[cfg](func(c *cfg) error { c.v = v; return nil })
	}

	base := []option.Option[cfg]{setV(1), setV(2), setV(3)}
	want, err := option.Apply(base)
	if err != nil {
		t.Fatalf("unexpected error on base Apply: %v", err)
	}

	rng := rand.New(rand.NewSource(7))
	for iter := 0; iter < 50; iter++ {
		// Build a permutation with random nil insertions.
		with := make([]option.Option[cfg], 0, len(base)*2)
		for _, o := range base {
			if rng.Intn(2) == 0 {
				with = append(with, nil)
			}
			with = append(with, o)
		}
		if rng.Intn(2) == 0 {
			with = append(with, nil)
		}

		got, err := option.Apply(with)
		if err != nil {
			t.Fatalf("iter %d: unexpected error: %v", iter, err)
		}
		if got != want {
			t.Errorf("iter %d: nil insertion changed result: want=%+v got=%+v", iter, want, got)
		}
	}
}

// ---- T#43 PropertyApplyLastWins ----
// Setting same field N times → final value == last setter.

func TestPropertyApplyLastWins(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	for iter := 0; iter < 100; iter++ {
		n := 1 + rng.Intn(10) // 1..10 setters
		values := make([]int, n)
		for i := range values {
			values[i] = rng.Intn(1000)
		}
		opts := make([]option.Option[testConfig], n)
		for i, v := range values {
			v := v
			opts[i] = option.OptionFunc[testConfig](func(c *testConfig) error {
				c.a = v
				return nil
			})
		}
		got, err := option.Apply(opts)
		if err != nil {
			t.Fatalf("iter %d: unexpected error: %v", iter, err)
		}
		want := values[n-1]
		if got.a != want {
			t.Errorf("iter %d: expected last-wins %d, got %d", iter, want, got.a)
		}
	}
}

// ---- T#44 PropertyValidateOrderInvariance ----
// errors.Join result is set-equivalent regardless of validator order.
// We compare the set of unwrapped errors (via errors.Unwrap multi-unwrap).

func TestPropertyValidateOrderInvariance(t *testing.T) {
	cfg := testConfig{}

	v1 := option.Validator[testConfig](func(_ *testConfig) error { return errA })
	v2 := option.Validator[testConfig](func(_ *testConfig) error { return errB })
	v3 := option.Validator[testConfig](func(_ *testConfig) error { return errC })

	orders := [][]option.Validator[testConfig]{
		{v1, v2, v3},
		{v3, v1, v2},
		{v2, v3, v1},
		{v3, v2, v1},
		{v1, v3, v2},
		{v2, v1, v3},
	}

	// unwrapAll returns the slice of errors from a joined error.
	unwrapAll := func(err error) []error {
		type multiUnwrap interface {
			Unwrap() []error
		}
		if u, ok := err.(multiUnwrap); ok {
			return u.Unwrap()
		}
		if err != nil {
			return []error{err}
		}
		return nil
	}

	// Collect result from first ordering.
	ref := unwrapAll(option.Validate(&cfg, orders[0]...))
	refSet := make(map[string]bool, len(ref))
	for _, e := range ref {
		refSet[e.Error()] = true
	}

	for i, order := range orders[1:] {
		got := unwrapAll(option.Validate(&cfg, order...))
		gotSet := make(map[string]bool, len(got))
		for _, e := range got {
			gotSet[e.Error()] = true
		}
		if !reflect.DeepEqual(refSet, gotSet) {
			t.Errorf("order %d: error sets differ: ref=%v got=%v", i+1, refSet, gotSet)
		}
	}
}
