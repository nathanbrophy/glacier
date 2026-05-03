// SPDX-License-Identifier: Apache-2.0

package commands

import "github.com/nathanbrophy/glacier/cli"

// init annotates every registered SDK command with its help metadata
// (summary line, category for the top-level grouping, multi-line long
// description). The cligen-generated zz_generated_cli.go calls Register
// with positional + flag options only; help text is layered on here so
// the SDK ships a complete `glacier --help` page without changing the
// marker grammar in spec 0011.
//
// Order: this init runs after the generated init() (Go orders init() funcs
// within a package by file name; "help_metadata.go" sorts after
// "zz_generated_cli.go" alphabetically). If that ordering ever flips, move
// this into a function called from main and invoke it explicitly.
func init() {
	_ = cli.Default.Annotate("glacier",
		cli.WithSummary("Less plumbing. More Go."),
		cli.WithLongDescription(`The Glacier SDK is the framework's longest-running integration test
and the public face for new developers. Nine commands cover the
developer day: scaffolding, codegen, lint, test, animation, and
reference lookup.`),
	)

	// CREATE: scaffold new code.
	_ = cli.Default.Annotate("init",
		cli.WithCategory("create"),
		cli.WithSummary("Scaffold a new Glacier project."),
		cli.WithLongDescription(`Walks the user through choosing a template, license, and mascot,
then writes a minimal Go module wired up with the framework's
batteries (signals, banner, version, completions).

Use --yes to accept all defaults non-interactively (cli-app template,
Apache-2.0, polar bear).`),
	)
	_ = cli.Default.Annotate("new",
		cli.WithCategory("create"),
		cli.WithSummary("Add a package, command, or option to an existing project."),
		cli.WithLongDescription(`Subcommands:
  package <name>    new Go package skeleton
  command <name>    new +glacier:command struct (regenerates cli)
  option <type>     new functional-option constructor`),
	)
	_ = cli.Default.Annotate("new.package",
		cli.WithSummary("Scaffold a new Go package skeleton."),
	)
	_ = cli.Default.Annotate("new.command",
		cli.WithSummary("Scaffold a new +glacier:command struct and regenerate cli."),
	)
	_ = cli.Default.Annotate("new.option",
		cli.WithSummary("Append a new functional-option constructor."),
	)

	// DEVELOP: produce, lint, test code.
	_ = cli.Default.Annotate("generate",
		cli.WithCategory("develop"),
		cli.WithSummary("Run all registered code generators (cli, mock, httpmock)."),
		cli.WithLongDescription(`Discovers every type with a +glacier:* marker in the targeted
packages and emits the corresponding zz_generated_*.go files.

Use --check in CI to detect drift; the command exits 69 when any
generated file is stale.`),
	)
	_ = cli.Default.Annotate("lint",
		cli.WithCategory("develop"),
		cli.WithSummary("Run gofmt + go vet + staticcheck + 6 Glacier-specific lints."),
		cli.WithLongDescription(`Glacier-specific lints:
  exported-doc-comment   every exported symbol has a doc comment
  package-example-test   every package has at least one Example*()
  panic-in-library       no panics in non-cmd library code
  no-em-dash             no U+2014 em-dash characters
  library-error-register every error string conforms to the register
  naked-any              opt-in: flag bare any/interface{}

Use --fix to auto-correct gofmt + em-dash + marker-normalization findings.`),
	)
	_ = cli.Default.Annotate("test",
		cli.WithCategory("develop"),
		cli.WithSummary("Run go test with a streaming summary, bench baseline, JUnit/SARIF emitters."),
		cli.WithLongDescription(`Wraps go test -json and renders a rolling status panel during the
run. Supports --format=junit, --format=sarif, and --format=json
(forwards every event plus a final glacier-summary aggregate).

Bench baseline lives at .glacier/bench-baseline.json. A regression
of more than 5% on any benchmark exits 66; --update-baseline writes
the current run's numbers as the new baseline.`),
	)

	// INSPECT: look up information.
	_ = cli.Default.Annotate("explain",
		cli.WithCategory("inspect"),
		cli.WithSummary("Show reference for a marker, exit code, or config key."),
		cli.WithLongDescription(`Lists or shows reference topics for the framework's markers, the
SDK's exit codes, and the SDK's config keys.

Examples:
  glacier explain --list
  glacier explain exit:66
  glacier explain +glacier:command
  glacier explain config:github.repo`),
	)
	_ = cli.Default.Annotate("version",
		cli.WithCategory("inspect"),
		cli.WithSummary("Print version info; --check fetches the latest release."),
		cli.WithLongDescription(`Prints the binary's version, build time, Go version, and OS/arch.

Use --check to fetch the latest release from GitHub. The result is
cached for 24h under <UserCacheDir>/glacier/. --strict exits 68
when the API is unreachable; without --strict, version --check
exits 0 with an (offline) annotation.`),
	)

	// UTILITY: dev-loop helpers.
	_ = cli.Default.Annotate("completions",
		cli.WithCategory("utility"),
		cli.WithSummary("Print a shell completion script (bash, zsh, fish, pwsh)."),
		cli.WithLongDescription(`Emits a shell completion script to stdout. Source it from your
shell's startup file:

  # bash
  source <(glacier completions bash)
  # zsh
  source <(glacier completions zsh)
  # fish
  glacier completions fish | source
  # PowerShell
  glacier completions pwsh | Out-String | Invoke-Expression`),
	)
	_ = cli.Default.Annotate("vibe",
		cli.WithCategory("utility"),
		cli.WithSummary("Animated bear + tip rotation; ambient framework reference."),
		cli.WithLongDescription(`Plays an animated polar bear with the Glacier wordmark and a
rotating tip from the framework's documentation. Press any key to
exit.

Use --ascii on terminals without UTF-8 or color support;
use --duration to bound the animation; use --no-tips to suppress
the tip footer.`),
	)
}
