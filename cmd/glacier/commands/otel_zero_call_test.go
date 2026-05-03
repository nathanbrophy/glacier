// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/cache"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/ghreleases"
	"github.com/nathanbrophy/glacier/httpc"
	"github.com/nathanbrophy/glacier/httpmock"
)

// TestOTEL_ZeroOutboundWhenUnset verifies that with OTEL_EXPORTER_OTLP_ENDPOINT
// unset (D-S24), running version --check issues zero outbound calls to any host
// other than api.github.com. The httpmock transport records every request;
// we count requests whose host is NOT api.github.com and assert zero.
func TestOTEL_ZeroOutboundWhenUnset(t *testing.T) {
	// Serialize: mutates env.
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")

	rt := httpmock.NewWithT(t)
	rt.OnRequest().
		Method("GET").
		PathPrefix("/repos/").
		Respond(httpmock.JSON(http.StatusOK, fakeRelease))

	hc := httpc.New(httpc.WithTransport(rt))
	fetcher := ghreleases.NewClient(ghreleases.WithHTTPClient(hc))

	var buf bytes.Buffer
	cmd := (&VersionCmd{Check: true}).
		withWriter(&buf).
		withFetcher(fetcher).
		withCache(cache.New[ghreleases.Release]())

	err := cmd.Run(context.Background())
	assert.NoError(t, err)

	// Count requests to hosts other than api.github.com.
	var nonGitHub int
	for _, req := range rt.AllRequests() {
		if req.URL.Host != "api.github.com" {
			nonGitHub++
		}
	}
	assert.True(t, nonGitHub == 0,
		"expected zero outbound calls to non-github hosts when OTEL endpoint is unset, got", nonGitHub)
}
