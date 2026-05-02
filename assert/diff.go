// SPDX-License-Identifier: Apache-2.0

package assert

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/nathanbrophy/glacier/term"
)

// renderDiff produces a human-readable diff between got and want for use in
// failure messages. It uses +/- markers and, when writing to a TTY, applies
// Glacier palette colors (cyan for additions/got, rose for removals/want).
//
// The diff is always returned as a string; callers write it via t.Errorf.
// TTY detection uses os.Stderr.
//
// §21.4 F18; spec 0002 F3b
func renderDiff(got, want any) string {
	gotLines := renderValue(got)
	wantLines := renderValue(want)

	caps := term.Capability(os.Stderr)
	colorOK := caps.IsTTY && !caps.NoColorEnv && caps.SupportsColor != term.ColorNone

	addStyle := term.New().Foreground(term.Cyan)
	remStyle := term.New().Foreground(term.Error) // rose/red

	var b strings.Builder
	b.WriteString("got:\n")
	for _, line := range gotLines {
		raw := "+ " + line
		if colorOK {
			var buf bytes.Buffer
			term.Fprint(&buf, addStyle, raw)
			raw = buf.String()
		}
		b.WriteString(raw)
		b.WriteByte('\n')
	}
	b.WriteString("want:\n")
	for _, line := range wantLines {
		raw := "- " + line
		if colorOK {
			var buf bytes.Buffer
			term.Fprint(&buf, remStyle, raw)
			raw = buf.String()
		}
		b.WriteString(raw)
		b.WriteByte('\n')
	}
	return b.String()
}

// renderValue renders a value as a slice of lines for diff output.
// It uses reflect to walk structs and produce per-field lines.
func renderValue(v any) []string {
	if v == nil {
		return []string{"<nil>"}
	}
	rv := reflect.ValueOf(v)
	return renderReflectValue(rv, "")
}

func renderReflectValue(rv reflect.Value, indent string) []string {
	if !rv.IsValid() {
		return []string{indent + "<nil>"}
	}
	switch rv.Kind() {
	case reflect.Struct:
		t := rv.Type()
		var lines []string
		lines = append(lines, indent+t.Name()+"{")
		for i := range t.NumField() {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			fv := rv.Field(i)
			sub := renderReflectValue(fv, indent+"  ")
			if len(sub) == 1 {
				lines = append(lines, fmt.Sprintf("%s  %s: %s", indent, f.Name, sub[0]))
			} else {
				lines = append(lines, fmt.Sprintf("%s  %s:", indent, f.Name))
				lines = append(lines, sub...)
			}
		}
		lines = append(lines, indent+"}")
		return lines
	case reflect.Ptr:
		if rv.IsNil() {
			return []string{indent + "<nil>"}
		}
		return renderReflectValue(rv.Elem(), indent)
	default:
		if rv.CanInterface() {
			return []string{indent + fmt.Sprintf("%#v", rv.Interface())}
		}
		return []string{indent + "<unexported>"}
	}
}
