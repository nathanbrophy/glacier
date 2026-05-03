// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"
	"unicode"
)

// fieldFlagName converts a Go field name (CamelCase or PascalCase) to its
// command-line flag form (kebab-case). The conversion inserts a hyphen at
// every lower-to-upper or acronym-to-word boundary, then lowercases the
// result, matching how the spec and docs render flag names.
//
//	NoColor       → "no-color"
//	VeryVerbose   → "very-verbose"
//	OtelEndpoint  → "otel-endpoint"
//	JSON          → "json"
//	JSONOnly      → "json-only"
//	URL           → "url"
//	URLPath       → "url-path"
//
// Single-word fields (Verbose, Quiet, Profile) round-trip to their plain
// lowercase form so existing flag references continue to work.
func fieldFlagName(field string) string {
	var b strings.Builder
	b.Grow(len(field) + 4)
	runes := []rune(field)
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) {
			prev := runes[i-1]
			lowerToUpper := unicode.IsLower(prev) || unicode.IsDigit(prev)
			acronymBoundary := unicode.IsUpper(prev) && i+1 < len(runes) && unicode.IsLower(runes[i+1])
			if lowerToUpper || acronymBoundary {
				b.WriteRune('-')
			}
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// buildFlagSet constructs a flag.FlagSet from the exported fields of cmd's
// concrete struct type. cfg provides short aliases, env overrides, and help
// text. The returned map associates lowercase flag names to their struct
// field names for lookup after parsing.
func buildFlagSet(cmd any, cfg regConfig, cmdPath string) (*flag.FlagSet, map[string]string) {
	fs := flag.NewFlagSet(cmdPath, flag.ContinueOnError)
	fieldByFlag := make(map[string]string) // flagName → fieldName

	rv := reflect.ValueOf(cmd)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return fs, fieldByFlag
	}
	rt := rv.Type()

	for i := range rt.NumField() {
		f := rt.Field(i)
		if !f.IsExported() {
			continue
		}
		flagName := fieldFlagName(f.Name)
		helpText := flagName
		if h, ok := cfg.help[f.Name]; ok {
			helpText = h
		}
		fieldByFlag[flagName] = f.Name

		fv := rv.Field(i)
		switch f.Type.Kind() {
		case reflect.String:
			ptr := (*string)(fv.Addr().UnsafePointer())
			fs.StringVar(ptr, flagName, fv.String(), helpText)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if f.Type == reflect.TypeOf(time.Duration(0)) {
				ptr := (*time.Duration)(fv.Addr().UnsafePointer())
				fs.DurationVar(ptr, flagName, time.Duration(fv.Int()), helpText)
			} else {
				ptr := (*int)(fv.Addr().UnsafePointer())
				fs.IntVar(ptr, flagName, int(fv.Int()), helpText)
			}
		case reflect.Bool:
			ptr := (*bool)(fv.Addr().UnsafePointer())
			fs.BoolVar(ptr, flagName, fv.Bool(), helpText)
		case reflect.Float64:
			ptr := (*float64)(fv.Addr().UnsafePointer())
			fs.Float64Var(ptr, flagName, fv.Float(), helpText)
		default:
			// unsupported type :  skip silently
			delete(fieldByFlag, flagName)
		}
	}

	return fs, fieldByFlag
}

// applyEnvOverrides applies environment variable defaults to any flag not
// explicitly set on the FlagSet. It must be called after fs.Parse().
func applyEnvOverrides(fs *flag.FlagSet, cfg regConfig) {
	// Track which flags were explicitly set.
	explicit := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		explicit[f.Name] = true
	})

	for flagName, envVar := range cfg.envVars {
		if explicit[fieldFlagName(flagName)] {
			continue // argv wins
		}
		val := os.Getenv(envVar)
		if val == "" {
			continue
		}
		lower := fieldFlagName(flagName)
		f := fs.Lookup(lower)
		if f == nil {
			continue
		}
		if err := f.Value.Set(val); err != nil {
			// Best-effort: if env var is malformed, skip.
			_ = err
		}
	}
}

// flagValues collects all flag values from fs as a string map (flagName → string value).
func flagValues(fs *flag.FlagSet) map[string]string {
	vals := make(map[string]string)
	fs.VisitAll(func(f *flag.Flag) {
		vals[f.Name] = f.Value.String()
	})
	return vals
}

// validateRequiredFlags checks that all required flags have a non-zero value.
func validateRequiredFlags(fs *flag.FlagSet, cfg regConfig) error {
	// Collect which flags were set from any source (argv or env).
	set := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		set[f.Name] = true
	})
	// Also consider flags whose value differs from default (set by env overlay).
	fs.VisitAll(func(f *flag.Flag) {
		if f.Value.String() != f.DefValue {
			set[f.Name] = true
		}
	})

	for flagName := range cfg.required {
		lower := fieldFlagName(flagName)
		f := fs.Lookup(lower)
		if f == nil {
			return &RequiredError{Name: lower}
		}
		if !set[lower] {
			return &RequiredError{Name: lower}
		}
	}
	return nil
}

// validateChoices checks that all flags with choices constraints have an
// in-set value.
func validateChoices(fs *flag.FlagSet, cfg regConfig) error {
	for flagName, choices := range cfg.choices {
		lower := fieldFlagName(flagName)
		f := fs.Lookup(lower)
		if f == nil {
			continue
		}
		val := f.Value.String()
		if val == "" {
			continue // not set; required check handles missing
		}
		valid := false
		for _, ch := range choices {
			if ch == val {
				valid = true
				break
			}
		}
		if !valid {
			return &ChoicesError{Name: lower, Value: val, Choices: choices}
		}
	}
	return nil
}

// validateFns runs per-flag validate functions.
func validateFns(fs *flag.FlagSet, cfg regConfig) error {
	for flagName, fn := range cfg.validate {
		lower := fieldFlagName(flagName)
		f := fs.Lookup(lower)
		if f == nil {
			continue
		}
		if err := fn(f.Value.String()); err != nil {
			return fmt.Errorf("cli: validate flag %q: %w", lower, err)
		}
	}
	return nil
}

// stripANSI removes ANSI escape sequences from s.
func stripANSI(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			// Skip until 'm' or end.
			j := i + 2
			for j < len(s) && s[j] != 'm' {
				j++
			}
			if j < len(s) {
				j++ // skip 'm'
			}
			i = j
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}
