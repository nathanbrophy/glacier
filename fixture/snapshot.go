// SPDX-License-Identifier: Apache-2.0

package fixture

import (
	"bytes"
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/internal/safefile"
)

// Snapshot[T] serializes got to a deterministic human-readable text
// representation and compares it against the snapshot file at
// testdata/snapshots/<name>.snap. The comparison is delegated to
// assert.Equal[T] via re-deserialization, honoring the opts (IgnoreFields,
// IgnoreOrder, WithDelta, etc.). GLACIER_GOLDEN_UPDATE=1 creates or updates
// the snapshot file. Returns true on match, false on mismatch (with t.Errorf).
//
// The formatter guarantees: map keys are sorted, struct fields in declaration
// order, time.Time values in RFC 3339 date-only format, LF line endings.
func Snapshot[T any](t assert.TB, name string, got T, opts ...assert.EqualOption) bool {
	t.Helper()

	// Validate name for path safety.
	if _, err := safefile.Clean(name); err != nil {
		t.Errorf("fixture: Snapshot: path rejected for %q: %v", name, err)
		return false
	}

	// Serialize got to deterministic text.
	text, err := formatSnapshot(got)
	if err != nil {
		t.Errorf("fixture: Snapshot: format %T: %v", got, err)
		return false
	}

	// Resolve the snapshots subdirectory relative to the caller.
	root, resolveErr := resolveRoot(goldenConfig{}, 1)
	if resolveErr != nil {
		t.Errorf("fixture: Snapshot: %v", resolveErr)
		return false
	}
	snapshotRoot := filepath.Join(root, "snapshots")

	snapName := name + ".snap"
	display := "testdata/snapshots/" + snapName

	return goldenCompare(t, snapshotRoot, snapName, []byte(text), display)
}

// formatSnapshot serializes v into a deterministic human-readable string.
// Rules:
//   - time.Time → RFC3339 date-only (2006-01-02)
//   - map keys → sorted
//   - struct fields → declaration order
//   - line endings → LF only
//   - no trailing newline after the last field
func formatSnapshot(v any) (string, error) {
	var b bytes.Buffer
	if err := writeValue(&b, reflect.ValueOf(v), 0); err != nil {
		return "", err
	}
	// Normalize to LF line endings.
	result := strings.ReplaceAll(b.String(), "\r\n", "\n")
	return result, nil
}

func writeValue(b *bytes.Buffer, rv reflect.Value, depth int) error {
	if !rv.IsValid() {
		b.WriteString("<nil>")
		return nil
	}

	// Dereference pointers and interfaces.
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			b.WriteString("<nil>")
			return nil
		}
		rv = rv.Elem()
	}

	// Special case: time.Time → date-only RFC 3339.
	if rv.Type() == reflect.TypeOf(time.Time{}) {
		t := rv.Interface().(time.Time)
		fmt.Fprintf(b, "%q", t.UTC().Format("2006-01-02"))
		return nil
	}

	switch rv.Kind() {
	case reflect.Struct:
		return writeStruct(b, rv, depth)
	case reflect.Map:
		return writeMap(b, rv, depth)
	case reflect.Slice:
		return writeSlice(b, rv, depth)
	case reflect.Array:
		return writeArray(b, rv, depth)
	case reflect.String:
		fmt.Fprintf(b, "%q", rv.String())
	case reflect.Bool:
		if rv.Bool() {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Fprintf(b, "%d", rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		fmt.Fprintf(b, "%d", rv.Uint())
	case reflect.Float32, reflect.Float64:
		fmt.Fprintf(b, "%g", rv.Float())
	case reflect.Complex64, reflect.Complex128:
		fmt.Fprintf(b, "%g", rv.Complex())
	default:
		if rv.CanInterface() {
			fmt.Fprintf(b, "%v", rv.Interface())
		} else {
			b.WriteString("<unexported>")
		}
	}
	return nil
}

func writeStruct(b *bytes.Buffer, rv reflect.Value, depth int) error {
	t := rv.Type()
	b.WriteString(t.Name())
	b.WriteString("{\n")
	indent := strings.Repeat("  ", depth+1)
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		b.WriteString(indent)
		b.WriteString(f.Name)
		b.WriteString(": ")
		if err := writeValue(b, rv.Field(i), depth+1); err != nil {
			return err
		}
		b.WriteString("\n")
	}
	b.WriteString(strings.Repeat("  ", depth))
	b.WriteString("}")
	return nil
}

func writeMap(b *bytes.Buffer, rv reflect.Value, depth int) error {
	if rv.IsNil() {
		b.WriteString("<nil>")
		return nil
	}
	b.WriteString("{\n")
	indent := strings.Repeat("  ", depth+1)

	// Collect and sort keys.
	keys := rv.MapKeys()
	keyStrs := make([]string, len(keys))
	for i, k := range keys {
		keyStrs[i] = fmt.Sprintf("%v", k.Interface())
	}
	sort.Strings(keyStrs)

	// Build a map from stringified key to reflect.Value for lookup.
	keyMap := make(map[string]reflect.Value, len(keys))
	for _, k := range keys {
		ks := fmt.Sprintf("%v", k.Interface())
		keyMap[ks] = k
	}

	for _, ks := range keyStrs {
		k := keyMap[ks]
		b.WriteString(indent)
		fmt.Fprintf(b, "%q", ks)
		b.WriteString(": ")
		if err := writeValue(b, rv.MapIndex(k), depth+1); err != nil {
			return err
		}
		b.WriteString("\n")
	}
	b.WriteString(strings.Repeat("  ", depth))
	b.WriteString("}")
	return nil
}

func writeSlice(b *bytes.Buffer, rv reflect.Value, depth int) error {
	if rv.IsNil() {
		b.WriteString("<nil>")
		return nil
	}
	return writeSequence(b, rv, depth)
}

func writeArray(b *bytes.Buffer, rv reflect.Value, depth int) error {
	return writeSequence(b, rv, depth)
}

func writeSequence(b *bytes.Buffer, rv reflect.Value, depth int) error {
	b.WriteString("[\n")
	indent := strings.Repeat("  ", depth+1)
	for i := range rv.Len() {
		b.WriteString(indent)
		if err := writeValue(b, rv.Index(i), depth+1); err != nil {
			return err
		}
		b.WriteString("\n")
	}
	b.WriteString(strings.Repeat("  ", depth))
	b.WriteString("]")
	return nil
}
