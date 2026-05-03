// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nathanbrophy/glacier/cmd/glacier/internal/explain"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/report"
	"github.com/nathanbrophy/glacier/term"
)

// ExplainCmd prints an explanation for a Glacier topic.
//
// +glacier:command name=explain parent=glacier
type ExplainCmd struct {
	// Topic is the slug to explain (e.g. "+glacier:command", "exit:64", "config:github.repo").
	//
	// +glacier:positional
	Topic string

	// List prints all available topics grouped by category.
	//
	// +glacier:default false
	List bool
}

// Run prints the explanation for a topic, or lists all topics with --list.
func (c *ExplainCmd) Run(_ context.Context) error {
	report.Status(report.Calm, "glacier explain")

	if c.List {
		return c.runList()
	}

	// Positional arg: if Topic wasn't set via flag, read from os.Args.
	if c.Topic == "" {
		c.Topic = firstPositional(1)
	}

	if c.Topic == "" {
		return c.runList()
	}

	t, ok := explain.Get(c.Topic)
	if !ok {
		suggestion := explain.DidYouMean(c.Topic)
		if suggestion != "" {
			report.Status(report.Err, fmt.Sprintf("unknown topic %q", c.Topic))
			fmt.Fprintf(os.Stderr, "  did you mean %q?\n", suggestion)
		} else {
			report.Status(report.Err, fmt.Sprintf("unknown topic %q; run glacier explain --list to see all topics", c.Topic))
		}
		return &exitCodeError{code: exitUsage, cause: fmt.Errorf("explain: unknown topic %q", c.Topic)}
	}

	kaomoji := categoryKaomoji(t.Category)
	content := strings.Builder{}
	content.WriteString(t.Body)
	if len(t.SeeAlso) > 0 {
		content.WriteString("\n\nSee also: " + strings.Join(t.SeeAlso, ", "))
	}

	box := term.Box(
		content.String(),
		term.WithTitle(kaomoji+"  "+t.Title),
		term.WithRoundedCorners(),
		term.WithPadding(1, 2, 1, 2),
	)
	fmt.Fprintln(os.Stdout, box)
	return nil
}

// runList prints all topics grouped by category.
func (c *ExplainCmd) runList() error {
	all := explain.All()

	type catGroup struct {
		name    string
		kaomoji string
		topics  []explain.Topic
	}

	order := []catGroup{
		{name: "marker", kaomoji: "ʕ•_•ʔ"},
		{name: "exit-code", kaomoji: "ʕ× ×ʔ"},
		{name: "config-key", kaomoji: "ʕ⌐■-■ʔ"},
	}

	index := make(map[string]int, len(order))
	for i, g := range order {
		index[g.name] = i
	}

	for _, t := range all {
		i, ok := index[t.Category]
		if !ok {
			continue
		}
		order[i].topics = append(order[i].topics, t)
	}

	for _, g := range order {
		if len(g.topics) == 0 {
			continue
		}
		fmt.Fprintf(os.Stdout, "\n%s %s\n", g.kaomoji, g.name)
		for _, t := range g.topics {
			fmt.Fprintf(os.Stdout, "  %-40s %s\n", t.Slug, t.Title)
		}
	}
	return nil
}

// categoryKaomoji maps a topic category to its kaomoji.
func categoryKaomoji(category string) string {
	switch category {
	case "marker":
		return "ʕ•_•ʔ"
	case "exit-code":
		return "ʕ× ×ʔ"
	case "config-key":
		return "ʕ⌐■-■ʔ"
	default:
		return "ʕ•ᴥ•ʔ"
	}
}
