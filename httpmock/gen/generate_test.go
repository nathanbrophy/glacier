// SPDX-License-Identifier: Apache-2.0

package gen_test

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	"github.com/nathanbrophy/glacier/httpmock/gen"
)

// fakeScanner is a test double for gen.PackageScanner.
type fakeScanner struct {
	count int
	err   error
}

func (f *fakeScanner) Scan(_, _ string, _ *slog.Logger) (int, error) {
	return f.count, f.err
}

var _ gen.PackageScanner = (*fakeScanner)(nil)

func TestGeneratePatternRequired(t *testing.T) {
	t.Parallel()
	err := gen.Generate(gen.Options{})
	assert.Error(t, err)
}

func TestGenerateWithFakeScanner_Success(t *testing.T) {
	t.Parallel()
	sc := &fakeScanner{count: 5}
	err := gen.GenerateWith(gen.Options{Pattern: ".", Logger: slog.Default()}, sc)
	require.NoError(t, err)
}

func TestGenerateWithFakeScanner_ScanError(t *testing.T) {
	t.Parallel()
	sc := &fakeScanner{err: errors.New("scan: injected failure")}
	err := gen.GenerateWith(gen.Options{Pattern: ".", Logger: slog.Default()}, sc)
	assert.Error(t, err)
}

func TestGenerateCheckMode_AlwaysNil(t *testing.T) {
	t.Parallel()
	// In v0, httpmock/gen emits no files, so Check mode with a clean tree is nil.
	sc := &fakeScanner{count: 3}
	err := gen.GenerateWith(gen.Options{
		Pattern: ".",
		Check:   true,
		Logger:  slog.Default(),
	}, sc)
	require.NoError(t, err)
}
