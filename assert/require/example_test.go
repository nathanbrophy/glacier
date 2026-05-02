// SPDX-License-Identifier: Apache-2.0

package require_test

import (
	"context"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
)

// ExampleNoError demonstrates halt-on-failure semantics using require.
func ExampleNoError() {
	var t *testing.T

	// Hypothetical pipeline.
	type Pipeline struct{}
	newPipeline := func() (*Pipeline, error) { return &Pipeline{}, nil }

	pipeline, err := newPipeline()
	require.NoError(t, err) // halt here if construction failed

	_ = pipeline
	_ = context.Background()
	_ = assert.TB(nil) // suppress unused import
}

// ExampleEqual demonstrates require.Equal halting the test on failure.
func ExampleEqual() {
	var t *testing.T

	require.Equal(t, "expected", "expected") // passes; test continues
}
