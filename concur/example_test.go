// SPDX-License-Identifier: Apache-2.0

package concur_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/nathanbrophy/glacier/concur"
)

// ExampleGroup shows basic goroutine error collection.
func ExampleGroup() {
	g := concur.NewGroup(concur.WithLimit(4))
	ctx := context.Background()

	sentinel := errors.New("task failed")
	for i := range 3 {
		i := i
		g.Go(ctx, func() error {
			if i == 1 {
				return sentinel
			}
			return nil
		})
	}

	err := g.WaitDone(ctx)
	fmt.Println(errors.Is(err, sentinel))
	// Output:
	// true
}

// ExampleSemaphore shows bounded concurrency with a semaphore.
func ExampleSemaphore() {
	s, err := concur.NewSemaphore(2)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()

	// Acquire 1 permit.
	if err := s.Acquire(ctx, 1); err != nil {
		panic(err)
	}

	// TryAcquire the remaining permit.
	ok, err := s.TryAcquire(1)
	if err != nil {
		panic(err)
	}
	fmt.Println(ok) // true

	// Now at capacity — TryAcquire should fail.
	ok2, err := s.TryAcquire(1)
	if err != nil {
		panic(err)
	}
	fmt.Println(ok2) // false

	_ = s.Release(2)
	// Output:
	// true
	// false
}

// ExamplePool shows typed pool usage.
func ExamplePool() {
	p := concur.NewPool(func() []byte {
		return make([]byte, 0, 64)
	})

	buf := p.Get()
	buf = append(buf, "hello"...)
	fmt.Println(string(buf))
	p.Put(buf[:0]) // reset and return
	// Output:
	// hello
}
