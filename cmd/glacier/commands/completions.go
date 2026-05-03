// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/nathanbrophy/glacier/cmd/glacier/internal/completions"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/report"
)

// CompletionsCmd prints shell completion scripts.
//
// +glacier:command name=completions parent=glacier
type CompletionsCmd struct {
	// Shell is the target shell: bash, zsh, fish, or pwsh.
	// Accepted as a positional argument (Amendment E workaround: Run reads
	// os.Args directly until +glacier:positional is wired into cligen).
	//
	// +glacier:positional
	// +glacier:choices bash|zsh|fish|pwsh
	Shell string
}

// Run prints the completion script for the requested shell to stdout.
func (c *CompletionsCmd) Run(_ context.Context) error {
	report.Status(report.Calm, "glacier completions")

	// Positional arg: if Shell wasn't set via flag, read from os.Args.
	if c.Shell == "" {
		c.Shell = firstPositional(1)
	}

	script, ok := completions.Script(c.Shell)
	if !ok {
		report.Status(report.Err, fmt.Sprintf("unknown shell %q: must be bash, zsh, fish, or pwsh", c.Shell))
		return &exitCodeError{
			code:  exitUsage,
			cause: fmt.Errorf("completions: unknown shell %q", c.Shell),
		}
	}

	_, err := os.Stdout.WriteString(script)
	return err
}
