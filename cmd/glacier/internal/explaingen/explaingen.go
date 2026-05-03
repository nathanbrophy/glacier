// SPDX-License-Identifier: Apache-2.0

// Package explaingen is the build-time generator that renders explain topic
// files from accepted spec sections. It writes one Markdown file per topic
// under cmd/glacier/internal/explain/topics/. The runtime explain package
// embeds these files via //go:embed.
//
// Source-of-truth categories:
//   - marker: from specs/0011-cli.md §marker grammar table
//   - exit-code: from specs/0032-sdk.md D-S27 exit code table
//   - config-key: from specs/0032-sdk.md D-S21 config schema table
package explaingen

import (
	"bytes"
	"fmt"
	"io/fs"
	"strings"
)

// MarkerRow is one row extracted from the marker grammar table.
type MarkerRow struct {
	// Name is the full marker name including the +glacier: prefix.
	Name string
	// Notes is the human-readable description from the spec table.
	Notes string
}

// ExitCodeRow is one row extracted from the exit code table.
type ExitCodeRow struct {
	// Code is the numeric exit code as a string (e.g. "66").
	Code string
	// Meaning is the description from the spec table.
	Meaning string
}

// ConfigKeyRow is one row extracted from the SDK config schema table.
type ConfigKeyRow struct {
	// Key is the dotted config key name (e.g. "github.repo").
	Key string
	// Type is the value type (e.g. "string", "bool", "duration").
	Type string
	// Default is the default value string.
	Default string
	// Effect is the human-readable description.
	Effect string
}

// SpecSource reads spec files and extracts the three topic categories.
// Implement this interface with a real filesystem reader for production use,
// or a fixture implementation for tests.
//
// +glacier:mock
type SpecSource interface {
	// Markers returns all rows from the +glacier: marker grammar table in
	// specs/0011-cli.md.
	Markers() ([]MarkerRow, error)
	// ExitCodes returns all rows from the exit code table in
	// specs/0032-sdk.md D-S27.
	ExitCodes() ([]ExitCodeRow, error)
	// ConfigKeys returns all rows from the config schema table in
	// specs/0032-sdk.md D-S21.
	ConfigKeys() ([]ConfigKeyRow, error)
}

// FileSpecSource implements SpecSource by reading spec files from an fs.FS.
type FileSpecSource struct {
	fsys fs.FS
}

// NewFileSpecSource returns a SpecSource backed by fsys, which must contain
// the spec files at their canonical paths relative to the repo root
// (specs/0011-cli.md and specs/0032-sdk.md).
func NewFileSpecSource(fsys fs.FS) *FileSpecSource {
	return &FileSpecSource{fsys: fsys}
}

// Markers reads the marker grammar table from specs/0011-cli.md.
func (s *FileSpecSource) Markers() ([]MarkerRow, error) {
	data, err := fs.ReadFile(s.fsys, "specs/0011-cli.md")
	if err != nil {
		return nil, fmt.Errorf("explaingen: read 0011-cli.md: %w", err)
	}
	return parseMarkerTable(string(data))
}

// ExitCodes reads the exit code table from specs/0032-sdk.md.
func (s *FileSpecSource) ExitCodes() ([]ExitCodeRow, error) {
	data, err := fs.ReadFile(s.fsys, "specs/0032-sdk.md")
	if err != nil {
		return nil, fmt.Errorf("explaingen: read 0032-sdk.md: %w", err)
	}
	return parseExitCodeTable(string(data))
}

// ConfigKeys reads the config schema table from specs/0032-sdk.md.
func (s *FileSpecSource) ConfigKeys() ([]ConfigKeyRow, error) {
	data, err := fs.ReadFile(s.fsys, "specs/0032-sdk.md")
	if err != nil {
		return nil, fmt.Errorf("explaingen: read 0032-sdk.md: %w", err)
	}
	return parseConfigTable(string(data))
}

// parseMarkerTable scans text for the marker grammar table (the block starting
// with the header "| Marker | Regex |") and returns one MarkerRow per data row.
// Rows for +glacier:command sub-attributes (name=, parent=, alias=, app=) are
// collapsed into the base +glacier:command row.
func parseMarkerTable(text string) ([]MarkerRow, error) {
	const header = "| Marker | Regex |"
	idx := strings.Index(text, header)
	if idx < 0 {
		return nil, fmt.Errorf("marker grammar table not found in 0011-cli.md")
	}
	lines := strings.Split(text[idx:], "\n")

	seen := make(map[string]bool)
	var rows []MarkerRow
	for _, line := range lines[2:] { // skip header + separator
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			break
		}
		cols := splitTableRow(line)
		if len(cols) < 4 {
			continue
		}
		name := strings.TrimSpace(cols[0])
		notes := strings.TrimSpace(cols[3])

		// Normalise: strip backtick wrapping.
		name = strings.Trim(name, "`")

		// Collapse sub-attribute variants (+glacier:command name=, etc.)
		// into the base marker name.
		base := name
		if i := strings.Index(name, " "); i > 0 {
			base = name[:i]
		}
		// Skip blank or separator rows.
		if base == "" || base == "Marker" {
			continue
		}
		if seen[base] {
			continue
		}
		seen[base] = true
		rows = append(rows, MarkerRow{Name: base, Notes: notes})
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no marker rows parsed from 0011-cli.md")
	}
	return rows, nil
}

// parseExitCodeTable scans text for the exit code table (the block following
// the "#### Exit codes" heading) and returns one ExitCodeRow per data row.
func parseExitCodeTable(text string) ([]ExitCodeRow, error) {
	const anchor = "#### Exit codes"
	idx := strings.Index(text, anchor)
	if idx < 0 {
		return nil, fmt.Errorf("exit code table not found in 0032-sdk.md")
	}
	// Find the table start: first line beginning with "| Code"
	const tableHeader = "| Code | Meaning |"
	tIdx := strings.Index(text[idx:], tableHeader)
	if tIdx < 0 {
		return nil, fmt.Errorf("exit code table header not found after anchor")
	}
	lines := strings.Split(text[idx+tIdx:], "\n")

	var rows []ExitCodeRow
	for _, line := range lines[2:] { // skip header + separator
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			break
		}
		cols := splitTableRow(line)
		if len(cols) < 2 {
			continue
		}
		code := strings.TrimSpace(cols[0])
		meaning := strings.TrimSpace(cols[1])
		if code == "" || code == "Code" {
			continue
		}
		rows = append(rows, ExitCodeRow{Code: code, Meaning: meaning})
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no exit code rows parsed from 0032-sdk.md")
	}
	return rows, nil
}

// parseConfigTable scans text for the config schema table (the block following
// the "#### Configuration" heading with a "| Key | Type |" header) and returns
// one ConfigKeyRow per data row.
func parseConfigTable(text string) ([]ConfigKeyRow, error) {
	const anchor = "#### Configuration"
	idx := strings.Index(text, anchor)
	if idx < 0 {
		return nil, fmt.Errorf("config table not found in 0032-sdk.md")
	}
	const tableHeader = "| Key | Type | Default | Effect |"
	tIdx := strings.Index(text[idx:], tableHeader)
	if tIdx < 0 {
		return nil, fmt.Errorf("config table header not found after anchor")
	}
	lines := strings.Split(text[idx+tIdx:], "\n")

	var rows []ConfigKeyRow
	for _, line := range lines[2:] { // skip header + separator
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			break
		}
		cols := splitTableRow(line)
		if len(cols) < 4 {
			continue
		}
		key := strings.TrimSpace(cols[0])
		typ := strings.TrimSpace(cols[1])
		def := strings.TrimSpace(cols[2])
		effect := strings.TrimSpace(cols[3])
		// Strip backtick wrapping from key.
		key = strings.Trim(key, "`")
		if key == "" || key == "Key" {
			continue
		}
		rows = append(rows, ConfigKeyRow{Key: key, Type: typ, Default: def, Effect: effect})
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no config key rows parsed from 0032-sdk.md")
	}
	return rows, nil
}

// splitTableRow splits a Markdown table row on | delimiters and returns the
// trimmed cell values, excluding the leading and trailing empty cells that
// result from the surrounding | characters.
func splitTableRow(line string) []string {
	parts := strings.Split(line, "|")
	if len(parts) < 2 {
		return nil
	}
	// Trim first and last (empty due to surrounding |).
	return parts[1 : len(parts)-1]
}

// Generate reads from src and writes topic Markdown files into outDir.
// outDir is the path to the topics/ directory (e.g. "topics").
// Returns the map of filename -> content for all generated files.
func Generate(src SpecSource) (map[string][]byte, error) {
	markers, err := src.Markers()
	if err != nil {
		return nil, err
	}
	exitCodes, err := src.ExitCodes()
	if err != nil {
		return nil, err
	}
	configKeys, err := src.ConfigKeys()
	if err != nil {
		return nil, err
	}

	files := make(map[string][]byte)

	for _, m := range markers {
		slug := m.Name
		filename := slugToFilename(slug)
		seeAlso := markerSeeAlso(slug, markers)
		body := markerBody(m)
		files[filename] = renderTopic(slug, markerTitle(slug), "marker", seeAlso, body)
	}

	// Synthesize the +glacier:flag meta-topic (D-S70: 12 markers). This topic
	// groups all field-level markers into one reference entry. It is not a row
	// in the spec grammar table; it is an aggregate summary generated from the
	// set of field-level marker names.
	flagSlug := "+glacier:flag"
	files[slugToFilename(flagSlug)] = renderTopic(
		flagSlug,
		markerTitle(flagSlug),
		"marker",
		[]string{"+glacier:default", "+glacier:short", "+glacier:env", "+glacier:required", "+glacier:choices"},
		"Meta-category covering all field-level +glacier: markers that configure flag behavior: "+
			"+glacier:default, +glacier:short, +glacier:env, +glacier:required, +glacier:choices, "+
			"+glacier:deprecated, +glacier:validate, +glacier:positional.",
	)

	for _, ec := range exitCodes {
		slug := "exit:" + ec.Code
		filename := slugToFilename(slug)
		seeAlso := exitCodeSeeAlso(ec.Code, exitCodes)
		body := exitCodeBody(ec)
		files[filename] = renderTopic(slug, exitCodeTitle(ec), "exit-code", seeAlso, body)
	}

	for _, ck := range configKeys {
		slug := "config:" + ck.Key
		filename := slugToFilename(slug)
		seeAlso := configKeySeeAlso(ck.Key, configKeys)
		body := configKeyBody(ck)
		files[filename] = renderTopic(slug, configKeyTitle(ck), "config-key", seeAlso, body)
	}

	return files, nil
}

// Check re-generates topics from src and compares them against the on-disk
// files in fsys (rooted at "topics/"). Returns a non-nil error listing any
// file that would change. This is the CI freshness gate (X17).
func Check(src SpecSource, fsys fs.FS) error {
	generated, err := Generate(src)
	if err != nil {
		return fmt.Errorf("explaingen --check: generate: %w", err)
	}

	var stale []string
	for name, want := range generated {
		have, err := fs.ReadFile(fsys, "topics/"+name)
		if err != nil {
			stale = append(stale, fmt.Sprintf("%s: missing (generate it with explaingen)", name))
			continue
		}
		if !bytes.Equal(have, want) {
			stale = append(stale, fmt.Sprintf("%s: content differs", name))
		}
	}

	// Also check for files in fsys that are no longer generated.
	entries, err := fs.ReadDir(fsys, "topics")
	if err != nil {
		return fmt.Errorf("explaingen --check: read topics dir: %w", err)
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		if _, ok := generated[e.Name()]; !ok {
			stale = append(stale, fmt.Sprintf("%s: extra file not in spec", e.Name()))
		}
	}

	if len(stale) > 0 {
		return fmt.Errorf("explaingen --check: %d stale topic file(s):\n  %s\nRun explaingen to regenerate.",
			len(stale), strings.Join(stale, "\n  "))
	}
	return nil
}

// slugToFilename converts a topic slug to a filesystem-safe filename.
// Colons (:) become underscores (_); the result is lowercased.
// The +glacier: prefix retains its + and letters.
func slugToFilename(slug string) string {
	return strings.ReplaceAll(slug, ":", "_") + ".md"
}

// renderTopic formats a single topic file with YAML front matter.
func renderTopic(slug, title, category string, seeAlso []string, body string) []byte {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("slug: " + slug + "\n")
	b.WriteString("title: " + fmt.Sprintf("%q", title) + "\n")
	b.WriteString("category: " + category + "\n")
	if len(seeAlso) > 0 {
		quoted := make([]string, len(seeAlso))
		for i, s := range seeAlso {
			quoted[i] = fmt.Sprintf("%q", s)
		}
		b.WriteString("see_also: [" + strings.Join(quoted, ", ") + "]\n")
	}
	b.WriteString("---\n")
	b.WriteString(body)
	b.WriteString("\n")
	return []byte(b.String())
}

// markerTitle returns the canonical title for a marker topic.
func markerTitle(slug string) string {
	return "Marker: " + slug
}

// markerBody builds the body text for a marker topic. The spec notes column
// provides a concise grammar description; this function maps to the richer
// user-facing prose that matches D-S70 topic content.
func markerBody(m MarkerRow) string {
	bodies := map[string]string{
		"+glacier:command":    "Annotates a struct as a CLI command. Glaciergen reads this marker and registers the struct in the generated command tree. Required attributes: name=<slug>. Optional: parent=<path>, alias=<name>, app=<var>.\n\nThe struct must implement Run(ctx context.Context) error.",
		"+glacier:root":       "Marks a command struct as the root of the CLI application. Exactly one +glacier:root must exist per binary. Glaciergen emits cli.WithRoot() for the registration call.",
		"+glacier:mock":       "Annotates an interface for mock generation. Glaciergen emits a full mock implementation under the mock sub-package of the annotated package.",
		"+glacier:default":    "Sets the default value for a command flag field. The value is parsed according to the field's type (bool, string, int, duration, etc.).\n\nExample: // +glacier:default 30s",
		"+glacier:short":      "Assigns a single-character short flag to a field. The value must be a single ASCII letter (A-Z, a-z).\n\nExample: // +glacier:short q",
		"+glacier:env":        "Binds an environment variable to a flag field. When the env var is set and the flag is not provided, the env var's value is used. The value must be UPPER_CASE_WITH_UNDERSCORES.\n\nExample: // +glacier:env OTEL_EXPORTER_OTLP_ENDPOINT",
		"+glacier:required":   "Marks a flag field as required. If the flag is not provided (via argv or env var), the command exits with code 2 before Run() is called.",
		"+glacier:choices":    "Constrains a flag field to a fixed set of allowed values. Values are pipe-separated. Providing a value outside the set exits with code 2.\n\nExample: // +glacier:choices text|json|sarif",
		"+glacier:deprecated": "Marks a flag field as deprecated. Glaciergen emits cli.WithFlagDeprecated() which prints a deprecation notice when the flag is used. The value is the deprecation message.",
		"+glacier:validate":   "Wires a custom validator function to a flag field. The value is the name of a func(string) error in the same package. Called after flag parsing, before Run().\n\nExample: // +glacier:validate validateGeneratorList",
		"+glacier:positional": "Marks a field as a positional argument rather than a named flag. Positional arguments are read from os.Args in the order they appear on the command struct.\n\nNote: full +glacier:positional support (Amendment E) is implemented in cli/gen; until a binary is regenerated the field reads from os.Args directly in Run().",
	}
	if b, ok := bodies[m.Name]; ok {
		return b
	}
	// Fall back to spec notes for any unrecognised marker.
	return m.Notes
}

// markerSeeAlso returns related slugs for a marker topic.
func markerSeeAlso(slug string, all []MarkerRow) []string {
	// Hard-coded adjacency from the existing topic definitions.
	adj := map[string][]string{
		"+glacier:command":    {"+glacier:root", "+glacier:mock"},
		"+glacier:root":       {"+glacier:command"},
		"+glacier:mock":       {"+glacier:command"},
		"+glacier:default":    {"+glacier:required"},
		"+glacier:short":      {"+glacier:command"},
		"+glacier:env":        {"+glacier:default"},
		"+glacier:required":   {"+glacier:default"},
		"+glacier:choices":    {"+glacier:validate"},
		"+glacier:deprecated": {"+glacier:command"},
		"+glacier:validate":   {"+glacier:choices"},
		"+glacier:positional": {"+glacier:required"},
		"+glacier:flag":       {"+glacier:default", "+glacier:short", "+glacier:env", "+glacier:required", "+glacier:choices"},
	}
	return adj[slug]
}

// exitCodeTitle returns the canonical title for an exit code topic.
func exitCodeTitle(ec ExitCodeRow) string {
	return "Exit code " + ec.Code + ": " + exitCodeShortMeaning(ec.Code)
}

// exitCodeShortMeaning maps an exit code to its short description.
func exitCodeShortMeaning(code string) string {
	meanings := map[string]string{
		"0":   "success",
		"1":   "general error",
		"2":   "usage error",
		"64":  "generate failed",
		"65":  "lint findings",
		"66":  "tests failed",
		"67":  "scaffolding failed",
		"68":  "version check unreachable",
		"69":  "codegen drift",
		"70":  "subprocess failure",
		"130": "interrupted",
		"143": "terminated",
	}
	if m, ok := meanings[code]; ok {
		return m
	}
	return "unknown"
}

// exitCodeBody builds the body text for an exit code topic.
func exitCodeBody(ec ExitCodeRow) string {
	bodies := map[string]string{
		"0":   "The command completed successfully. All outputs are valid and no errors were encountered.",
		"1":   "An unclassified runtime error occurred. No more specific exit code applies. Check stderr for the error message.",
		"2":   "The command was invoked incorrectly: unknown flag, malformed argument, conflicting options, or missing required flag. Correct the invocation and retry.\n\nRun glacier explain +glacier:required for flag validation details.",
		"64":  "glacier generate encountered an error and could not produce one or more output files. Check stderr for the generator name and cause.\n\nRun glacier explain +glacier:command to review marker syntax.",
		"65":  "glacier lint reported one or more findings at or above the configured severity threshold. Fix the listed findings or lower the threshold with --severity.",
		"66":  "glacier test reported one or more test failures, or a benchmark regressed by more than 5% vs the stored baseline.\n\nFor benchmark regressions, run glacier test --update-baseline to accept the new performance level.",
		"67":  "glacier init or glacier new could not complete the scaffolding operation. Common causes: target directory already exists (use --force to overwrite), disk full, or invalid module name.",
		"68":  "glacier version --check could not reach the GitHub Releases API and --strict was set. Without --strict the command exits 0 with an (offline) annotation.\n\nRun glacier explain config:versioncheck.strict for the config key.",
		"69":  "glacier generate --check detected that one or more generated files are stale. Run glacier generate (without --check) to regenerate and commit the result.",
		"70":  "A subprocess launched by glacier (go test, staticcheck) exited non-zero for a reason unrelated to test or lint findings. Check stderr for the subprocess output.",
		"130": "The process was interrupted by SIGINT (Ctrl-C). glacier handles SIGINT gracefully via context cancellation; any in-progress writes are flushed before exit.",
		"143": "The process was terminated by SIGTERM. glacier handles SIGTERM gracefully via the internal/sigh package.",
	}
	if b, ok := bodies[ec.Code]; ok {
		return b
	}
	return ec.Meaning
}

// exitCodeSeeAlso returns related slugs for an exit code topic.
func exitCodeSeeAlso(code string, all []ExitCodeRow) []string {
	adj := map[string][]string{
		"0":   {"exit:1"},
		"1":   {"exit:0", "exit:2"},
		"2":   {"exit:64"},
		"64":  {"exit:2", "exit:69"},
		"65":  {"exit:66"},
		"66":  {"exit:65"},
		"67":  {"exit:68"},
		"68":  {"exit:67", "config:versioncheck.strict"},
		"69":  {"exit:64"},
		"70":  {"exit:66", "exit:65"},
		"130": {"exit:143"},
		"143": {"exit:130"},
	}
	return adj[code]
}

// configKeyTitle returns the canonical title for a config key topic.
func configKeyTitle(ck ConfigKeyRow) string {
	return "Config key: " + ck.Key
}

// configKeyBody builds the body text for a config key topic.
func configKeyBody(ck ConfigKeyRow) string {
	bodies := map[string]string{
		"github.repo":              "The GitHub repository in owner/repo format used by glacier version when checking for a newer release.\n\nDefault: nathanbrophy/glacier\nEnv override: GLACIER__GITHUB__REPO",
		"versioncheck.cache_ttl":   "How long the version check result is cached before re-fetching from GitHub Releases. Accepts Go duration syntax.\n\nDefault: 24h\nEnv override: GLACIER__VERSIONCHECK__CACHE_TTL",
		"versioncheck.enabled":     "Enables or disables the background version check performed by glacier version. Set to false to skip the GitHub Releases request entirely.\n\nDefault: true\nEnv override: GLACIER__VERSIONCHECK__ENABLED",
		"versioncheck.strict":      "When true, glacier version exits with code 68 if the GitHub Releases API is unreachable. Without this, unreachability exits 0 with an (offline) annotation.\n\nDefault: false\nEnv override: GLACIER__VERSIONCHECK__STRICT",
		"banner.show_on_help":      "Controls whether the Glacier block-character banner is printed when --help is invoked. Suppress it with --no-banner or NO_COLOR.\n\nDefault: true\nEnv override: GLACIER__BANNER__SHOW_ON_HELP",
		"palette.override":         "Overrides the active colour palette. Accepted values: 'default', 'accessible', 'none'. 'none' disables all ANSI colour output.\n\nDefault: default\nEnv override: GLACIER__PALETTE__OVERRIDE",
	}
	if b, ok := bodies[ck.Key]; ok {
		return b
	}
	return ck.Effect
}

// configKeySeeAlso returns related slugs for a config key topic.
func configKeySeeAlso(key string, all []ConfigKeyRow) []string {
	adj := map[string][]string{
		"github.repo":            {"config:versioncheck.enabled"},
		"versioncheck.cache_ttl": {"config:versioncheck.enabled"},
		"versioncheck.enabled":   {"config:versioncheck.strict", "config:versioncheck.cache_ttl"},
		"versioncheck.strict":    {"config:versioncheck.enabled", "exit:68"},
		"banner.show_on_help":    {"config:palette.override"},
		"palette.override":       {"config:banner.show_on_help"},
	}
	return adj[key]
}
