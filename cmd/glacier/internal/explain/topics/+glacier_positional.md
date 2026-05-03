---
slug: +glacier:positional
title: "Marker: +glacier:positional"
category: marker
see_also: ["+glacier:required"]
---
Marks a field as a positional argument rather than a named flag. Positional arguments are read from os.Args in the order they appear on the command struct.

Note: full +glacier:positional support (Amendment E) is implemented in cli/gen; until a binary is regenerated the field reads from os.Args directly in Run().
