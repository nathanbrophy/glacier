// SPDX-License-Identifier: Apache-2.0

// Package sigh wires OS signals into the context lifecycle. Its primary export
// is Notify, which returns a derived context that cancels when SIGINT or
// SIGTERM is received. The cli package's App.Main and any future sandboxed
// runner use this to give user handlers a ctx that cancels on Ctrl-C
// automatically, with no bespoke signal-handling code at the call site.
// The package is internal to the Glacier module and may not be imported
// outside it. It does not import cli/ (F6 forbidden edge).
package sigh
