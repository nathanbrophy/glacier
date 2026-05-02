// SPDX-License-Identifier: Apache-2.0

// Package term provides Glacier's terminal-interaction layer: TTY capability
// detection, ANSI color and style primitives (with NO_COLOR and
// GLACIER_NO_COLOR support), Unicode glyph registry with ASCII fallback,
// beauty-writer helpers (box layout, alignment, text wrapping, column
// formatting, banner rendering), interactive prompts (text, password, confirm,
// select, multi-select), and the Animator — a coordinating frame-loop that
// renders progress bars and spinners while buffering log output above the
// animation so log lines never tear the display. The package replaces
// internal/ttyx and is Tier 0 kernel, importable by every Glacier package that
// needs terminal capability or styled output.
package term
