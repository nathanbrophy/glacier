// SPDX-License-Identifier: Apache-2.0

// Package explain provides structured reference topics for the glacier explain
// command. Topics cover markers, exit codes, and config keys. Fuzzy-match
// via DidYouMean helps users when they mistype a slug.
package explain

import (
	"fmt"
	"io/fs"
	"strings"
)

// Topic is a single explain entry.
type Topic struct {
	// Slug is the stable machine-readable identifier used on the CLI.
	Slug string
	// Title is the short human-readable title.
	Title string
	// Body is the multi-line explanation text.
	Body string
	// SeeAlso is a list of related topic slugs.
	SeeAlso []string
	// Category groups topics: "marker", "exit-code", or "config-key".
	Category string
}

// topics is the full set of explain entries in registration order.
var topics []Topic

// index maps slug to position in topics for O(1) lookup.
var index map[string]int

func init() {
	loaded, err := loadTopics(FS)
	if err != nil {
		panic(fmt.Sprintf("explain: failed to load topics from embed.FS: %v", err))
	}
	topics = loaded
	index = make(map[string]int, len(topics))
	for i, t := range topics {
		index[t.Slug] = i
	}
}

// loadTopics parses all *.md files from fsys and returns topics in the
// canonical category order: markers, exit codes, config keys. Within each
// category topics appear in file name sort order.
func loadTopics(fsys fs.FS) ([]Topic, error) {
	entries, err := fs.ReadDir(fsys, "topics")
	if err != nil {
		return nil, fmt.Errorf("explain: read topics dir: %w", err)
	}

	var raw []Topic
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := fs.ReadFile(fsys, "topics/"+e.Name())
		if err != nil {
			return nil, fmt.Errorf("explain: read %s: %w", e.Name(), err)
		}
		t, err := parseTopic(data)
		if err != nil {
			return nil, fmt.Errorf("explain: parse %s: %w", e.Name(), err)
		}
		raw = append(raw, t)
	}

	// Sort into canonical order: markers, exit-codes, config-keys.
	order := []string{"marker", "exit-code", "config-key"}
	catIndex := make(map[string]int, len(order))
	for i, c := range order {
		catIndex[c] = i
	}

	// Stable partition by category while preserving file-name order within each.
	buckets := make([][]Topic, len(order))
	for _, t := range raw {
		i, ok := catIndex[t.Category]
		if !ok {
			return nil, fmt.Errorf("explain: unknown category %q in topic %q", t.Category, t.Slug)
		}
		buckets[i] = append(buckets[i], t)
	}

	var out []Topic
	for _, b := range buckets {
		out = append(out, b...)
	}
	return out, nil
}

// parseTopic reads a topic file: YAML front matter block (between --- delimiters)
// followed by the body. Only the fields slug, title, category, and see_also are
// parsed. see_also is a bracketed comma-separated list of quoted strings.
//
// Tolerant of both LF and CRLF line endings: git's autocrlf may rewrite
// committed files to CRLF on Windows checkouts, and the embed.FS preserves
// whatever the on-disk bytes are.
func parseTopic(data []byte) (Topic, error) {
	// Normalise line endings so the rest of the parser can assume LF.
	text := strings.ReplaceAll(string(data), "\r\n", "\n")

	// Require opening ---.
	if !strings.HasPrefix(text, "---\n") {
		return Topic{}, fmt.Errorf("missing opening --- delimiter")
	}
	rest := text[4:] // strip leading "---\n"

	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		return Topic{}, fmt.Errorf("missing closing --- delimiter")
	}
	frontMatter := rest[:end]
	body := strings.TrimSpace(rest[end+5:]) // +5 to skip "\n---\n"

	var t Topic
	for _, line := range strings.Split(frontMatter, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		switch key {
		case "slug":
			t.Slug = unquoteSimple(val)
		case "title":
			t.Title = unquoteSimple(val)
		case "category":
			t.Category = unquoteSimple(val)
		case "see_also":
			t.SeeAlso = parseSeeAlso(val)
		}
	}

	if t.Slug == "" {
		return Topic{}, fmt.Errorf("missing slug in front matter")
	}
	t.Body = body
	return t, nil
}

// unquoteSimple strips surrounding double-quotes if present.
func unquoteSimple(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// parseSeeAlso parses a YAML inline sequence such as
// ["exit:65", "exit:70"] or ["+glacier:root", "+glacier:mock"].
func parseSeeAlso(s string) []string {
	s = strings.TrimSpace(s)
	if s == "[]" || s == "" {
		return nil
	}
	// Strip surrounding [ ].
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = unquoteSimple(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// All returns all topics in registration order, grouped by category.
func All() []Topic {
	out := make([]Topic, len(topics))
	copy(out, topics)
	return out
}

// Get returns the topic with the given slug.
// Returns (Topic{}, false) when no topic matches.
func Get(slug string) (Topic, bool) {
	i, ok := index[slug]
	if !ok {
		return Topic{}, false
	}
	return topics[i], true
}

// DidYouMean returns the slug of the closest topic within Levenshtein distance 2.
// Returns "" if no match is within that threshold.
func DidYouMean(input string) string {
	best := ""
	bestDist := 3 // exclusive upper bound
	for _, t := range topics {
		d := levenshtein(input, t.Slug)
		if d < bestDist {
			bestDist = d
			best = t.Slug
		}
	}
	return best
}

// levenshtein computes the edit distance between a and b.
// Uses a single-row DP implementation to keep allocations minimal.
func levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	if len(ra) == 0 {
		return len(rb)
	}
	if len(rb) == 0 {
		return len(ra)
	}
	row := make([]int, len(rb)+1)
	for j := range row {
		row[j] = j
	}
	for i, ca := range ra {
		prev := i + 1
		for j, cb := range rb {
			var cost int
			if ca != cb {
				cost = 1
			}
			next := min3(row[j+1]+1, prev+1, row[j]+cost)
			row[j] = prev
			prev = next
		}
		row[len(rb)] = prev
	}
	return row[len(rb)]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
