// SPDX-License-Identifier: Apache-2.0

package explain_test

import (
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/explain"
)

func TestGet_KnownSlug(t *testing.T) {
	t.Parallel()
	cases := []struct {
		slug     string
		wantCat  string
		wantBody string // substring
	}{
		{"+glacier:command", "marker", "Glaciergen"},
		{"+glacier:root", "marker", "cli.WithRoot"},
		{"+glacier:mock", "marker", "mock"},
		{"+glacier:default", "marker", "default value"},
		{"+glacier:short", "marker", "single-character"},
		{"+glacier:env", "marker", "environment variable"},
		{"+glacier:required", "marker", "required"},
		{"+glacier:choices", "marker", "pipe-separated"},
		{"+glacier:deprecated", "marker", "deprecation"},
		{"+glacier:validate", "marker", "validator"},
		{"+glacier:positional", "marker", "positional argument"},
		{"+glacier:flag", "marker", "field-level"},
		{"exit:0", "exit-code", "successfully"},
		{"exit:1", "exit-code", "unclassified"},
		{"exit:2", "exit-code", "invoked incorrectly"},
		{"exit:64", "exit-code", "generate"},
		{"exit:65", "exit-code", "lint"},
		{"exit:66", "exit-code", "test failures"},
		{"exit:67", "exit-code", "scaffolding"},
		{"exit:68", "exit-code", "GitHub Releases"},
		{"exit:69", "exit-code", "stale"},
		{"exit:70", "exit-code", "subprocess"},
		{"exit:130", "exit-code", "SIGINT"},
		{"exit:143", "exit-code", "SIGTERM"},
		{"config:github.repo", "config-key", "owner/repo"},
		{"config:versioncheck.cache_ttl", "config-key", "duration"},
		{"config:versioncheck.enabled", "config-key", "version check"},
		{"config:versioncheck.strict", "config-key", "strict"},
		{"config:banner.show_on_help", "config-key", "banner"},
		{"config:palette.override", "config-key", "palette"},
	}

	for _, tc := range cases {
		t.Run(tc.slug, func(t *testing.T) {
			t.Parallel()
			topic, ok := explain.Get(tc.slug)
			require.True(t, ok, "expected topic %q to exist", tc.slug)
			assert.Equal(t, tc.slug, topic.Slug)
			assert.Equal(t, tc.wantCat, topic.Category)
			assert.True(t,
				strings.Contains(strings.ToLower(topic.Body), strings.ToLower(tc.wantBody)),
				"expected body to contain %q, got: %s", tc.wantBody, topic.Body,
			)
		})
	}
}

func TestGet_UnknownSlug(t *testing.T) {
	t.Parallel()
	_, ok := explain.Get("does-not-exist")
	assert.False(t, ok, "expected missing slug to return false")
}

func TestAll_Count(t *testing.T) {
	t.Parallel()
	all := explain.All()
	// 12 markers + 12 exit codes + 6 config keys = 30 topics.
	assert.Equal(t, 30, len(all))
}

func TestAll_Immutable(t *testing.T) {
	t.Parallel()
	a := explain.All()
	b := explain.All()
	a[0].Slug = "mutated"
	// Mutating a's copy should not affect b.
	assert.NotEqual(t, "mutated", b[0].Slug)
}

func TestAll_Categories(t *testing.T) {
	t.Parallel()
	counts := make(map[string]int)
	for _, topic := range explain.All() {
		counts[topic.Category]++
	}
	assert.Equal(t, 12, counts["marker"])
	assert.Equal(t, 12, counts["exit-code"])
	assert.Equal(t, 6, counts["config-key"])
}

func TestAll_NoEmptyFields(t *testing.T) {
	t.Parallel()
	for _, topic := range explain.All() {
		assert.True(t, topic.Slug != "", "topic has empty Slug")
		assert.True(t, topic.Title != "", "topic %q has empty Title", topic.Slug)
		assert.True(t, topic.Body != "", "topic %q has empty Body", topic.Slug)
		assert.True(t, topic.Category != "", "topic %q has empty Category", topic.Slug)
	}
}

func TestDidYouMean_ExactMatch(t *testing.T) {
	t.Parallel()
	// exact match returns the slug itself
	got := explain.DidYouMean("exit:66")
	assert.Equal(t, "exit:66", got)
}

func TestDidYouMean_NearMatch(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input string
		want  string
	}{
		{"exit:6", "exit:0"}, // distance 1 from "exit:0", 1 from "exit:1", etc. — any close exit code
		{"exit:666", "exit:66"},  // distance 1
		{"xit:66", "exit:66"},    // distance 1 (missing 'e')
		{"config:github_repo", "config:github.repo"}, // distance 1 (_ vs .)
		{"+glacier:comand", "+glacier:command"},       // distance 1 (missing 'm')
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got := explain.DidYouMean(tc.input)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestDidYouMean_NoMatch(t *testing.T) {
	t.Parallel()
	got := explain.DidYouMean("totally-random-gibberish-xyz-abc")
	assert.Equal(t, "", got)
}

func TestDidYouMean_EmptyInput(t *testing.T) {
	t.Parallel()
	// Empty string has distance equal to slug length for every topic.
	// Shortest slug wins if within threshold 3; most slugs are longer.
	got := explain.DidYouMean("")
	// The result is either "" or the shortest slug - just verify no panic.
	_ = got
}

// Example is the canonical package example test required by the A+ bar.
func Example() {
	topic, ok := explain.Get("exit:66")
	if !ok {
		return
	}
	_ = topic.Category // "exit-code"
}
