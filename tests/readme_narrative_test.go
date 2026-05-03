// SPDX-License-Identifier: Apache-2.0

package tests

// TestREADMENarrativeMatchesCommandTree asserts that the command table in
// README.md §Glacier SDK lists exactly the top-level commands registered in
// cmd/glacier/commands/zz_generated_cli.go.
//
// Drift in either direction (README names a command that is not registered, or
// a registered top-level command is absent from the README) fails the test.
//
// Verification gate 9 (specs/0032-sdk.md §Verification).
//
// Discovery strategy: the test parses zz_generated_cli.go via go/ast to
// extract WithName("...") calls that are NOT paired with a WithParent call
// referencing a non-root command. The root command ("glacier") and its
// direct children are top-level commands. Subcommands (e.g. new.package,
// new.command, new.option) are skipped because the README only lists
// top-level commands.

import (
	"bufio"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert/require"
)

// readmeCommandRE matches a backtick-quoted `glacier <name>` entry in a
// Markdown table cell, capturing the command name.
var readmeCommandRE = regexp.MustCompile("`glacier ([a-z]+)`")

// extractREADMECommands reads README.md from root and extracts the command
// names listed in the Glacier SDK command table.
func extractREADMECommands(t *testing.T, root string) map[string]struct{} {
	t.Helper()
	f, err := os.Open(filepath.Join(root, "README.md"))
	require.NoError(t, err, "open README.md")
	defer f.Close()

	inSDKSection := false
	inCommandTable := false
	commands := make(map[string]struct{})

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Track section boundaries.
		if strings.HasPrefix(trimmed, "## ") {
			if trimmed == "## Glacier SDK" {
				inSDKSection = true
				continue
			} else if inSDKSection {
				// Any subsequent h2 ends the SDK section.
				break
			}
		}
		if !inSDKSection {
			continue
		}

		// Detect the command table header.
		if strings.Contains(trimmed, "| Command |") {
			inCommandTable = true
			continue
		}
		// Skip the separator row.
		if inCommandTable && strings.HasPrefix(trimmed, "|---") {
			continue
		}
		// Exit table on blank line or non-table line.
		if inCommandTable && !strings.HasPrefix(trimmed, "|") {
			break
		}

		if inCommandTable {
			m := readmeCommandRE.FindStringSubmatch(line)
			if m != nil {
				commands[m[1]] = struct{}{}
			}
		}
	}
	require.NoError(t, scanner.Err(), "scanning README.md")
	return commands
}

// extractRegisteredTopLevelCommands parses zz_generated_cli.go and returns
// the names of commands registered without a WithParent option (or with
// WithParent pointing to the root command name, which is excluded from
// the top-level list).
//
// The generated file contains calls like:
//
//	cli.Default.Register(&FooCmd{}, cli.WithName("foo"), ...)
//	cli.Default.Register(&BarCmd{}, cli.WithName("bar"), cli.WithParent("new"), ...)
//
// Top-level commands have no WithParent call, or have WithParent("glacier")
// which is the binary root and is itself not a user-visible subcommand.
func extractRegisteredTopLevelCommands(t *testing.T, root string) map[string]struct{} {
	t.Helper()

	genFile := filepath.Join(root, "cmd", "glacier", "commands", "zz_generated_cli.go")
	src, err := os.ReadFile(genFile)
	require.NoError(t, err, "read zz_generated_cli.go")

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, genFile, src, 0)
	require.NoError(t, err, "parse zz_generated_cli.go")

	// rootCommandName is the binary's own registered name (not user-visible as
	// a subcommand); we skip it when building the top-level set.
	const rootCommandName = "glacier"

	commands := make(map[string]struct{})

	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		// We want calls of the form cli.Default.Register(...).
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Register" {
			return true
		}

		// Walk the arguments to find WithName and WithParent values.
		var name, parent string
		for _, arg := range call.Args {
			inner, ok := arg.(*ast.CallExpr)
			if !ok {
				continue
			}
			fnSel, ok := inner.Fun.(*ast.SelectorExpr)
			if !ok {
				continue
			}
			switch fnSel.Sel.Name {
			case "WithName":
				if len(inner.Args) == 1 {
					if lit, ok := inner.Args[0].(*ast.BasicLit); ok {
						name = strings.Trim(lit.Value, `"`)
					}
				}
			case "WithParent":
				if len(inner.Args) == 1 {
					if lit, ok := inner.Args[0].(*ast.BasicLit); ok {
						parent = strings.Trim(lit.Value, `"`)
					}
				}
			}
		}

		// Include the command only when it has a name, is not the root itself,
		// and does not declare a non-root parent.
		if name != "" && name != rootCommandName && parent == "" {
			commands[name] = struct{}{}
		}
		return true
	})

	return commands
}

func TestREADMENarrativeMatchesCommandTree(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)

	readmeCmds := extractREADMECommands(t, root)
	registeredCmds := extractRegisteredTopLevelCommands(t, root)

	// Every registered top-level command must appear in the README.
	for cmd := range registeredCmds {
		if _, ok := readmeCmds[cmd]; !ok {
			t.Errorf("registered command %q is missing from the README.md Glacier SDK command table", cmd)
		}
	}

	// Every README command must be registered.
	for cmd := range readmeCmds {
		if _, ok := registeredCmds[cmd]; !ok {
			t.Errorf("README.md lists command %q but it is not registered in zz_generated_cli.go", cmd)
		}
	}

	if t.Failed() {
		t.Logf("README commands:    %v", setKeys(readmeCmds))
		t.Logf("Registered commands: %v", setKeys(registeredCmds))
	}
}

// setKeys returns a sorted slice of keys from a set map for display.
func setKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// sort inline without importing sort to stay under 150 LOC
	for i := range keys {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}
