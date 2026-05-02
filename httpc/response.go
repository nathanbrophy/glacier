// SPDX-License-Identifier: Apache-2.0

package httpc

import (
	"io"
	"net/http"
	"time"
)

// Response wraps *http.Response with httpc-specific metadata.
type Response struct {
	*http.Response
	// Body holds the bytes read and buffered by typed methods (Get, Post, etc.).
	// Nil for Head and Do; the caller owns the body via http.Response.Body in
	// those cases.
	Body []byte
	// Elapsed is the wall-clock time from the first byte sent to the last byte
	// of the body read.
	Elapsed time.Duration
}

// Drain discards and closes any unread response body. Call Drain when you
// have a *Response but do not intend to read the body, to release the
// underlying TCP connection. Safe to call when Body is already fully read.
//
// Concurrency: not goroutine-safe; do not call Drain concurrently with body reads.
func (r *Response) Drain() error {
	if r == nil || r.Response == nil || r.Response.Body == nil {
		return nil
	}
	_, _ = io.Copy(io.Discard, r.Response.Body)
	return r.Response.Body.Close()
}
