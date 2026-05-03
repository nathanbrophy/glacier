//go:build glacier_codegen_fixture

// Package apiclient defines the HTTP transport for the strict API client.
package apiclient

import "net/http"

// StrictTransport is an HTTP transport that must not make unregistered requests.
//
// +glacier:mock
// +glacier:httpmock recording=disabled
type StrictTransport interface {
	// RoundTrip executes an HTTP request.
	RoundTrip(req *http.Request) (*http.Response, error)
}