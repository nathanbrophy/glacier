---
slug: +glacier:env
title: "Marker: +glacier:env"
category: marker
see_also: ["+glacier:default"]
---
Binds an environment variable to a flag field. When the env var is set and the flag is not provided, the env var's value is used. The value must be UPPER_CASE_WITH_UNDERSCORES.

Example: // +glacier:env OTEL_EXPORTER_OTLP_ENDPOINT
