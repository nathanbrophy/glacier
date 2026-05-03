// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	cligen "github.com/nathanbrophy/glacier/cli/gen"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/report"
	"github.com/nathanbrophy/glacier/concur"
	httpmockgen "github.com/nathanbrophy/glacier/httpmock/gen"
	mockgen "github.com/nathanbrophy/glacier/mock/gen"
	"github.com/nathanbrophy/glacier/term"
)

// GenerateCmd runs Glacier code generators.
//
// +glacier:command name=generate parent=glacier
type GenerateCmd struct {
	// Patterns are the go/packages patterns to generate for (default ./...).
	//
	// +glacier:positional
	Patterns []string

	// Check performs drift detection without writing files.
	//
	// +glacier:default false
	Check bool

	// Only filters to a comma-separated subset of generators: cli, mock, httpmock.
	//
	// +glacier:validate validateGeneratorList
	Only []string

	// Parallel caps the number of concurrent generators (0 = NumCPU).
	//
	// +glacier:default 0
	Parallel int

	// NoStatus disables the status panel animation.
	//
	// +glacier:default false
	NoStatus bool
}

// validateGeneratorList validates that each element is a known generator name.
func validateGeneratorList(s string) error {
	valid := map[string]bool{"cli": true, "mock": true, "httpmock": true}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if !valid[part] {
			return fmt.Errorf("unknown generator %q: must be cli, mock, or httpmock", part)
		}
	}
	return nil
}

// generator is a named codegen step.
type generator struct {
	name string
	run  func(ctx context.Context, patterns []string, check bool) error
}

// allGenerators returns the full set of registered generators.
func allGenerators() []generator {
	return []generator{
		{
			name: "cli",
			run: func(ctx context.Context, patterns []string, check bool) error {
				for _, p := range patterns {
					if err := cligen.Generate(cligen.Options{
						Pattern: p,
						Check:   check,
						Logger:  slog.Default(),
					}); err != nil {
						return err
					}
				}
				return nil
			},
		},
		{
			name: "mock",
			run: func(ctx context.Context, patterns []string, check bool) error {
				for _, p := range patterns {
					if err := mockgen.Generate(mockgen.Options{
						Pattern: p,
						Check:   check,
						Logger:  slog.Default(),
					}); err != nil {
						return err
					}
				}
				return nil
			},
		},
		{
			name: "httpmock",
			run: func(ctx context.Context, patterns []string, check bool) error {
				for _, p := range patterns {
					if err := httpmockgen.Generate(httpmockgen.Options{
						Pattern: p,
						Check:   check,
						Logger:  slog.Default(),
					}); err != nil {
						return err
					}
				}
				return nil
			},
		},
	}
}

// Run runs the selected generators over patterns.
func (c *GenerateCmd) Run(ctx context.Context) error {
	report.Status(report.Calm, "glacier generate")

	patterns := c.Patterns
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}

	gens := allGenerators()
	if len(c.Only) > 0 {
		allowed := make(map[string]bool, len(c.Only))
		for _, name := range c.Only {
			for _, part := range strings.Split(name, ",") {
				allowed[strings.TrimSpace(part)] = true
			}
		}
		var filtered []generator
		for _, g := range gens {
			if allowed[g.name] {
				filtered = append(filtered, g)
			}
		}
		gens = filtered
	}

	// Optionally show status panel.
	var sb *term.StatusBar
	var animator *term.Animator
	caps := term.Capability(os.Stderr)
	if !c.NoStatus && caps.IsTTY {
		sb = term.NewStatusBar()
		animator = term.NewAnimator(slog.Default())
		animator.Add(sb)
		animCtx, animCancel := context.WithCancel(ctx)
		defer animCancel()
		go func() {
			_ = animator.Run(animCtx)
		}()
	}

	var grp *concur.Group
	if c.Parallel > 0 {
		grp = concur.NewGroup(concur.WithLimit(c.Parallel))
	} else {
		grp = concur.NewGroup()
	}

	var driftMu sync.Mutex
	var driftFiles []string
	var genErrs []error
	var genErrsMu sync.Mutex

	for i := range gens {
		g := gens[i]
		grp.Go(ctx, func() error {
			if sb != nil {
				sb.SetSection(g.name, "ʕ•_•ʔ running "+g.name+"...")
			}
			err := g.run(ctx, patterns, c.Check)
			if sb != nil {
				if err != nil {
					sb.SetSection(g.name, "ʕ× ×ʔ "+g.name+": "+err.Error())
				} else {
					sb.SetSection(g.name, "ʕ⌐■-■ʔ "+g.name+" done")
				}
			}
			if err != nil {
				// Detect drift specifically.
				if strings.Contains(err.Error(), "stale") {
					driftMu.Lock()
					driftFiles = append(driftFiles, err.Error())
					driftMu.Unlock()
					return nil
				}
				genErrsMu.Lock()
				genErrs = append(genErrs, fmt.Errorf("%s: %w", g.name, err))
				genErrsMu.Unlock()
				return nil // collected above; don't double-count
			}
			return nil
		})
	}

	_ = grp.WaitDone(ctx)

	if animator != nil {
		animator.Close()
	}

	if len(driftFiles) > 0 {
		report.Status(report.Err, "codegen drift detected")
		for _, f := range driftFiles {
			fmt.Fprintf(os.Stderr, "  drift: %s\n", f)
		}
		return &exitCodeError{code: exitCodegenDrift, cause: fmt.Errorf("codegen drift detected")}
	}

	if len(genErrs) > 0 {
		for _, e := range genErrs {
			report.Status(report.Err, e.Error())
		}
		return &exitCodeError{code: exitGenerateFailed, cause: genErrs[0]}
	}

	report.Status(report.Confident, "nice.")
	return nil
}
