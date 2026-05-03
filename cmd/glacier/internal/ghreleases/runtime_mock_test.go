// SPDX-License-Identifier: Apache-2.0

package ghreleases_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/ghreleases"
	"github.com/nathanbrophy/glacier/mock"
)

// init registers a runtime adapter for ghreleases.ReleaseFetcher with the
// mock package. The adapter is what mock.Of[ReleaseFetcher] uses to build a
// synthesized interface value at test time. mock/gen will eventually emit
// this automatically; until then, registering it here proves the runtime
// (reflection) mock pipeline works end-to-end without any code generation.
func init() {
	mock.RegisterAdapter[ghreleases.ReleaseFetcher](
		func(dispatch func(string, []reflect.Value) []reflect.Value) ghreleases.ReleaseFetcher {
			return &releaseFetcherAdapter{dispatch: dispatch}
		},
	)
}

// releaseFetcherAdapter is a hand-written concrete implementation of
// ghreleases.ReleaseFetcher that forwards every method call into the mock
// package's dispatch function. The mock state inside Mock[T] then matches
// the call against registered expectations and returns the configured values.
type releaseFetcherAdapter struct {
	dispatch func(string, []reflect.Value) []reflect.Value
}

// Latest implements ghreleases.ReleaseFetcher by reflecting the call into
// dispatch. The return slice carries the matched expectation's Return values.
func (a *releaseFetcherAdapter) Latest(ctx context.Context, repo string) (ghreleases.Release, error) {
	results := a.dispatch("Latest", []reflect.Value{
		reflect.ValueOf(&ctx).Elem(),
		reflect.ValueOf(repo),
	})
	rel, _ := results[0].Interface().(ghreleases.Release)
	err, _ := results[1].Interface().(error)
	return rel, err
}

// TestReleaseFetcher_RuntimeMock_HappyPath demonstrates the reflection-based
// runtime mock library by:
//  1. Constructing a Mock[ReleaseFetcher] via mock.Of (no codegen needed).
//  2. Registering a "Latest(any, "owner/repo")" expectation that returns a
//     canned Release.
//  3. Calling Latest through the synthesized interface value.
//  4. Asserting the returned Release is the canned value and verifying
//     the expectation count via the cleanup hook.
//
// This test exercises code paths that are NOT exercised by the
// mock/gen-generated typed wrapper or by the httpmock-based integration
// tests, so it is the canary for "the runtime mock actually works".
func TestReleaseFetcher_RuntimeMock_HappyPath(t *testing.T) {
	t.Parallel()
	m := mock.Of[ghreleases.ReleaseFetcher](t)
	want := ghreleases.Release{
		TagName:     "v1.2.3",
		Name:        "Glacier v1.2.3",
		HTMLURL:     "https://github.com/nathanbrophy/glacier/releases/tag/v1.2.3",
		PublishedAt: time.Date(2026, 5, 3, 0, 0, 0, 0, time.UTC),
	}
	m.OnCall("Latest").
		With(mock.Any[context.Context](), mock.Eq("nathanbrophy/glacier")).
		Return(want, error(nil))

	got, err := m.Interface().Latest(context.Background(), "nathanbrophy/glacier")
	assert.NoError(t, err)
	assert.Equal(t, want.TagName, got.TagName)
	assert.Equal(t, want.HTMLURL, got.HTMLURL)
}

// TestReleaseFetcher_RuntimeMock_ErrorPath confirms the runtime mock can
// return an error from a method, that the error round-trips through the
// reflection plumbing intact, and that the SDK's caller-side error handling
// (e.g. RateLimitError) is preserved by simply returning a typed error.
func TestReleaseFetcher_RuntimeMock_ErrorPath(t *testing.T) {
	t.Parallel()
	wantErr := &ghreleases.RateLimitError{Cause: errors.New("HTTP 403 from GitHub")}

	m := mock.Of[ghreleases.ReleaseFetcher](t)
	m.OnCall("Latest").
		With(mock.Any[context.Context](), mock.Eq("nathanbrophy/glacier")).
		Return(ghreleases.Release{}, error(wantErr))

	_, err := m.Interface().Latest(context.Background(), "nathanbrophy/glacier")
	assert.NotNil(t, err)

	var rl *ghreleases.RateLimitError
	assert.True(t, errors.As(err, &rl), "expected RateLimitError to round-trip through the runtime mock")
}

// TestReleaseFetcher_RuntimeMock_MultipleCalls verifies that successive
// calls to the same method dispatch to the correct expectation in order
// (mock.Mock's default mode is "ordered match"; the first matching
// expectation wins per call).
func TestReleaseFetcher_RuntimeMock_MultipleCalls(t *testing.T) {
	t.Parallel()
	m := mock.Of[ghreleases.ReleaseFetcher](t)

	first := ghreleases.Release{TagName: "v1.0.0"}
	second := ghreleases.Release{TagName: "v1.1.0"}

	m.OnCall("Latest").With(mock.Any[context.Context](), mock.Eq("repo/a")).Return(first, error(nil))
	m.OnCall("Latest").With(mock.Any[context.Context](), mock.Eq("repo/b")).Return(second, error(nil))

	gotA, _ := m.Interface().Latest(context.Background(), "repo/a")
	gotB, _ := m.Interface().Latest(context.Background(), "repo/b")

	assert.Equal(t, "v1.0.0", gotA.TagName)
	assert.Equal(t, "v1.1.0", gotB.TagName)
}
