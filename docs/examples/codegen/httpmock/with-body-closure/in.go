//go:build glacier_codegen_fixture

// Package apiclient defines the HTTP transport for the dynamic API client.
package apiclient

import "net/http"

// DynamicTransport is an HTTP transport that generates response bodies per call.
//
// +glacier:mock
// +glacier:httpmock body=closure
type DynamicTransport interface {
	// RoundTrip executes an HTTP request.
	RoundTrip(req *http.Request) (*http.Response, error)
}