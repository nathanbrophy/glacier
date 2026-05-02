// SPDX-License-Identifier: Apache-2.0

package conf

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"unicode"

	safejson "github.com/nathanbrophy/glacier/internal/safejson"
)

// buildMerged constructs the fully resolved map[string]any for a single
// section by applying all configured sources in precedence order:
//
//	struct defaults → WithDefaults fns → JSON file → env vars → flags → WithSet
//
// sectionPath is the dot-separated key used to extract the relevant subtree
// from the JSON file and to scope env-var and flag lookups. Pass "" to use
// the entire document.
func buildMerged(cfg loadConfig, sectionPath string, defaults any) (map[string]any, error) {
	// Layer 1: struct defaults — marshal the zero/defaults value into a map.
	merged := map[string]any{}
	if defaults != nil {
		data, _ := json.Marshal(defaults)
		_ = json.Unmarshal(data, &merged)
	}

	// Layer 2: WithDefaults functions.
	for _, fn := range cfg.defaultsFns {
		for k, v := range fn() {
			rel := relativeKey(k, sectionPath)
			if rel == "" && sectionPath != "" {
				continue
			}
			key := rel
			if sectionPath == "" {
				key = k
			}
			setNestedKey(merged, key, v)
		}
	}

	// Layer 3: JSON file.
	if cfg.filePath != "" {
		fileMap, err := loadJSONFile(cfg.filePath, sectionPath)
		if err != nil {
			return nil, err
		}
		mergeMaps(merged, fileMap)
	}

	// Layer 4: Environment variables.
	if cfg.envPrefix != "" {
		sep := cfg.envSliceSep
		if sep == "" {
			sep = ","
		}
		envMap := loadEnvVars(cfg.envPrefix, sectionPath, sep)
		mergeMaps(merged, envMap)
	}

	// Layer 5 and 6 (flags and WithSet) are applied by the caller after
	// buildMerged because flag lookup requires the concrete struct type for
	// field enumeration, and WithSet is applied in the calling decode path.

	return merged, nil
}

// applyWithSet merges WithSet entries into merged for the given sectionPath.
func applyWithSet(merged map[string]any, sets map[string]any, sectionPath string) {
	for k, v := range sets {
		rel := relativeKey(k, sectionPath)
		if rel == "" && sectionPath != "" {
			continue
		}
		key := rel
		if sectionPath == "" {
			key = k
		}
		setNestedKey(merged, key, v)
	}
}

// relativeKey strips the sectionPath prefix from k and returns the remainder.
// Returns "" when k does not belong to sectionPath (and sectionPath is non-empty).
func relativeKey(k, sectionPath string) string {
	if sectionPath == "" {
		return k
	}
	prefix := sectionPath + "."
	if strings.HasPrefix(k, prefix) {
		return k[len(prefix):]
	}
	return ""
}

// setNestedKey sets a dot-separated key in a nested map[string]any, creating
// intermediate maps as needed.
func setNestedKey(m map[string]any, key string, value any) {
	parts := strings.SplitN(key, ".", 2)
	if len(parts) == 1 {
		m[key] = value
		return
	}
	sub, ok := m[parts[0]].(map[string]any)
	if !ok {
		sub = map[string]any{}
		m[parts[0]] = sub
	}
	setNestedKey(sub, parts[1], value)
}

// mergeMaps merges src into dst. For nested maps both sides own, it recurses;
// otherwise src wins (src is the higher-priority layer).
func mergeMaps(dst, src map[string]any) {
	for k, v := range src {
		if srcSub, ok := v.(map[string]any); ok {
			if dstSub, ok2 := dst[k].(map[string]any); ok2 {
				mergeMaps(dstSub, srcSub)
				continue
			}
		}
		dst[k] = v
	}
}

// loadJSONFile opens path, decodes it as JSON, and returns the subtree
// at sectionPath. When sectionPath is "", the entire document is returned.
func loadJSONFile(path, sectionPath string) (map[string]any, error) {
	f, err := os.Open(path) //nolint:gosec // path comes from the caller's WithFile option
	if err != nil {
		return nil, &DecodeError{Cause: err, Layer: "file"}
	}
	defer f.Close()

	var full map[string]any
	if err := safejson.Decode(f, &full); err != nil {
		switch err {
		case safejson.ErrTooLarge:
			return nil, &DecodeError{Cause: ErrFileTooLarge, Layer: "file"}
		case safejson.ErrDepthExceeded:
			return nil, &DecodeError{Cause: ErrDepthExceeded, Layer: "file"}
		default:
			return nil, &DecodeError{Cause: err, Layer: "file"}
		}
	}

	if sectionPath == "" {
		return full, nil
	}

	// Navigate the dot-separated path to the requested subtree.
	var cur any = full
	for _, p := range strings.Split(sectionPath, ".") {
		m, ok := cur.(map[string]any)
		if !ok {
			return map[string]any{}, nil
		}
		cur = m[p]
	}
	if cur == nil {
		return map[string]any{}, nil
	}
	if m, ok := cur.(map[string]any); ok {
		return m, nil
	}
	return map[string]any{}, nil
}

// loadEnvVars reads environment variables whose names begin with
// <PREFIX>__<SECTION>__ and returns them as a nested map. Keys within the
// returned map use the lowercased field name (with __ translated to dot
// separators for nested keys).
func loadEnvVars(prefix, sectionPath, _ string) map[string]any {
	result := map[string]any{}

	// Build the expected env-var prefix for this section.
	envBase := strings.ToUpper(prefix)
	if sectionPath != "" {
		envBase += "__" + strings.ToUpper(strings.ReplaceAll(sectionPath, ".", "__"))
	}
	envBase += "__"

	for _, env := range os.Environ() {
		idx := strings.IndexByte(env, '=')
		if idx < 0 {
			continue
		}
		key, val := env[:idx], env[idx+1:]
		if !strings.HasPrefix(strings.ToUpper(key), envBase) {
			continue
		}
		// Strip the base prefix (case-insensitive match, use original key length).
		suffix := key[len(envBase):]
		// Translate UPPER__SNAKE to lower.dot for nested keys.
		fieldKey := strings.ToLower(strings.ReplaceAll(suffix, "__", "."))

		// Attempt JSON parse so numbers, booleans, and arrays round-trip correctly.
		var parsed any
		if err := json.Unmarshal([]byte(val), &parsed); err != nil {
			// Not valid JSON — treat as plain string.
			parsed = val
		}
		setNestedKey(result, fieldKey, parsed)
	}
	return result
}

// applyFlagSourceToMerged iterates over the exported struct fields of t and
// queries the FlagSource for each field path. Values are merged into merged at
// the appropriate key.
func applyFlagSourceToMerged(fs FlagSource, sectionPath string, t reflect.Type, merged map[string]any) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		jsonKey := fieldJSONKey(f)
		dotPath := jsonKey
		if sectionPath != "" {
			dotPath = sectionPath + "." + jsonKey
		}
		if val, ok := fs.Lookup(dotPath); ok {
			var parsed any
			if err := json.Unmarshal([]byte(val), &parsed); err != nil {
				parsed = val
			}
			merged[jsonKey] = parsed
		}
		// Recurse into embedded or nested struct fields.
		ft := f.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		if ft.Kind() == reflect.Struct {
			subMerged, ok := merged[jsonKey].(map[string]any)
			if !ok {
				subMerged = map[string]any{}
				merged[jsonKey] = subMerged
			}
			applyFlagSourceToMerged(fs, dotPath, ft, subMerged)
		}
	}
}

// fieldJSONKey returns the JSON key for a struct field: the name from the
// json struct tag when present, otherwise the UPPER_SNAKE_CASE of the Go field name.
func fieldJSONKey(f reflect.StructField) string {
	if tag := f.Tag.Get("json"); tag != "" {
		name := strings.Split(tag, ",")[0]
		if name != "" && name != "-" {
			return name
		}
	}
	return toUpperSnake(f.Name)
}

// toUpperSnake converts a CamelCase Go identifier to UPPER_SNAKE_CASE.
func toUpperSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			b.WriteByte('_')
		}
		b.WriteRune(unicode.ToUpper(r))
	}
	return b.String()
}
