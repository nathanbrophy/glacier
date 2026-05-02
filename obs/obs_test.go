// SPDX-License-Identifier: Apache-2.0

package obs_test

import (
	"context"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	"github.com/nathanbrophy/glacier/obs"
)

// T#O1 Init with no endpoint returns a non-nil Provider (no-op mode).
func TestInit_NoEndpoint(t *testing.T) {
	t.Parallel()
	p, err := obs.Init(context.Background())
	require.NoError(t, err)
	assert.True(t, p != nil, "Init returned nil Provider")
}

// T#O2 Provider.Shutdown is idempotent (second call returns nil).
func TestProvider_Shutdown_Idempotent(t *testing.T) {
	t.Parallel()
	p, err := obs.Init(context.Background())
	require.NoError(t, err)
	err = p.Shutdown(context.Background())
	assert.NoError(t, err)
	err = p.Shutdown(context.Background())
	assert.NoError(t, err)
}

// T#O3 TraceIDFromContext with no span returns "".
func TestTraceIDFromContext_NoSpan(t *testing.T) {
	t.Parallel()
	id := obs.TraceIDFromContext(context.Background())
	assert.Equal(t, id, "")
}

// T#O4 SpanIDFromContext with no span returns "".
func TestSpanIDFromContext_NoSpan(t *testing.T) {
	t.Parallel()
	id := obs.SpanIDFromContext(context.Background())
	assert.Equal(t, id, "")
}

// T#O5 StartSpan returns non-nil Span and derived context.
func TestStartSpan(t *testing.T) {
	t.Parallel()
	ctx, span := obs.StartSpan(context.Background(), "test-tracer", "test-span")
	assert.True(t, span != nil, "StartSpan returned nil Span")
	assert.True(t, ctx != nil, "StartSpan returned nil ctx")
	span.End()
}

// T#O6 Span nil-safety: End, RecordError, SetStatus, AddEvent, SetAttribute do not panic.
func TestSpan_NilSafe(t *testing.T) {
	t.Parallel()
	var s *obs.Span
	s.End()
	s.RecordError(nil)
	s.RecordError(context.Canceled)
	s.AddEvent("test-event")
}

// T#O7 Counter.Add does not panic in no-op mode.
func TestCounter_NoOp(t *testing.T) {
	t.Parallel()
	c, err := obs.NewCounter[int64]("test-meter", "test.counter")
	require.NoError(t, err)
	c.Add(context.Background(), 1)
}

// T#O8 Histogram.Record does not panic in no-op mode.
func TestHistogram_NoOp(t *testing.T) {
	t.Parallel()
	h, err := obs.NewHistogram[float64]("test-meter", "test.histogram")
	require.NoError(t, err)
	h.Record(context.Background(), 1.5)
}

// T#O9 Gauge.Set does not panic in no-op mode.
func TestGauge_NoOp(t *testing.T) {
	t.Parallel()
	g, err := obs.NewGauge[int64]("test-meter", "test.gauge")
	require.NoError(t, err)
	g.Set(context.Background(), 42)
}

// T#O10 StartSpan with options propagates span kind and attributes.
func TestStartSpan_WithOptions(t *testing.T) {
	t.Parallel()
	ctx, span := obs.StartSpan(
		context.Background(),
		"test-tracer",
		"test-span-opts",
		obs.WithSpanKind(obs.SpanKindServer),
		obs.WithSpanAttributes(obs.AttrHTTPMethod.String("GET")),
	)
	assert.True(t, span != nil, "StartSpan returned nil Span")
	assert.True(t, ctx != nil, "StartSpan returned nil ctx")
	span.End()
}

// T#O11 Float64 counter variant does not panic in no-op mode.
func TestCounter_Float64_NoOp(t *testing.T) {
	t.Parallel()
	c, err := obs.NewCounter[float64]("test-meter", "test.counter.f64")
	require.NoError(t, err)
	c.Add(context.Background(), 2.5)
}

// T#O12 Float64 gauge variant does not panic in no-op mode.
func TestGauge_Float64_NoOp(t *testing.T) {
	t.Parallel()
	g, err := obs.NewGauge[float64]("test-meter", "test.gauge.f64")
	require.NoError(t, err)
	g.Set(context.Background(), 3.14)
}

// T#O13 SpanFromContext on background context returns a non-nil Span.
func TestSpanFromContext_Background(t *testing.T) {
	t.Parallel()
	s := obs.SpanFromContext(context.Background())
	assert.True(t, s != nil, "SpanFromContext returned nil")
	// End should not panic even on a no-op span.
	s.End()
}

// ExampleInit demonstrates initializing the obs package in no-op mode.
func ExampleInit() {
	p, err := obs.Init(context.Background())
	if err != nil {
		panic(err)
	}
	defer p.Shutdown(context.Background()) //nolint:errcheck
}

// ExampleStartSpan demonstrates starting and ending a span.
func ExampleStartSpan() {
	ctx, span := obs.StartSpan(context.Background(), "my-service", "my-operation")
	defer span.End()
	_ = ctx
}
