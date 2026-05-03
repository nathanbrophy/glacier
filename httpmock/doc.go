// SPDX-License-Identifier: Apache-2.0

// Package httpmock provides a programmable http.RoundTripper for testing code
// that uses *http.Client. The transport never makes real network calls :  it is
// purely an in-memory mock that intercepts every request and returns a scripted
// response. Stubs are declared via a chained builder (Method → Path → Header →
// Respond), typed responses use generic JSON[T any] for compile-time-safe
// response bodies, and the transport records every request for post-test
// assertion. Strict by default: unmatched requests return ErrNoRouteMatch.
// Scope is deliberately tight: replay only, no recording, no real-network
// proxying. Full API in specs/0013-httpmock.md.
package httpmock
