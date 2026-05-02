// SPDX-License-Identifier: Apache-2.0

package mock_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/mock"
)

func TestConcurrentCallsRecordedCorrectly(t *testing.T) {
	const goroutines = 200
	m := mock.Of[Greeter](t)
	m.OnCall("Greet").With(mock.Any[string]()).Return("hi").AnyTimes()

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			m.Interface().Greet("concurrent")
		}()
	}
	wg.Wait()

	calls := m.CallsTo("Greet")
	assert.True(t, len(calls) == goroutines, fmt.Sprintf("expected %d recorded calls, got %d", goroutines, len(calls)))
}

func TestConcurrentCallsAllMatched(t *testing.T) {
	const goroutines = 100
	m := mock.Of[Calculator](t)
	m.OnCall("Add").With(mock.Any[int](), mock.Any[int]()).Return(0).AnyTimes()

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			m.Interface().Add(1, 2)
		}()
	}
	wg.Wait()

	calls := m.CallsTo("Add")
	for _, c := range calls {
		assert.True(t, c.Matched, "all concurrent calls should be matched")
	}
}
