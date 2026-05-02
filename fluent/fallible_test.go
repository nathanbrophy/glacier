// SPDX-License-Identifier: Apache-2.0

package fluent_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/fluent"
)

func TestMapErr(t *testing.T) {
	t.Parallel()

	errBad := errors.New("bad value")

	t.Run("empty source yields nothing", func(t *testing.T) {
		t.Parallel()
		count := 0
		for range fluent.MapErr(fluent.From([]string{}), strconv.Atoi) {
			count++
		}
		assert.Equal(t, count, 0)
	})

	t.Run("all success yields (value, nil) pairs", func(t *testing.T) {
		t.Parallel()
		src := fluent.From([]string{"1", "2", "3"})
		var vals []int
		var errs []error
		for v, err := range fluent.MapErr(src, strconv.Atoi) {
			vals = append(vals, v)
			errs = append(errs, err)
		}
		assert.Equal(t, vals, []int{1, 2, 3})
		for _, err := range errs {
			assert.Equal(t, err, nil)
		}
	})

	t.Run("some failures yield (zero, err) pairs", func(t *testing.T) {
		t.Parallel()
		f := func(v string) (int, error) {
			if v == "bad" {
				return 0, errBad
			}
			n, err := strconv.Atoi(v)
			return n, err
		}
		src := fluent.From([]string{"1", "bad", "3"})
		type pair struct {
			v   int
			err error
		}
		var results []pair
		for v, err := range fluent.MapErr(src, f) {
			results = append(results, pair{v, err})
		}
		assert.Equal(t, len(results), 3)
		assert.Equal(t, results[0], pair{1, nil})
		assert.Equal(t, results[1], pair{0, errBad})
		assert.Equal(t, results[2], pair{3, nil})
	})

	t.Run("caller can break on first error", func(t *testing.T) {
		t.Parallel()
		src := fluent.From([]string{"bad", "1", "2"})
		f := func(v string) (int, error) {
			if v == "bad" {
				return 0, errBad
			}
			return strconv.Atoi(v)
		}
		count := 0
		for _, err := range fluent.MapErr(src, f) {
			count++
			if err != nil {
				break
			}
		}
		assert.Equal(t, count, 1)
	})
}

func TestFilterErr(t *testing.T) {
	t.Parallel()

	errCheck := errors.New("check error")

	t.Run("empty source yields nothing", func(t *testing.T) {
		t.Parallel()
		count := 0
		pred := func(v int) (bool, error) { return v > 0, nil }
		for range fluent.FilterErr(fluent.From([]int{}), pred) {
			count++
		}
		assert.Equal(t, count, 0)
	})

	t.Run("(false, nil) elements are dropped", func(t *testing.T) {
		t.Parallel()
		pred := func(v int) (bool, error) { return v > 2, nil }
		src := fluent.From([]int{1, 2, 3, 4})
		var vals []int
		for v, err := range fluent.FilterErr(src, pred) {
			assert.Equal(t, err, nil)
			vals = append(vals, v)
		}
		assert.Equal(t, vals, []int{3, 4})
	})

	t.Run("errors propagate as (elem, err)", func(t *testing.T) {
		t.Parallel()
		pred := func(v int) (bool, error) {
			if v == 2 {
				return false, errCheck
			}
			return v > 0, nil
		}
		src := fluent.From([]int{1, 2, 3})
		type pair struct {
			v   int
			err error
		}
		var results []pair
		for v, err := range fluent.FilterErr(src, pred) {
			results = append(results, pair{v, err})
		}
		// 1 passes (true, nil); 2 produces (2, errCheck); 3 passes (true, nil)
		assert.Equal(t, len(results), 3)
		assert.Equal(t, results[0], pair{1, nil})
		assert.Equal(t, results[1], pair{2, errCheck})
		assert.Equal(t, results[2], pair{3, nil})
	})

	t.Run("caller can break on first error", func(t *testing.T) {
		t.Parallel()
		pred := func(v int) (bool, error) {
			if v == 1 {
				return false, errCheck
			}
			return true, nil
		}
		src := fluent.From([]int{1, 2, 3})
		count := 0
		for _, err := range fluent.FilterErr(src, pred) {
			count++
			if err != nil {
				break
			}
		}
		assert.Equal(t, count, 1)
	})
}
