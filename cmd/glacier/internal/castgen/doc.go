// SPDX-License-Identifier: Apache-2.0

// Package castgen records terminal sessions of the SDK binary into
// asciinema v2 .cast files and renders a static .svg snapshot of the
// final frame. Both artifacts land under site/public/casts/.
//
// The package exists so the public site can ship colored, accurate
// reproductions of `glacier --help`, `glacier vibe`, etc. without a
// runtime dependency on asciinema or agg.
//
// Usage:
//
//	go run ./cmd/glacier/internal/castgen
//
// Each scenario re-runs the live binary with FORCE_COLOR=1 so the cast
// captures real ANSI output, then writes site/public/casts/<name>.cast
// and site/public/casts/<name>.svg.
package castgen
