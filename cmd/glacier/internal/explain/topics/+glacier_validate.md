---
slug: +glacier:validate
title: "Marker: +glacier:validate"
category: marker
see_also: ["+glacier:choices"]
---
Wires a custom validator function to a flag field. The value is the name of a func(string) error in the same package. Called after flag parsing, before Run().

Example: // +glacier:validate validateGeneratorList
