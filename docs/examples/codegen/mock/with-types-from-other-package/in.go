//go:build glacier_codegen_fixture

// Package client defines the HTTP transport interface for the application.
package client

import (
	"context"
	"net/http"
)

// HTTPClient performs HTTP requests.
//
// +glacier:mock
type HTTPClient interface {
	// Do sends an HTTP request and returns an HTTP response.
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}