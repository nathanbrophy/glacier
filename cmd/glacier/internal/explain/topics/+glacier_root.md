---
slug: +glacier:root
title: "Marker: +glacier:root"
category: marker
see_also: ["+glacier:command"]
---
Marks a command struct as the root of the CLI application. Exactly one +glacier:root must exist per binary. Glaciergen emits cli.WithRoot() for the registration call.
