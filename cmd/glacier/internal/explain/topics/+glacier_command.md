---
slug: +glacier:command
title: "Marker: +glacier:command"
category: marker
see_also: ["+glacier:root", "+glacier:mock"]
---
Annotates a struct as a CLI command. Glaciergen reads this marker and registers the struct in the generated command tree. Required attributes: name=<slug>. Optional: parent=<path>, alias=<name>, app=<var>.

The struct must implement Run(ctx context.Context) error.
