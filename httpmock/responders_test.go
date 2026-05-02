// SPDX-License-Identifier: Apache-2.0

package httpmock_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/httpmock"
)

type testUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestRespondJSON(t *testing.T) {
	rt := httpmock.New()
	rt.OnRequest().Path("/user").Respond(httpmock.JSON(200, testUser{ID: 1, Name: "Ada"}))

	req := newReq(t, "GET", "https://example.com/user", nil)
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
	assert.Equal(t, resp.Header.Get("Content-Type"), "application/json")

	var u testUser
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&u))
	assert.Equal(t, u.ID, 1)
	assert.Equal(t, u.Name, "Ada")
}

func TestRespondJSONFrom(t *testing.T) {
	raw := `{"id":2,"name":"Grace"}`
	rt := httpmock.New()
	rt.OnRequest().Path("/user").Respond(httpmock.JSONFrom[testUser](200, strings.NewReader(raw)))

	req := newReq(t, "GET", "https://example.com/user", nil)
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	var u testUser
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&u))
	assert.Equal(t, u.ID, 2)
	assert.Equal(t, u.Name, "Grace")
}

func TestRespondJSONFromMalformed(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r, "expected panic for malformed JSON")
	}()
	_ = httpmock.JSONFrom[testUser](200, strings.NewReader(`{not valid json`))
}

func TestRespondStatus(t *testing.T) {
	rt := httpmock.New()
	rt.OnRequest().Path("/ping").Respond(httpmock.Status(204))

	req := newReq(t, "GET", "https://example.com/ping", nil)
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 204)

	// Body must be non-nil.
	assert.NotNil(t, resp.Body)
	data, _ := io.ReadAll(resp.Body)
	assert.Len(t, data, 0)
}

func TestRespondBody(t *testing.T) {
	rt := httpmock.New()
	rt.OnRequest().Path("/text").Respond(httpmock.Body(200, []byte("hello"), "text/plain"))

	req := newReq(t, "GET", "https://example.com/text", nil)
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
	assert.Equal(t, resp.Header.Get("Content-Type"), "text/plain")

	data, _ := io.ReadAll(resp.Body)
	assert.Equal(t, string(data), "hello")
}

func TestRespondStream(t *testing.T) {
	rt := httpmock.New()
	pr, pw := io.Pipe()
	go func() {
		pw.Write([]byte("streamed"))
		pw.Close()
	}()
	rt.OnRequest().Path("/stream").Respond(httpmock.Stream(200, pr, "text/event-stream"))

	req := newReq(t, "GET", "https://example.com/stream", nil)
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	data, _ := io.ReadAll(resp.Body)
	assert.Equal(t, string(data), "streamed")
}

func TestRespondStreamReaderError(t *testing.T) {
	rt := httpmock.New()
	pr, pw := io.Pipe()
	go func() {
		pw.CloseWithError(errors.New("read failure"))
	}()
	rt.OnRequest().Path("/stream").Respond(httpmock.Stream(200, pr, "text/event-stream"))

	req := newReq(t, "GET", "https://example.com/stream", nil)
	resp, err := rt.RoundTrip(req)
	// Stream responder succeeds (err is nil), but reading body fails.
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	_, readErr := io.ReadAll(resp.Body)
	assert.Error(t, readErr)
}

func TestRespondError(t *testing.T) {
	rt := httpmock.New()
	rt.OnRequest().Path("/fail").Respond(httpmock.Error(context.DeadlineExceeded))

	req := newReq(t, "GET", "https://example.com/fail", nil)
	_, err := rt.RoundTrip(req)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestRespondJSONPanicsOnMarshalFailure(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r, "expected panic for unmarshalable type")
	}()
	// Channels cannot be marshaled.
	ch := make(chan int)
	_ = httpmock.JSON(200, ch)
}

func TestRespondBodyNilBody(t *testing.T) {
	rt := httpmock.New()
	rt.OnRequest().Path("/empty").Respond(httpmock.Body(200, nil, ""))

	req := newReq(t, "GET", "https://example.com/empty", nil)
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
	assert.NotNil(t, resp.Body)

	data, _ := io.ReadAll(resp.Body)
	assert.Len(t, data, 0)
}

func TestRoundTripNeverDials(t *testing.T) {
	// This test verifies the runtime property: a successful RoundTrip produces
	// a response without ever needing a real network connection. The import
	// audit is covered by TestNoNetworkImports in import_audit_test.go.
	rt := httpmock.New()
	rt.OnRequest().Path("/ok").Respond(httpmock.Status(200))

	// Using http.Client to route through Transport.
	client := &http.Client{Transport: rt}
	resp, err := client.Get("https://definitely-not-a-real-server.invalid/ok")
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestEmptyBodyMatcherSemantics(t *testing.T) {
	rt := httpmock.New()
	rt.OnRequest().Path("/x").Body(httpmock.BodyContains("")).Respond(httpmock.Status(200))

	// Request with no body still matches BodyContains("").
	req := newReq(t, "GET", "https://example.com/x", nil)
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestWithDefaultStatusPanic(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r, "expected panic for status out of range")
	}()
	_ = httpmock.WithDefaultStatus(99)
}

func TestWithLoggerPanic(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r, "expected panic for nil logger")
	}()
	_ = httpmock.WithLogger(nil)
}

func TestStubTimesZeroPanic(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r, "expected panic for Times(0)")
	}()
	rt := httpmock.New()
	rt.OnRequest().Times(0)
}

func TestBodyJSONTypeParam(t *testing.T) {
	// Compile-time check: BodyJSON carries T.
	type MyType struct{ Val string }
	m := httpmock.BodyJSON(MyType{Val: "hi"})
	assert.NotNil(t, m)
}

func TestRespondBodyHTTPClient(t *testing.T) {
	// Verify Body response is readable through http.Client.
	rt := httpmock.New()
	rt.OnRequest().Path("/data").Respond(httpmock.Body(200, []byte(`{"ok":true}`), "application/json"))
	client := &http.Client{Transport: rt}
	resp, err := client.Get("https://api.example.com/data")
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
	defer resp.Body.Close()

	var m map[string]any
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&m))
	assert.Equal(t, m["ok"], true)
}
