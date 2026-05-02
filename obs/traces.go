// SPDX-License-Identifier: Apache-2.0

package obs

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/nathanbrophy/glacier/option"
)

// SpanKind wraps OTEL trace span kinds.
type SpanKind = oteltrace.SpanKind

// Span kind constants re-exported for convenience.
const (
	SpanKindInternal = oteltrace.SpanKindInternal
	SpanKindServer   = oteltrace.SpanKindServer
	SpanKindClient   = oteltrace.SpanKindClient
	SpanKindProducer = oteltrace.SpanKindProducer
	SpanKindConsumer = oteltrace.SpanKindConsumer
)

// Span wraps an OTEL span with Glacier-flavored helpers.
type Span struct {
	inner oteltrace.Span
}

// spanConfig holds StartSpan options.
type spanConfig struct {
	kind  oteltrace.SpanKind
	attrs []attribute.KeyValue
}

// WithSpanKind sets the span kind.
func WithSpanKind(k oteltrace.SpanKind) option.Option[spanConfig] {
	return option.OptionFunc[spanConfig](func(c *spanConfig) error {
		c.kind = k
		return nil
	})
}

// WithSpanAttributes adds attributes to a span at start time.
func WithSpanAttributes(attrs ...attribute.KeyValue) option.Option[spanConfig] {
	return option.OptionFunc[spanConfig](func(c *spanConfig) error {
		c.attrs = append(c.attrs, attrs...)
		return nil
	})
}

// StartSpan begins a new span and returns a derived context and the Span.
//
// The caller must call span.End() when done, typically via defer.
func StartSpan(ctx context.Context, tracerName, spanName string, opts ...option.Option[spanConfig]) (context.Context, *Span) {
	cfg, _ := option.Apply(opts)
	tracer := otel.GetTracerProvider().Tracer(tracerName)
	ctx, inner := tracer.Start(ctx, spanName,
		oteltrace.WithSpanKind(cfg.kind),
		oteltrace.WithAttributes(cfg.attrs...),
	)
	return ctx, &Span{inner: inner}
}

// End ends the span.
func (s *Span) End() {
	if s != nil && s.inner != nil {
		s.inner.End()
	}
}

// RecordError records err as an exception on the span.
func (s *Span) RecordError(err error) {
	if s != nil && s.inner != nil && err != nil {
		s.inner.RecordError(err)
	}
}

// SetStatus sets the span's status code and description.
func (s *Span) SetStatus(code codes.Code, desc string) {
	if s != nil && s.inner != nil {
		s.inner.SetStatus(code, desc)
	}
}

// AddEvent adds a named event with optional attributes.
func (s *Span) AddEvent(name string, attrs ...attribute.KeyValue) {
	if s != nil && s.inner != nil {
		s.inner.AddEvent(name, oteltrace.WithAttributes(attrs...))
	}
}

// SetAttribute sets an attribute on the span.
func (s *Span) SetAttribute(k attribute.Key, v attribute.Value) {
	if s != nil && s.inner != nil {
		s.inner.SetAttributes(attribute.KeyValue{Key: k, Value: v})
	}
}

// SpanFromContext returns the current span from ctx, or a no-op span.
func SpanFromContext(ctx context.Context) *Span {
	return &Span{inner: oteltrace.SpanFromContext(ctx)}
}

// TraceIDFromContext returns the hex trace ID from the current span in ctx,
// or "" if no span is active.
func TraceIDFromContext(ctx context.Context) string {
	sc := oteltrace.SpanFromContext(ctx).SpanContext()
	if !sc.IsValid() {
		return ""
	}
	return sc.TraceID().String()
}

// SpanIDFromContext returns the hex span ID from the current span in ctx,
// or "" if no span is active.
func SpanIDFromContext(ctx context.Context) string {
	sc := oteltrace.SpanFromContext(ctx).SpanContext()
	if !sc.IsValid() {
		return ""
	}
	return sc.SpanID().String()
}
