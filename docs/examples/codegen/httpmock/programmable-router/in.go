//go:build glacier_codegen_fixture

// Package apiclient defines the HTTP transport for the API client.
package apiclient

import "net/http"

// Transport is the HTTP transport interface used by the API client.
//
// +glacier:mock
type Transport interface {
	// RoundTrip executes an HTTP request.
	RoundTrip(req *http.Request) (*http.Response, error)
}