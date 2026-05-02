// SPDX-License-Identifier: Apache-2.0

package httpmock_test

import (
	"net/http"
	"testing"

	"github.com/nathanbrophy/glacier/httpmock"
)

func BenchmarkRoundTripFirstStubMatches(b *testing.B) {
	rt := httpmock.New()
	rt.OnRequest().Method("GET").Path("/bench").AnyTimes().Respond(httpmock.Status(200))

	req := mustNewReq("GET", "https://example.com/bench")
	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		resp, err := rt.RoundTrip(req)
		if err != nil {
			b.Fatal(err)
		}
		_ = resp
	}
}

func BenchmarkRoundTripScanThirty(b *testing.B) {
	rt := httpmock.New()
	for i := range 29 {
		path := "/bench/miss/" + string(rune('a'+i))
		rt.OnRequest().Path(path).AnyTimes().Respond(httpmock.Status(200))
	}
	rt.OnRequest().Path("/bench/hit").AnyTimes().Respond(httpmock.Status(200))

	req := mustNewReq("GET", "https://example.com/bench/hit")
	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		resp, err := rt.RoundTrip(req)
		if err != nil {
			b.Fatal(err)
		}
		_ = resp
	}
}

func BenchmarkBodyJSONMatch(b *testing.B) {
	type Item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	matcher := httpmock.BodyJSON(Item{ID: 1, Name: "foo"})
	body := []byte(`{"id":1,"name":"foo"}`)

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_ = matcher.Match(body, "application/json")
	}
}

func BenchmarkResponseJSON(b *testing.B) {
	type Item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	rt := httpmock.New()
	rt.OnRequest().Path("/bench").AnyTimes().Respond(httpmock.JSON(200, Item{ID: 1, Name: "foo"}))
	req := mustNewReq("GET", "https://example.com/bench")

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		resp, err := rt.RoundTrip(req)
		if err != nil {
			b.Fatal(err)
		}
		_ = resp
	}
}

// mustNewReq constructs a request; panics on error.
func mustNewReq(method, url string) *http.Request {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic("mustNewReq: " + err.Error())
	}
	return req
}
