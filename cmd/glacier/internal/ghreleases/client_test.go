// SPDX-License-Identifier: Apache-2.0

package ghreleases_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/ghreleases"
	"github.com/nathanbrophy/glacier/httpc"
	"github.com/nathanbrophy/glacier/httpmock"
)

// published is a fixed timestamp used across all table rows.
var published = time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC)

func TestLatest(t *testing.T) {
	t.Parallel()

	type row struct {
		name       string
		responder  httpmock.Responder
		wantTag    string
		wantErrIs  error // checked with errors.Is when non-nil
		wantErrAs  any   // pointer to target for errors.As when non-nil
	}

	rows := []row{
		{
			name: "happy path",
			responder: httpmock.JSON(http.StatusOK, ghreleases.Release{
				TagName:     "v0.1.2",
				Name:        "Release v0.1.2",
				HTMLURL:     "https://github.com/nathanbrophy/glacier/releases/tag/v0.1.2",
				PublishedAt: published,
			}),
			wantTag: "v0.1.2",
		},
		{
			name: "pre-release tag accepted",
			responder: httpmock.JSON(http.StatusOK, ghreleases.Release{
				TagName:     "v1.0.0-rc.1",
				HTMLURL:     "https://github.com/nathanbrophy/glacier/releases/tag/v1.0.0-rc.1",
				PublishedAt: published,
			}),
			wantTag: "v1.0.0-rc.1",
		},
		{
			name:      "http 403 becomes RateLimitError",
			responder: httpmock.Status(http.StatusForbidden),
			wantErrAs: new(*ghreleases.RateLimitError),
		},
		{
			name:      "http 500 returns error",
			responder: httpmock.Status(http.StatusInternalServerError),
			wantErrAs: new(*httpc.StatusError),
		},
		{
			name:      "transport error propagated",
			responder: httpmock.Error(errors.New("dial tcp: connect refused")),
		},
		{
			name: "invalid tag format rejected",
			responder: httpmock.JSON(http.StatusOK, ghreleases.Release{
				TagName:     "not-a-semver",
				HTMLURL:     "https://github.com/nathanbrophy/glacier/releases/tag/bad",
				PublishedAt: published,
			}),
		},
		{
			name: "invalid html_url prefix rejected",
			responder: httpmock.JSON(http.StatusOK, ghreleases.Release{
				TagName:     "v1.0.0",
				HTMLURL:     "https://evil.example.com/releases/tag/v1.0.0",
				PublishedAt: published,
			}),
		},
	}

	for _, r := range rows {
		r := r
		t.Run(r.name, func(t *testing.T) {
			t.Parallel()

			rt := httpmock.NewWithT(t)
			rt.OnRequest().
				Method("GET").
				PathPrefix("/repos/").
				Respond(r.responder)

			hc := httpc.New(httpc.WithTransport(rt))
			fetcher := ghreleases.NewClient(ghreleases.WithHTTPClient(hc))

			rel, err := fetcher.Latest(context.Background(), "nathanbrophy/glacier")

			if r.wantTag != "" {
				assert.Equal(t, r.wantTag, rel.TagName)
				assert.NoError(t, err)
				return
			}

			assert.Error(t, err)

			if r.wantErrIs != nil {
				assert.True(t, errors.Is(err, r.wantErrIs), "errors.Is match")
			}
			if r.wantErrAs != nil {
				assert.True(t, errors.As(err, r.wantErrAs), "errors.As match")
			}
		})
	}
}

func TestLatestInvalidRepo(t *testing.T) {
	t.Parallel()
	fetcher := ghreleases.NewClient()
	_, err := fetcher.Latest(context.Background(), "no-slash")
	assert.Error(t, err)
}
