// SPDX-License-Identifier: Apache-2.0

// Package httpc is a first-class HTTP client wrapper around stdlib *http.Client.
// It eliminates the boilerplate of "make request → read body → unmarshal →
// check error → retry on 5xx" that every Go program writes by hand. Headline
// ergonomics: typed responses via generics (httpc.Get[User](ctx, url) reads,
// unmarshals, and returns a User); closure-generated request bodies for safe
// retry (the closure is invoked fresh for every attempt); built-in retry
// policies plus a RetryIf escape hatch; and ctx-propagated dry-run
// (httpc.WithDryRun flips the entire pipeline to emit RequestPlan audit events
// with no network calls). Distinct from httpmock: httpc is production HTTP
// client code; httpmock is the testing transport. Full API in
// specs/0015-httpc.md.
package httpc
