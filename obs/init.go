// SPDX-License-Identifier: Apache-2.0

package obs

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	glacierErrs "github.com/nathanbrophy/glacier/errs"
	"github.com/nathanbrophy/glacier/option"
)

// Provider wraps OTEL SDK providers and their shared lifecycle.
type Provider struct {
	tracer *sdktrace.TracerProvider
	meter  *sdkmetric.MeterProvider
	once   sync.Once
}

// Default is the package-level shared provider. nil until Init is called.
var Default *Provider

// initConfig holds construction-time settings for Init.
type initConfig struct {
	endpoint        string
	sampler         sdktrace.Sampler
	resAttrs        []attribute.KeyValue
	metricsInterval time.Duration
	logger          *slog.Logger
	insecure        bool
}

// Init creates a Provider with an OTLP gRPC exporter and sets obs.Default.
// If the OTEL endpoint is not configured (env or WithEndpoint), Init returns
// a no-op Provider (zero overhead).
//
// Init is idempotent: the second call is a no-op if Default is already set.
func Init(ctx context.Context, opts ...option.Option[initConfig]) (*Provider, error) {
	cfg, _ := option.Apply(opts)
	if cfg.logger == nil {
		cfg.logger = slog.Default()
	}
	if cfg.metricsInterval == 0 {
		cfg.metricsInterval = 30 * time.Second
	}
	if cfg.sampler == nil {
		cfg.sampler = sdktrace.AlwaysSample()
	}

	endpoint := cfg.endpoint
	if endpoint == "" {
		endpoint = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	}

	if endpoint == "" {
		// No endpoint configured: return no-op provider (zero overhead).
		p := &Provider{}
		if Default == nil {
			Default = p
		}
		return p, nil
	}

	traceExpOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint),
	}
	metricExpOpts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(endpoint),
	}
	if cfg.insecure {
		traceExpOpts = append(traceExpOpts, otlptracegrpc.WithInsecure())
		metricExpOpts = append(metricExpOpts, otlpmetricgrpc.WithInsecure())
	}

	traceExp, err := otlptracegrpc.New(ctx, traceExpOpts...)
	if err != nil {
		return nil, glacierErrs.Wrap(err, "obs: init: trace exporter")
	}

	res, _ := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(cfg.resAttrs...),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithSampler(cfg.sampler),
		sdktrace.WithResource(res),
	)

	metricExp, err := otlpmetricgrpc.New(ctx, metricExpOpts...)
	if err != nil {
		_ = tp.Shutdown(ctx)
		return nil, glacierErrs.Wrap(err, "obs: init: metrics exporter")
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp,
			sdkmetric.WithInterval(cfg.metricsInterval))),
		sdkmetric.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)

	p := &Provider{tracer: tp, meter: mp}
	if Default == nil {
		Default = p
	}
	return p, nil
}

// WithEndpoint sets the OTLP gRPC endpoint (e.g., "localhost:4317").
func WithEndpoint(endpoint string) option.Option[initConfig] {
	return option.OptionFunc[initConfig](func(c *initConfig) error {
		c.endpoint = endpoint
		return nil
	})
}

// WithInsecure disables TLS for the OTLP gRPC connection (for dev/test).
func WithInsecure() option.Option[initConfig] {
	return option.OptionFunc[initConfig](func(c *initConfig) error {
		c.insecure = true
		return nil
	})
}

// WithSampler sets the trace sampling strategy.
func WithSampler(s sdktrace.Sampler) option.Option[initConfig] {
	return option.OptionFunc[initConfig](func(c *initConfig) error {
		c.sampler = s
		return nil
	})
}

// WithMetricsInterval sets how often metrics are exported.
func WithMetricsInterval(d time.Duration) option.Option[initConfig] {
	return option.OptionFunc[initConfig](func(c *initConfig) error {
		c.metricsInterval = d
		return nil
	})
}

// WithLogger sets the slog.Logger for internal obs operational events.
func WithLogger(l *slog.Logger) option.Option[initConfig] {
	return option.OptionFunc[initConfig](func(c *initConfig) error {
		c.logger = l
		return nil
	})
}

// WithResourceAttributes adds key-value pairs to the OTEL resource description.
func WithResourceAttributes(attrs ...attribute.KeyValue) option.Option[initConfig] {
	return option.OptionFunc[initConfig](func(c *initConfig) error {
		c.resAttrs = append(c.resAttrs, attrs...)
		return nil
	})
}
