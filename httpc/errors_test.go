// SPDX-License-Identifier: Apache-2.0

package httpc_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/httpc"
)

func TestStatusErrorContains(t *testing.T) {
	t.Parallel()
	e := &httpc.StatusError{Status: 404, Body: []byte("not found body")}
	assert.Equal(t, 404, e.Status)
	assert.True(t, len(e.Body) > 0)
}

func TestStatusErrorErrorOmitsBody(t *testing.T) {
	t.Parallel()
	e := &httpc.StatusError{Status: 500, Body: []byte("sensitive body data")}
	msg := e.Error()
	assert.True(t, strings.Contains(msg, "500"))
	assert.False(t, strings.Contains(msg, "sensitive body data"))
}

func TestBodyParseErrorErrorOmitsBody(t *testing.T) {
	t.Parallel()
	cause := errors.New("json: unexpected token")
	e := &httpc.BodyParseError{
		Cause:       cause,
		Body:        []byte("raw body that should not appear in error string"),
		ContentType: "application/json",
	}
	msg := e.Error()
	assert.True(t, strings.Contains(msg, "application/json"))
	assert.True(t, strings.Contains(msg, "json: unexpected token"))
	assert.False(t, strings.Contains(msg, "raw body that should not appear"))
}

func TestStatusErrorUnwrap(t *testing.T) {
	t.Parallel()
	cause := errors.New("underlying")
	e := &httpc.StatusError{Status: 503, Cause: cause}
	assert.True(t, errors.Is(e, cause))
}

func TestBodyParseErrorUnwrap(t *testing.T) {
	t.Parallel()
	cause := errors.New("parse failure")
	e := &httpc.BodyParseError{Cause: cause}
	assert.True(t, errors.Is(e, cause))
}

func TestErrMaxAttemptsSentinel(t *testing.T) {
	t.Parallel()
	assert.True(t, errors.Is(httpc.ErrMaxAttempts, httpc.ErrMaxAttempts))
}

func TestErrMaxElapsedSentinel(t *testing.T) {
	t.Parallel()
	assert.True(t, errors.Is(httpc.ErrMaxElapsed, httpc.ErrMaxElapsed))
}

func TestErrBodyTooLarge(t *testing.T) {
	t.Parallel()
	assert.True(t, errors.Is(httpc.ErrBodyTooLarge, httpc.ErrBodyTooLarge))
}

func TestBodyParseErrorBodySnippet(t *testing.T) {
	t.Parallel()
	body := make([]byte, 2048)
	for i := range body {
		body[i] = 'x'
	}
	e := &httpc.BodyParseError{Cause: errors.New("err"), Body: body[:1024]}
	assert.Equal(t, 1024, len(e.Body))
}
