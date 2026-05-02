// SPDX-License-Identifier: Apache-2.0

package fluent_test

import (
	"sync"
	"testing"

	"github.com/nathanbrophy/glacier/fluent"
)

func TestRaceFromSlice(t *testing.T) {
	t.Parallel()

	src := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	const goroutines = 8

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			got := fluent.ToSlice(fluent.From(src))
			_ = got
		}()
	}
	wg.Wait()
}

func TestRaceRange(t *testing.T) {
	t.Parallel()

	const goroutines = 8

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := range goroutines {
		i := i
		go func() {
			defer wg.Done()
			got := fluent.ToSlice(fluent.Range(i*10, i*10+5, 1))
			_ = got
		}()
	}
	wg.Wait()
}

func TestRaceDistinct(t *testing.T) {
	t.Parallel()

	src := []int{1, 1, 2, 2, 3, 3}
	const goroutines = 8

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			got := fluent.ToSlice(fluent.Distinct(fluent.From(src)))
			_ = got
		}()
	}
	wg.Wait()
}
