// SPDX-License-Identifier: Apache-2.0

package obs

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
)

// NumericType is the constraint for metric value types.
type NumericType interface {
	~int64 | ~float64
}

// Counter[T] is a monotonically increasing counter.
type Counter[T NumericType] struct {
	inner   otelmetric.Int64Counter
	innerF  otelmetric.Float64Counter
	isFloat bool
}

// NewCounter creates a Counter[T] using the named meter and instrument.
func NewCounter[T NumericType](meterName, counterName string, opts ...otelmetric.MeterOption) (*Counter[T], error) {
	meter := otel.GetMeterProvider().Meter(meterName, opts...)
	var zero T
	c := &Counter[T]{}
	switch any(zero).(type) {
	case float64:
		inner, err := meter.Float64Counter(counterName)
		if err != nil {
			return nil, err
		}
		c.innerF = inner
		c.isFloat = true
	default:
		inner, err := meter.Int64Counter(counterName)
		if err != nil {
			return nil, err
		}
		c.inner = inner
	}
	return c, nil
}

// Add increments the counter by v.
func (c *Counter[T]) Add(ctx context.Context, v T, attrs ...attribute.KeyValue) {
	if c == nil {
		return
	}
	if c.isFloat {
		c.innerF.Add(ctx, float64(v), otelmetric.WithAttributes(attrs...))
	} else {
		c.inner.Add(ctx, int64(v), otelmetric.WithAttributes(attrs...))
	}
}

// Histogram[T] records value distributions.
type Histogram[T NumericType] struct {
	inner   otelmetric.Int64Histogram
	innerF  otelmetric.Float64Histogram
	isFloat bool
}

// NewHistogram creates a Histogram[T] using the named meter and instrument.
func NewHistogram[T NumericType](meterName, histName string, opts ...otelmetric.MeterOption) (*Histogram[T], error) {
	meter := otel.GetMeterProvider().Meter(meterName, opts...)
	var zero T
	h := &Histogram[T]{}
	switch any(zero).(type) {
	case float64:
		inner, err := meter.Float64Histogram(histName)
		if err != nil {
			return nil, err
		}
		h.innerF = inner
		h.isFloat = true
	default:
		inner, err := meter.Int64Histogram(histName)
		if err != nil {
			return nil, err
		}
		h.inner = inner
	}
	return h, nil
}

// Record records a value in the histogram.
func (h *Histogram[T]) Record(ctx context.Context, v T, attrs ...attribute.KeyValue) {
	if h == nil {
		return
	}
	if h.isFloat {
		h.innerF.Record(ctx, float64(v), otelmetric.WithAttributes(attrs...))
	} else {
		h.inner.Record(ctx, int64(v), otelmetric.WithAttributes(attrs...))
	}
}

// Gauge[T] is a point-in-time measurement.
type Gauge[T NumericType] struct {
	inner   otelmetric.Int64Gauge
	innerF  otelmetric.Float64Gauge
	isFloat bool
}

// NewGauge creates a Gauge[T] using the named meter and instrument.
func NewGauge[T NumericType](meterName, gaugeName string, opts ...otelmetric.MeterOption) (*Gauge[T], error) {
	meter := otel.GetMeterProvider().Meter(meterName, opts...)
	var zero T
	g := &Gauge[T]{}
	switch any(zero).(type) {
	case float64:
		inner, err := meter.Float64Gauge(gaugeName)
		if err != nil {
			return nil, err
		}
		g.innerF = inner
		g.isFloat = true
	default:
		inner, err := meter.Int64Gauge(gaugeName)
		if err != nil {
			return nil, err
		}
		g.inner = inner
	}
	return g, nil
}

// Set records the current value of the gauge.
func (g *Gauge[T]) Set(ctx context.Context, v T, attrs ...attribute.KeyValue) {
	if g == nil {
		return
	}
	if g.isFloat {
		g.innerF.Record(ctx, float64(v), otelmetric.WithAttributes(attrs...))
	} else {
		g.inner.Record(ctx, int64(v), otelmetric.WithAttributes(attrs...))
	}
}
