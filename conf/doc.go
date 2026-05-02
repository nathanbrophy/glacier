// SPDX-License-Identifier: Apache-2.0

// Package conf is the Glacier answer to "I have a few config sources
// (defaults, JSON file, env vars, flags) and a few struct-shaped config values
// I want them to populate." Two complementary APIs: a registration pattern
// where packages declare typed config structs with sensible defaults and
// conf.Load walks every registered struct and populates from layered sources;
// and a one-shot Decode[T] for programs with a single root config struct. JSON
// struct tags drive field naming (consumers reuse the tag they already know).
// v0 reads JSON files only (Falcon's ruling); env vars parsed inline; flag
// sources via the FlagSource interface that cli later implements. Validation is
// separate from decode — typo'd JSON or missing required env values surface as
// DecodeErrors during Load; semantic invariants are checked via option.Validate
// after Load returns successfully.
package conf
