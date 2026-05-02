// SPDX-License-Identifier: Apache-2.0

package gen

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// marker kinds supported by the parser.
const (
	markerCommand    = "+glacier:command"
	markerDefault    = "+glacier:default"
	markerShort      = "+glacier:short"
	markerEnv        = "+glacier:env"
	markerRequired   = "+glacier:required"
	markerChoices    = "+glacier:choices"
	markerDeprecated = "+glacier:deprecated"
	markerValidate   = "+glacier:validate"
	markerRoot       = "+glacier:root"
	markerMock       = "+glacier:mock"
)

var (
	nameRe       = regexp.MustCompile(`^[a-z][a-z0-9-]{0,31}$`)
	parentSegRe  = regexp.MustCompile(`^([a-z][a-z0-9-]{0,31}\.)*[a-z][a-z0-9-]{0,31}$`)
	appIdentRe   = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	shortRe      = regexp.MustCompile(`^[a-zA-Z]$`)
	envRe        = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
	choiceRe     = regexp.MustCompile(`^[a-z0-9_-]+(\|[a-z0-9_-]+){0,31}$`)
	choiceElemRe = regexp.MustCompile(`^[a-z0-9_-]+$`)
	deprecatedRe = regexp.MustCompile(`^[^\x00-\x1f"\\]{0,128}$`)
	validateRe   = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
)

// CommandMarker holds parsed +glacier:command attributes.
type CommandMarker struct {
	Name   string // from name=
	Parent string // from parent=
	Alias  string // from alias=
	App    string // from app=; defaults to "cli.Default"
}

// FieldMarker holds parsed per-field markers.
type FieldMarker struct {
	FieldName  string
	Short      rune
	HasShort   bool
	Env        string
	Required   bool
	Choices    []string
	Deprecated string
	HasDepr    bool
	Validate   string
	Default    string
	HasDefault bool
}

// ParseResult is the result of parsing markers from a single type.
type ParseResult struct {
	TypeName string
	Cmd      CommandMarker
	IsRoot   bool
	Fields   map[string]*FieldMarker // keyed by field name
	Unknown  []string                // unrecognized +glacier: markers
}

// ParseMarkers parses +glacier:* marker lines from a doc comment block.
// lines are the raw lines of the doc comment (without the leading "// ").
// typeName is the struct type name for context in error messages.
func ParseMarkers(typeName string, lines []string) (*ParseResult, []error) {
	result := &ParseResult{
		TypeName: typeName,
		Cmd:      CommandMarker{App: "cli.Default"},
		Fields:   make(map[string]*FieldMarker),
	}
	var errs []error

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "+glacier:") {
			continue
		}
		if err := parseLine(result, line); err != nil {
			errs = append(errs, err)
		}
	}

	return result, errs
}

// ParseFieldMarkers parses +glacier:* marker lines from a field's doc comment.
func ParseFieldMarkers(fieldName string, lines []string) (*FieldMarker, []error) {
	fm := &FieldMarker{FieldName: fieldName}
	var errs []error

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "+glacier:") {
			continue
		}
		if err := parseFieldLine(fm, fieldName, line); err != nil {
			errs = append(errs, err)
		}
	}

	return fm, errs
}

func parseLine(r *ParseResult, line string) error {
	// Reject embedded newlines and NUL bytes.
	if strings.ContainsAny(line, "\x00\n\r") {
		return fmt.Errorf("gen: marker %q contains control character", line)
	}

	switch {
	case line == markerRoot:
		r.IsRoot = true

	case line == markerRequired:
		// +glacier:required on a type level is ignored; handled at field level.

	case strings.HasPrefix(line, markerCommand+" ") || line == markerCommand:
		payload := strings.TrimPrefix(line, markerCommand)
		payload = strings.TrimSpace(payload)
		return parseCommandAttrs(r, payload)

	case strings.HasPrefix(line, markerMock):
		// Governed by spec 0012; skip.

	default:
		// Unknown marker.
		r.Unknown = append(r.Unknown, line)
	}

	return nil
}

func parseCommandAttrs(r *ParseResult, payload string) error {
	// payload is like: name=serve parent=root.deploy alias=s app=myApp
	attrs := splitAttrs(payload)
	for _, attr := range attrs {
		kv := strings.SplitN(attr, "=", 2)
		if len(kv) != 2 {
			return fmt.Errorf("gen: malformed command attribute %q", attr)
		}
		k, v := kv[0], kv[1]
		switch k {
		case "name":
			if len(v) > 32 || !nameRe.MatchString(v) || strings.HasSuffix(v, "-") {
				return fmt.Errorf("gen: command name %q does not match ^[a-z][a-z0-9-]{0,31}$ (no trailing dash)", v)
			}
			r.Cmd.Name = v
		case "parent":
			if len(v) > 128 || !parentSegRe.MatchString(v) {
				return fmt.Errorf("gen: parent path %q is not a valid dot-separated name", v)
			}
			r.Cmd.Parent = v
		case "alias":
			if len(v) > 32 || !nameRe.MatchString(v) || strings.HasSuffix(v, "-") {
				return fmt.Errorf("gen: alias %q does not match ^[a-z][a-z0-9-]{0,31}$ (no trailing dash)", v)
			}
			r.Cmd.Alias = v
		case "app":
			if len(v) > 64 || !appIdentRe.MatchString(v) {
				return fmt.Errorf("gen: app identifier %q is not a valid Go identifier", v)
			}
			r.Cmd.App = v
		default:
			return fmt.Errorf("gen: unknown command attribute %q", k)
		}
	}
	return nil
}

func parseFieldLine(fm *FieldMarker, fieldName, line string) error {
	// Reject control chars and NUL.
	if strings.ContainsAny(line, "\x00\n\r") {
		return fmt.Errorf("gen: field marker %q contains control character", line)
	}

	switch {
	case line == markerRequired:
		fm.Required = true

	case line == markerRoot:
		// root on a field is invalid; ignore.

	case strings.HasPrefix(line, markerShort+" ") || strings.HasPrefix(line, markerShort+"\t"):
		payload := strings.TrimSpace(strings.TrimPrefix(line, markerShort))
		if len(payload) != 1 || !shortRe.MatchString(payload) {
			return fmt.Errorf("gen: field %q: +glacier:short %q must be a single [a-zA-Z] character", fieldName, payload)
		}
		fm.Short = rune(payload[0])
		fm.HasShort = true

	case strings.HasPrefix(line, markerEnv+" ") || strings.HasPrefix(line, markerEnv+"\t"):
		payload := strings.TrimSpace(strings.TrimPrefix(line, markerEnv))
		if len(payload) > 64 || !envRe.MatchString(payload) {
			return fmt.Errorf("gen: field %q: +glacier:env %q does not match ^[A-Z][A-Z0-9_]*$", fieldName, payload)
		}
		fm.Env = payload

	case strings.HasPrefix(line, markerChoices+" ") || strings.HasPrefix(line, markerChoices+"\t"):
		payload := strings.TrimSpace(strings.TrimPrefix(line, markerChoices))
		if len(payload) > 256 || !choiceRe.MatchString(payload) {
			return fmt.Errorf("gen: field %q: +glacier:choices %q has invalid format", fieldName, payload)
		}
		parts := strings.Split(payload, "|")
		if len(parts) > 32 {
			return fmt.Errorf("gen: field %q: +glacier:choices has too many choices (%d > 32)", fieldName, len(parts))
		}
		for _, p := range parts {
			if !choiceElemRe.MatchString(p) {
				return fmt.Errorf("gen: field %q: +glacier:choices element %q is invalid", fieldName, p)
			}
		}
		fm.Choices = parts

	case strings.HasPrefix(line, markerDeprecated+" ") || strings.HasPrefix(line, markerDeprecated+"\t"):
		payload := strings.TrimSpace(strings.TrimPrefix(line, markerDeprecated))
		if len(payload) > 128 || !deprecatedRe.MatchString(payload) {
			return fmt.Errorf("gen: field %q: +glacier:deprecated message %q is invalid", fieldName, payload)
		}
		fm.Deprecated = payload
		fm.HasDepr = true

	case line == markerDeprecated:
		fm.HasDepr = true

	case strings.HasPrefix(line, markerValidate+" ") || strings.HasPrefix(line, markerValidate+"\t"):
		payload := strings.TrimSpace(strings.TrimPrefix(line, markerValidate))
		if len(payload) > 64 || !validateRe.MatchString(payload) {
			return fmt.Errorf("gen: field %q: +glacier:validate %q is not a valid Go identifier", fieldName, payload)
		}
		fm.Validate = payload

	case strings.HasPrefix(line, markerDefault+" ") || strings.HasPrefix(line, markerDefault+"\t"):
		payload := strings.TrimSpace(strings.TrimPrefix(line, markerDefault))
		if len(payload) > 256 {
			return fmt.Errorf("gen: field %q: +glacier:default value exceeds 256 chars", fieldName)
		}
		fm.Default = payload
		fm.HasDefault = true

	default:
		if strings.HasPrefix(line, "+glacier:") {
			// Unknown field-level marker — not an error per spec (warn at caller).
		}
	}

	return nil
}

// splitAttrs splits a space-separated attribute string into individual attrs.
func splitAttrs(s string) []string {
	return strings.Fields(s)
}

// quoteString wraps s with strconv.Quote. All string literals that appear in
// generated code MUST pass through this function (§23.8 strconv.Quote invariant).
func quoteString(s string) string {
	return strconv.Quote(s)
}
