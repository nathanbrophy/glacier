// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/nathanbrophy/glacier/term"
)

// renderHelp writes the help page for the command at path to w.
// path == "" or path matching the root path renders the format-A
// "top-level" page (banner + subcommand listing + global flags + footer).
// Any other path renders the format-B "per-command" page (synopsis +
// long-form description + flags + global-flags pointer + see-also).
//
// Color is emitted when w is a TTY that supports color and neither
// NO_COLOR nor GLACIER_NO_COLOR is set in the environment.
func (a *App) renderHelp(w io.Writer, path string) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if path == "" {
		path = a.rootPath
	}

	e, ok := a.commands[path]
	if !ok {
		return &ErrUnknownCommand{Path: path}
	}

	useColor := term.ShouldColor(w)

	if e.cfg.root {
		a.renderRoot(w, e, useColor)
		return nil
	}
	a.renderSubcommand(w, e, useColor)
	return nil
}

// renderRoot writes the top-level help page (format A).
func (a *App) renderRoot(w io.Writer, root *entry, useColor bool) {
	// 1. Banner.
	if !a.cfg.noBanner {
		writeBanner(w)
	}

	// 2. Tagline (cli prints this as part of the banner already; skip duplication).

	// 3. Usage line.
	fmt.Fprintln(w, header("USAGE", useColor))
	fmt.Fprintf(w, "  %s [global flags] <command> [command flags] [args]\n\n", root.cfg.name)

	// 4. Two-column command listing grouped by category.
	groups := a.groupCommands()
	categoryOrder := []string{"create", "develop", "inspect", "utility", "other"}
	first := true
	for _, cat := range categoryOrder {
		entries, ok := groups[cat]
		if !ok || len(entries) == 0 {
			continue
		}
		if !first {
			fmt.Fprintln(w)
		}
		first = false
		fmt.Fprintln(w, header(strings.ToUpper(cat), useColor))
		// Sort entries by command name for deterministic output.
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].cfg.name < entries[j].cfg.name
		})
		// Find longest name for column alignment.
		maxName := 0
		for _, c := range entries {
			if len(c.cfg.name) > maxName {
				maxName = len(c.cfg.name)
			}
		}
		colWidth := maxName + 2
		for _, c := range entries {
			summary := c.cfg.summary
			if summary == "" {
				summary = ""
			}
			fmt.Fprintf(w, "  %s%s%s\n",
				colorize(c.cfg.name, useColor, ansiCyan),
				strings.Repeat(" ", colWidth-len(c.cfg.name)),
				summary)
		}
	}

	// 5. GLOBAL FLAGS table.
	fmt.Fprintln(w)
	fmt.Fprintln(w, header("GLOBAL FLAGS", useColor))
	a.writeRootFlags(w, root, useColor)

	// 6. Footer.
	fmt.Fprintln(w)
	fmt.Fprintln(w, footer(root.cfg.name, useColor))
}

// renderSubcommand writes the per-command help page (format B).
func (a *App) renderSubcommand(w io.Writer, e *entry, useColor bool) {
	// Synopsis.
	fmt.Fprintln(w, header("SYNOPSIS", useColor))
	fmt.Fprintf(w, "  %s %s [flags]\n", a.rootPath, strings.ReplaceAll(e.path, ".", " "))
	if e.cfg.summary != "" {
		fmt.Fprintf(w, "\n  %s\n", e.cfg.summary)
	}
	fmt.Fprintln(w)

	// Long description (if any).
	if e.cfg.longDesc != "" {
		fmt.Fprintln(w, header("DESCRIPTION", useColor))
		for _, line := range strings.Split(strings.TrimSpace(e.cfg.longDesc), "\n") {
			fmt.Fprintf(w, "  %s\n", line)
		}
		fmt.Fprintln(w)
	}

	// Subcommands of THIS command (e.g. `glacier new` has command/option/package).
	if children := a.childrenOf(e.path); len(children) > 0 {
		fmt.Fprintln(w, header("SUBCOMMANDS", useColor))
		sort.Slice(children, func(i, j int) bool {
			return children[i].cfg.name < children[j].cfg.name
		})
		maxName := 0
		for _, c := range children {
			if len(c.cfg.name) > maxName {
				maxName = len(c.cfg.name)
			}
		}
		for _, c := range children {
			fmt.Fprintf(w, "  %s%s%s\n",
				colorize(c.cfg.name, useColor, ansiCyan),
				strings.Repeat(" ", maxName+2-len(c.cfg.name)),
				c.cfg.summary)
		}
		fmt.Fprintln(w)
	}

	// Flags.
	fs, _ := buildFlagSet(e.cmd, e.cfg, e.path)
	if hasFlags(fs) {
		fmt.Fprintln(w, header("FLAGS", useColor))
		writeFlags(w, fs, e.cfg, useColor)
		fmt.Fprintln(w)
	}

	// Pointer to global flags.
	fmt.Fprintf(w, "Run %s for global flags.\n",
		colorize(a.rootPath+" --help", useColor, ansiUnderline))
}

// groupCommands returns top-level commands grouped by their declared category.
// The root command and sub-sub-commands (those whose parent is not the root)
// are excluded from the listing.
func (a *App) groupCommands() map[string][]*entry {
	groups := make(map[string][]*entry)
	for path, e := range a.commands {
		if e.cfg.root {
			continue
		}
		// Only include direct children of root: their path is just the name.
		// Sub-sub-commands have a parent set and a dotted path.
		if strings.Contains(path, ".") {
			continue
		}
		cat := e.cfg.category
		if cat == "" {
			cat = "other"
		}
		groups[cat] = append(groups[cat], e)
	}
	return groups
}

// childrenOf returns all direct children of the command at path.
func (a *App) childrenOf(path string) []*entry {
	var out []*entry
	prefix := path + "."
	for p, e := range a.commands {
		if !strings.HasPrefix(p, prefix) {
			continue
		}
		// Direct children only: no further dots after the prefix.
		rest := strings.TrimPrefix(p, prefix)
		if strings.Contains(rest, ".") {
			continue
		}
		out = append(out, e)
	}
	return out
}

// writeRootFlags renders the root command's flags in the GLOBAL FLAGS table.
func (a *App) writeRootFlags(w io.Writer, root *entry, useColor bool) {
	fs, _ := buildFlagSet(root.cmd, root.cfg, root.path)
	writeFlags(w, fs, root.cfg, useColor)
}

// writeFlags renders a flag.FlagSet using the cli's preferred two-column
// layout: "  --name <type>  description (default: <default>)".
func writeFlags(w io.Writer, fs *flag.FlagSet, cfg regConfig, useColor bool) {
	type flagRow struct {
		name, kind, desc, def, env string
		short                      rune
	}
	var rows []flagRow
	maxLeft := 0
	fs.VisitAll(func(f *flag.Flag) {
		row := flagRow{
			name: f.Name,
			kind: flagKind(f),
			desc: f.Usage,
			def:  f.DefValue,
		}
		if r, ok := cfg.short[exportedName(f.Name, cfg)]; ok {
			row.short = r
		}
		if env, ok := cfg.envVars[exportedName(f.Name, cfg)]; ok {
			row.env = env
		}
		left := flagLeft(row.name, row.short, row.kind)
		if len(left) > maxLeft {
			maxLeft = len(left)
		}
		rows = append(rows, row)
	})

	for _, r := range rows {
		left := flagLeft(r.name, r.short, r.kind)
		fmt.Fprintf(w, "  %s%s",
			colorize(left, useColor, ansiYellow),
			strings.Repeat(" ", maxLeft+2-len(left)))
		desc := r.desc
		if desc == "" {
			desc = r.name
		}
		fmt.Fprint(w, desc)
		if r.def != "" && r.def != "false" && r.def != "0" && r.def != "0s" && r.def != "[]" {
			fmt.Fprintf(w, " %s", colorize("(default: "+r.def+")", useColor, ansiDim))
		}
		if r.env != "" {
			fmt.Fprintf(w, " %s", colorize("[env: "+r.env+"]", useColor, ansiDim))
		}
		fmt.Fprintln(w)
	}
}

// exportedName converts a kebab-case flag name back to the field name
// registered with WithFlagShort/WithFlagEnv. cligen records the original
// CamelCase identifier in cfg's maps, so the lookup compares each key's
// kebab-case projection against flagName.
func exportedName(flagName string, cfg regConfig) string {
	for k := range cfg.short {
		if fieldFlagName(k) == flagName {
			return k
		}
	}
	for k := range cfg.envVars {
		if fieldFlagName(k) == flagName {
			return k
		}
	}
	return flagName
}

// flagKind returns a short type label for a flag's value, e.g. "string" or "bool".
// Used in the help page to clarify what the flag accepts.
func flagKind(f *flag.Flag) string {
	if f.DefValue == "false" || f.DefValue == "true" {
		return "" // bool flags are obvious from the no-arg form
	}
	if _, err := fmt.Sscan(f.DefValue, new(int)); err == nil && f.DefValue != "" {
		return "<int>"
	}
	return "<value>"
}

// flagLeft returns the left-column rendering of a flag (short + long + kind).
func flagLeft(name string, short rune, kind string) string {
	var b strings.Builder
	if short != 0 {
		fmt.Fprintf(&b, "-%c, ", short)
	} else {
		b.WriteString("    ")
	}
	fmt.Fprintf(&b, "--%s", name)
	if kind != "" {
		fmt.Fprintf(&b, " %s", kind)
	}
	return b.String()
}

// hasFlags returns whether fs has any flag registered.
func hasFlags(fs *flag.FlagSet) bool {
	count := 0
	fs.VisitAll(func(*flag.Flag) { count++ })
	return count > 0
}

// header returns a section heading. Bold + bright on color writers.
func header(text string, useColor bool) string {
	if !useColor {
		return text
	}
	return ansiBold + text + ansiReset
}

// footer returns the standard footer pointing at per-command help and explain.
func footer(rootName string, useColor bool) string {
	left := fmt.Sprintf("Run %s <command> --help for command-specific help.",
		colorize(rootName, useColor, ansiCyan))
	right := fmt.Sprintf("Run %s <topic> for marker, exit-code, or config-key reference.",
		colorize(rootName+" explain", useColor, ansiUnderline))
	return left + "\n" + right
}

// colorize wraps s in the given ANSI start sequence + reset when useColor is true.
func colorize(s string, useColor bool, start string) string {
	if !useColor {
		return s
	}
	return start + s + ansiReset
}

// ANSI sequences shared with renderRoot/renderSubcommand. Kept short so the
// generated help fits into 80-column terminals comfortably.
const (
	ansiReset     = "\x1b[0m"
	ansiBold      = "\x1b[1m"
	ansiDim       = "\x1b[2m"
	ansiUnderline = "\x1b[4m"
	ansiCyan      = "\x1b[36m"
	ansiYellow    = "\x1b[33m"
)
