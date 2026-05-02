module github.com/nathanbrophy/glacier

go 1.25.0

toolchain go1.26.0

require (
	golang.org/x/sys v0.29.0                                                   // term/ — raw-mode + Windows console; pinned per §23.3
	golang.org/x/tools v0.29.0                                                  // cli/gen — build-time only via go/packages; pinned per §23.3
	go.opentelemetry.io/otel v1.34.0                                            // obs/ — OTEL standards observability per §23.3
	go.opentelemetry.io/otel/metric v1.34.0                                     // obs/ — OTEL metric instruments
	go.opentelemetry.io/otel/trace v1.34.0                                      // obs/ — OTEL trace types
	go.opentelemetry.io/otel/sdk v1.34.0                                        // obs/ — OTEL SDK providers
	go.opentelemetry.io/otel/sdk/metric v1.34.0                                 // obs/ — OTEL metric SDK
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.34.0  // obs/ — gRPC metric exporter
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.34.0    // obs/ — gRPC trace exporter
)
