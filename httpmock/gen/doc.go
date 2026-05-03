// SPDX-License-Identifier: Apache-2.0

// Package gen is the codegen library for the Glacier httpmock package. It
// provides a Generate entry point with the same Options shape as cli/gen and
// mock/gen so that the glacier SDK's generate command can invoke all three
// generators uniformly. In v0, httpmock defines no +glacier:httpmock source
// markers, so Generate is a well-formed no-op: it loads the requested packages,
// confirms they compile, logs a discovery summary, and returns nil. The package
// exists to hold the place in the generator registry and to be wired as a real
// function reference per spec 0032 D-S65. Full context: specs/0013-httpmock.md
// and specs/0032-sdk.md.
package gen
