// SPDX-License-Identifier: Apache-2.0

package obs

import "go.opentelemetry.io/otel/attribute"

// Standard attribute keys for Glacier instrumentation.
var (
	// HTTP attributes.
	AttrHTTPMethod     = attribute.Key("http.method")
	AttrHTTPStatusCode = attribute.Key("http.status_code")
	AttrHTTPURL        = attribute.Key("http.url")
	AttrHTTPRoute      = attribute.Key("http.route")

	// CLI attributes.
	AttrCLICommand  = attribute.Key("cli.command")
	AttrCLIExitCode = attribute.Key("cli.exit_code")

	// Conf attributes.
	AttrConfSection = attribute.Key("conf.section")
	AttrConfSource  = attribute.Key("conf.source")

	// Common.
	AttrErrorType   = attribute.Key("error.type")
	AttrServiceName = attribute.Key("service.name")
)
