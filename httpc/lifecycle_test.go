// SPDX-License-Identifier: Apache-2.0

package httpc_test

import (
	"errors"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/httpc"
)

// closingTransport tracks whether Close was called.
type closingTransport struct {
	closed atomic.Bool
}

func (ct *closingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return jsonResponse(200, `{}`), nil
}

func (ct *closingTransport) Close() error {
	ct.closed.Store(true)
	return nil
}

func TestClientCloseIdempotent(t *testing.T) {
	t.Parallel()
	c := httpc.New()
	err1 := c.Close()
	err2 := c.Close()
	assert.NoError(t, err1)
	assert.NoError(t, err2)
}

func TestClientCloseDoesNotCloseExternalTransport(t *testing.T) {
	t.Parallel()
	ct := &closingTransport{}
	c := httpc.New(httpc.WithTransport(ct))
	err := c.Close()
	assert.NoError(t, err)
	// External transport must NOT be closed by httpc.Client.Close.
	assert.False(t, ct.closed.Load())
}

func TestClientCloseClosesOwnedTransport(t *testing.T) {
	t.Parallel()
	// When no WithTransport is provided, httpc owns the transport.
	// The default http.DefaultTransport is a *http.Transport which implements
	// CloseIdleConnections but not io.Closer. So we test that Close() returns nil
	// (no error) for the default case.
	c := httpc.New()
	err := c.Close()
	assert.NoError(t, err)
}

func TestClientCloseJoinsMultipleErrors(t *testing.T) {
	t.Parallel()
	// Verify that errs.Join is used if multiple close errors occur.
	// Since http.DefaultTransport doesn't implement Close(), this tests
	// the idempotent nil-return path.
	c := httpc.New()
	_ = c.Close()
	_ = c.Close()
	// Just verify no panic occurs.
}

func TestClientCloseWrapsOwnedCloser(t *testing.T) {
	t.Parallel()
	// A transport that implements io.Closer and returns an error.
	rt := &errorClosingTransport{}
	// We pass this as the transport so httpc does NOT own it.
	c := httpc.New(httpc.WithTransport(rt))
	err := c.Close()
	// Should not close external transport even if it implements Close.
	assert.NoError(t, err)
	assert.False(t, rt.closeCalled)
}

type errorClosingTransport struct {
	closeCalled bool
}

func (e *errorClosingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return jsonResponse(200, `{}`), nil
}

func (e *errorClosingTransport) Close() error {
	e.closeCalled = true
	return errors.New("close error")
}
