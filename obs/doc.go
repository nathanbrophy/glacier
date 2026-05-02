// SPDX-License-Identifier: Apache-2.0

// Package obs provides Glacier's OpenTelemetry-based observability — metrics
// and traces. Logs stay in log/. Initialize via obs.Init, which configures a
// shared Provider holding a MeterProvider and TracerProvider with Glacier
// defaults (resource attributes from build info, configurable sampler and
// exporter). Typed counter, histogram, and gauge constructors use Go generics
// over int64 and float64 (the OTEL instrument types). Span helpers wrap
// go.opentelemetry.io/otel/trace with Glacier-flavored ergonomics and ctx
// propagation. The log/ bridge automatically appends trace_id and span_id
// attributes to log records when an active span is in ctx. Every other
// framework package gains opt-in instrumentation via WithTracing() and
// WithMetrics() options; when unset, overhead is zero. Full API in
// specs/0017-obs.md.
package obs
