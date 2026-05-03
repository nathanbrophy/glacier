// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/cache"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/ghreleases"
	"github.com/nathanbrophy/glacier/httpc"
	"github.com/nathanbrophy/glacier/httpmock"
)

// fakeRelease is a minimal ghreleases.Release used across version tests.
var fakeRelease = ghreleases.Release{
	TagName:     "v0.2.0",
	Name:        "Release v0.2.0",
	HTMLURL:     "https://github.com/nathanbrophy/glacier/releases/tag/v0.2.0",
	PublishedAt: time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC),
}

// runVersionCmd runs cmd and returns its text output as a string.
// Uses withWriter so tests are safe to run in parallel without os.Stdout races,
// and withCache so tests are isolated from the per-user disk cache that the
// default constructor would otherwise use.
func runVersionCmd(t *testing.T, cmd *VersionCmd) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	c := cmd.
		withWriter(&buf).
		withCache(cache.New[ghreleases.Release]()) // fresh in-memory cache per test
	err := c.Run(context.Background())
	return buf.String(), err
}

func TestVersionDefaultOutput(t *testing.T) {
	t.Parallel()

	out, err := runVersionCmd(t, &VersionCmd{})
	assert.NoError(t, err)

	assert.True(t, strings.Contains(out, "go:"), "output should contain go: field")
	assert.True(t, strings.Contains(out, "os:"), "output should contain os: field")
}

func TestVersionCheckFreshCacheAlwaysMisses(t *testing.T) {
	// Verify that --check with an httpmock-backed client reaches the network
	// layer (cache always misses in the stub) and parses the response.
	t.Parallel()

	rt := httpmock.NewWithT(t)
	rt.OnRequest().
		Method("GET").
		PathPrefix("/repos/nathanbrophy/glacier/").
		Respond(httpmock.JSON(http.StatusOK, fakeRelease))

	hc := httpc.New(httpc.WithTransport(rt))
	fetcher := ghreleases.NewClient(ghreleases.WithHTTPClient(hc))

	out, err := runVersionCmd(t, (&VersionCmd{Check: true}).withFetcher(fetcher))
	assert.NoError(t, err)
	assert.True(t, strings.Contains(out, "v0.2.0"), "latest tag should appear in output")
}

func TestVersionCheckOfflineGraceful(t *testing.T) {
	// --check with a transport that returns a network error should exit 0 and
	// print the offline annotation.
	t.Parallel()

	rt := httpmock.NewWithT(t)
	rt.OnRequest().
		PathPrefix("/repos/").
		Respond(httpmock.Error(errors.New("dial tcp: no route to host")))

	hc := httpc.New(httpc.WithTransport(rt))
	fetcher := ghreleases.NewClient(ghreleases.WithHTTPClient(hc))

	_, err := runVersionCmd(t, (&VersionCmd{Check: true}).withFetcher(fetcher))
	assert.NoError(t, err, "--check offline should exit 0 without --strict")
}

func TestVersionCheckOfflineStrict(t *testing.T) {
	// --check --strict with a transport error should return exitCodeError(68).
	t.Parallel()

	rt := httpmock.NewWithT(t)
	rt.OnRequest().
		PathPrefix("/repos/").
		Respond(httpmock.Error(errors.New("dial tcp: no route to host")))

	hc := httpc.New(httpc.WithTransport(rt))
	fetcher := ghreleases.NewClient(ghreleases.WithHTTPClient(hc))

	_, err := runVersionCmd(t, (&VersionCmd{Check: true, Strict: true}).withFetcher(fetcher))
	assert.Error(t, err)

	var ecErr *exitCodeError
	assert.True(t, errors.As(err, &ecErr), "error should be exitCodeError")
	assert.Equal(t, exitVersionCheck, ecErr.ExitCode())
}

func TestVersionCheckRateLimit(t *testing.T) {
	// HTTP 403 is treated the same as offline (graceful by default).
	t.Parallel()

	rt := httpmock.NewWithT(t)
	rt.OnRequest().
		PathPrefix("/repos/").
		Respond(httpmock.Status(http.StatusForbidden))

	hc := httpc.New(httpc.WithTransport(rt))
	fetcher := ghreleases.NewClient(ghreleases.WithHTTPClient(hc))

	_, err := runVersionCmd(t, (&VersionCmd{Check: true}).withFetcher(fetcher))
	assert.NoError(t, err, "HTTP 403 should exit 0 without --strict")
}

func TestVersionJSONSchema(t *testing.T) {
	// --json output must include the required top-level fields.
	t.Parallel()

	out, err := runVersionCmd(t, &VersionCmd{JSON: true})
	assert.NoError(t, err)

	assert.True(t, strings.Contains(out, `"version"`), "json must have version field")
	assert.True(t, strings.Contains(out, `"go_version"`), "json must have go_version field")
	assert.True(t, strings.Contains(out, `"os"`), "json must have os field")
	assert.True(t, strings.Contains(out, `"arch"`), "json must have arch field")
}

func TestVersionTagAllowlist(t *testing.T) {
	// A tag value containing ANSI escapes is rejected before display.
	t.Parallel()

	rt := httpmock.NewWithT(t)
	rt.OnRequest().
		PathPrefix("/repos/").
		Respond(httpmock.JSON(http.StatusOK, ghreleases.Release{
			TagName:     "\x1b[31mv1.0.0\x1b[0m", // ANSI-escaped tag
			HTMLURL:     "https://github.com/nathanbrophy/glacier/releases/tag/v1.0.0",
			PublishedAt: fakeRelease.PublishedAt,
		}))

	hc := httpc.New(httpc.WithTransport(rt))
	fetcher := ghreleases.NewClient(ghreleases.WithHTTPClient(hc))

	// The ANSI-escaped tag fails tagRe validation; --check degrades gracefully.
	_, err := runVersionCmd(t, (&VersionCmd{Check: true}).withFetcher(fetcher))
	assert.NoError(t, err, "invalid tag should degrade gracefully")
}
