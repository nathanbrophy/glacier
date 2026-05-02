// SPDX-License-Identifier: Apache-2.0

// Run with: go test -race ./option/...

package option_test

import (
	"sync"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/option"
)

// ---- T#39 TestApplyConcurrent ----
// 100 goroutines call Apply(sameOpts) each with its own local T.
// Verifies goroutine-safety of Apply (NF2).

func TestApplyConcurrent(t *testing.T) {
	opts := []option.Option[testConfig]{
		withA(1),
		withB("concurrent"),
		withC(true),
	}

	const goroutines = 100
	var wg sync.WaitGroup
	errs := make([]error, goroutines)

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()
			_, err := option.Apply(opts)
			errs[i] = err
		}()
	}
	wg.Wait()

	for i, err := range errs {
		assert.NoError(t, err, "goroutine %d", i)
	}
}

// ---- T#40 TestValidateConcurrent ----
// 100 goroutines call Validate(&t, vs...) on independent t values.
// Verifies goroutine-safety of Validate (NF2).

func TestValidateConcurrent(t *testing.T) {
	validators := []option.Validator[testConfig]{
		func(_ *testConfig) error { return nil },
		func(_ *testConfig) error { return nil },
		func(_ *testConfig) error { return nil },
	}

	const goroutines = 100
	var wg sync.WaitGroup
	errs := make([]error, goroutines)

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()
			cfg := testConfig{a: i}
			errs[i] = option.Validate(&cfg, validators...)
		}()
	}
	wg.Wait()

	for i, err := range errs {
		assert.NoError(t, err, "goroutine %d", i)
	}
}

// ---- L-add-2: write-then-read variant inside option, under -race ----
// Options that close over a goroutine-shared variable (read-only in Apply).
// The shared slice itself is read concurrently; only the local *T is written.

func TestApplyConcurrentSharedOpts(t *testing.T) {
	sharedValue := 42 // read-only after init
	opts := []option.Option[testConfig]{
		option.OptionFunc[testConfig](func(c *testConfig) error {
			c.a = sharedValue // read of shared; write of local *T
			return nil
		}),
	}

	const goroutines = 100
	var wg sync.WaitGroup
	errs := make([]error, goroutines)

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()
			_, err := option.Apply(opts)
			errs[i] = err
		}()
	}
	wg.Wait()

	for i, err := range errs {
		assert.NoError(t, err, "goroutine %d", i)
	}
}
