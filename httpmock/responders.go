// SPDX-License-Identifier: Apache-2.0

package httpmock

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
)

// Responder produces an *http.Response (or transport-level error) for a matched
// request. Implementations must be goroutine-safe.
type Responder interface {
	Respond(req *http.Request) (*http.Response, error)
}

func makeResponse(req *http.Request, status int, h http.Header, body []byte) *http.Response {
	if h == nil {
		h = make(http.Header)
	}
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header:     h,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Request:    req,
	}
}

// jsonResponder marshals a value as JSON at construction time.
type jsonResponder struct {
	status int
	body   []byte
}

// JSON returns a Responder that marshals body as JSON and sets Content-Type: application/json.
// Panics at call time if json.Marshal fails.
func JSON[T any](status int, body T) Responder {
	data, err := json.Marshal(body)
	if err != nil {
		panic(fmt.Sprintf("httpmock: JSON: marshal %T: %s", body, err))
	}
	return &jsonResponder{status: status, body: data}
}

func (r *jsonResponder) Respond(req *http.Request) (*http.Response, error) {
	h := make(http.Header)
	h.Set("Content-Type", "application/json")
	return makeResponse(req, r.status, h, r.body), nil
}

// jsonFromResponder serves pre-read, validated JSON bytes.
type jsonFromResponder struct {
	status int
	body   []byte
}

// JSONFrom reads all bytes from r, validates JSON decoding into T, and serves
// the bytes on each RoundTrip. Panics if reading or decoding fails.
func JSONFrom[T any](status int, r io.Reader) Responder {
	data, err := io.ReadAll(r)
	if err != nil {
		panic("httpmock: JSONFrom: read: " + err.Error())
	}
	var zero T
	if err := json.Unmarshal(data, &zero); err != nil {
		panic("httpmock: JSONFrom: decode: " + err.Error())
	}
	return &jsonFromResponder{status: status, body: data}
}

func (r *jsonFromResponder) Respond(req *http.Request) (*http.Response, error) {
	h := make(http.Header)
	h.Set("Content-Type", "application/json")
	return makeResponse(req, r.status, h, r.body), nil
}

// Status returns a Responder that produces an empty-body response.
func Status(status int) Responder { return &statusResponder{status: status} }

type statusResponder struct{ status int }

func (r *statusResponder) Respond(req *http.Request) (*http.Response, error) {
	return makeResponse(req, r.status, nil, nil), nil
}

// Body returns a Responder with raw bytes and a Content-Type header.
func Body(status int, body []byte, contentType string) Responder {
	cp := make([]byte, len(body))
	copy(cp, body)
	return &bodyResponder{status: status, body: cp, contentType: contentType}
}

type bodyResponder struct {
	status      int
	body        []byte
	contentType string
}

func (r *bodyResponder) Respond(req *http.Request) (*http.Response, error) {
	h := make(http.Header)
	if r.contentType != "" {
		h.Set("Content-Type", r.contentType)
	}
	return makeResponse(req, r.status, h, r.body), nil
}

// Stream returns a Responder that streams r as the response body.
func Stream(status int, body io.Reader, contentType string) Responder {
	return &streamResponder{status: status, body: body, contentType: contentType}
}

type streamResponder struct {
	status      int
	body        io.Reader
	contentType string
}

func (r *streamResponder) Respond(req *http.Request) (*http.Response, error) {
	h := make(http.Header)
	if r.contentType != "" {
		h.Set("Content-Type", r.contentType)
	}
	var rc io.ReadCloser
	if c, ok := r.body.(io.ReadCloser); ok {
		rc = c
	} else {
		rc = io.NopCloser(r.body)
	}
	return &http.Response{
		StatusCode: r.status,
		Status:     fmt.Sprintf("%d %s", r.status, http.StatusText(r.status)),
		Header:     h,
		Body:       rc,
		Request:    req,
	}, nil
}

// Error returns a Responder that surfaces err as a transport-level error.
func Error(err error) Responder { return &errorResponder{err: err} }

type errorResponder struct{ err error }

func (r *errorResponder) Respond(_ *http.Request) (*http.Response, error) { return nil, r.err }

// sequenceResponder cycles or exhausts through a list of responders.
type sequenceResponder struct {
	rs    []Responder
	idx   atomic.Int64
	cycle bool
}

// Sequence serves responders in order, cycling after exhaustion. Equivalent to SequenceCycle.
func Sequence(rs ...Responder) Responder { return SequenceCycle(rs...) }

// SequenceCycle serves responders in order, cycling after exhaustion.
func SequenceCycle(rs ...Responder) Responder {
	if len(rs) == 0 {
		panic("httpmock: SequenceCycle: at least one responder required")
	}
	return &sequenceResponder{rs: rs, cycle: true}
}

// SequenceExhaust serves responders in order; returns an error after all are exhausted.
func SequenceExhaust(rs ...Responder) Responder {
	if len(rs) == 0 {
		panic("httpmock: SequenceExhaust: at least one responder required")
	}
	return &sequenceResponder{rs: rs, cycle: false}
}

func (s *sequenceResponder) Respond(req *http.Request) (*http.Response, error) {
	idx := s.idx.Add(1) - 1
	n := int64(len(s.rs))
	if idx >= n {
		if !s.cycle {
			return nil, errors.New("httpmock: sequence exhausted")
		}
		idx = idx % n
	} else if s.cycle {
		idx = idx % n
	}
	return s.rs[idx].Respond(req)
}
