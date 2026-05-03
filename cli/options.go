// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/nathanbrophy/glacier/option"
)

var (
	nameRe     = regexp.MustCompile(`^[a-z][a-z0-9-]{0,31}$`)
	envVarRe   = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
	choiceRe   = regexp.MustCompile(`^[a-z0-9_-]+$`)
	parentSegs = regexp.MustCompile(`^([a-z][a-z0-9-]{0,31}\.)*[a-z][a-z0-9-]{0,31}$`)
)

// WithVersion sets the version string reported by --version / -v.
// If not set, cli falls back to runtime/debug.ReadBuildInfo().
func WithVersion(v string) option.Option[appConfig] {
	return option.OptionFunc[appConfig](func(c *appConfig) error {
		c.version = v
		return nil
	})
}

// WithStdout redirects all user-facing output (help, banner, version) to w.
// Precondition: w must not be nil.
func WithStdout(w io.Writer) option.Option[appConfig] {
	return option.OptionFunc[appConfig](func(c *appConfig) error {
		if w == nil {
			return errors.New("cli: WithStdout: writer is nil")
		}
		c.stdout = w
		return nil
	})
}

// WithStderr redirects error output to w.
// Precondition: w must not be nil.
func WithStderr(w io.Writer) option.Option[appConfig] {
	return option.OptionFunc[appConfig](func(c *appConfig) error {
		if w == nil {
			return errors.New("cli: WithStderr: writer is nil")
		}
		c.stderr = w
		return nil
	})
}

// WithLogger sets the slog.Logger used for lifecycle events.
// Precondition: l must not be nil.
func WithLogger(l *slog.Logger) option.Option[appConfig] {
	return option.OptionFunc[appConfig](func(c *appConfig) error {
		if l == nil {
			return errors.New("cli: WithLogger: logger is nil")
		}
		c.logger = l
		return nil
	})
}

// WithoutBanner suppresses the automatic banner display on bare invocation. (§23.15)
// The Banner() method remains available for explicit calls.
func WithoutBanner() option.Option[appConfig] {
	return option.OptionFunc[appConfig](func(c *appConfig) error {
		c.noBanner = true
		return nil
	})
}

// WithName sets the command name as it appears in argv.
// Precondition: name must match ^[a-z][a-z0-9-]{0,31}$.
func WithName(name string) option.Option[regConfig] {
	return option.OptionFunc[regConfig](func(c *regConfig) error {
		if !nameRe.MatchString(name) {
			return fmt.Errorf("cli: WithName: %q does not match ^[a-z][a-z0-9-]{0,31}$", name)
		}
		c.name = name
		return nil
	})
}

// WithParent sets the dot-separated path to the parent command.
// Precondition: each segment of path must satisfy the name regex.
func WithParent(path string) option.Option[regConfig] {
	return option.OptionFunc[regConfig](func(c *regConfig) error {
		if !parentSegs.MatchString(path) {
			return fmt.Errorf("cli: WithParent: %q is not a valid dot-separated path", path)
		}
		c.parent = path
		return nil
	})
}

// WithAlias adds an additional name that routes to this command.
// Multiple WithAlias calls accumulate. Each alias must satisfy the name regex.
func WithAlias(alias string) option.Option[regConfig] {
	return option.OptionFunc[regConfig](func(c *regConfig) error {
		if !nameRe.MatchString(alias) {
			return fmt.Errorf("cli: WithAlias: %q does not match ^[a-z][a-z0-9-]{0,31}$", alias)
		}
		c.aliases = append(c.aliases, alias)
		return nil
	})
}

// WithRoot marks this command as the tree root.
// Only one command per App may be root; Register returns *ErrMultipleRoots on conflict.
func WithRoot() option.Option[regConfig] {
	return option.OptionFunc[regConfig](func(c *regConfig) error {
		c.root = true
		return nil
	})
}

// WithSummary sets the one-line description shown next to the command in
// the top-level help listing. Newlines are converted to spaces.
func WithSummary(summary string) option.Option[regConfig] {
	return option.OptionFunc[regConfig](func(c *regConfig) error {
		c.summary = strings.ReplaceAll(summary, "\n", " ")
		return nil
	})
}

// WithCategory groups commands in the top-level help. Conventional values
// for the SDK are "create", "develop", "inspect", "utility"; any short
// lowercase string works. Unknown categories are still accepted; help
// renders them in registration order at the bottom of the listing.
func WithCategory(category string) option.Option[regConfig] {
	return option.OptionFunc[regConfig](func(c *regConfig) error {
		c.category = category
		return nil
	})
}

// WithLongDescription sets the multi-line description rendered below the
// synopsis on a per-command help page (e.g. `glacier vibe --help`).
// Use plain text; ANSI codes are stripped on non-color writers.
func WithLongDescription(desc string) option.Option[regConfig] {
	return option.OptionFunc[regConfig](func(c *regConfig) error {
		c.longDesc = desc
		return nil
	})
}

// WithFlagShort registers a single-character shorthand for the named flag.
// Precondition: short satisfies [a-zA-Z] (single ASCII letter).
func WithFlagShort(name string, short rune) option.Option[regConfig] {
	return option.OptionFunc[regConfig](func(c *regConfig) error {
		if !utf8.ValidRune(short) || short > unicode.MaxASCII {
			return fmt.Errorf("cli: WithFlagShort: short rune %q is not ASCII", short)
		}
		if !((short >= 'a' && short <= 'z') || (short >= 'A' && short <= 'Z')) {
			return fmt.Errorf("cli: WithFlagShort: short rune %q is not [a-zA-Z]", short)
		}
		if c.short == nil {
			c.short = make(map[string]rune)
		}
		c.short[name] = short
		return nil
	})
}

// WithFlagEnv binds the named flag to an environment variable.
// Precondition: envVar satisfies ^[A-Z][A-Z0-9_]*$.
func WithFlagEnv(name string, envVar string) option.Option[regConfig] {
	return option.OptionFunc[regConfig](func(c *regConfig) error {
		if !envVarRe.MatchString(envVar) {
			return fmt.Errorf("cli: WithFlagEnv: env var %q does not match ^[A-Z][A-Z0-9_]*$", envVar)
		}
		if c.envVars == nil {
			c.envVars = make(map[string]string)
		}
		c.envVars[name] = envVar
		return nil
	})
}

// WithFlagHelp overrides the help string for the named flag.
func WithFlagHelp(name string, help string) option.Option[regConfig] {
	return option.OptionFunc[regConfig](func(c *regConfig) error {
		if c.help == nil {
			c.help = make(map[string]string)
		}
		c.help[name] = stripANSI(help)
		return nil
	})
}

// WithFlagRequired marks the named flag as required.
func WithFlagRequired(name string) option.Option[regConfig] {
	return option.OptionFunc[regConfig](func(c *regConfig) error {
		if c.required == nil {
			c.required = make(map[string]struct{})
		}
		c.required[name] = struct{}{}
		return nil
	})
}

// WithFlagChoices constrains the named flag to an enumerated set of values.
// Precondition: each choice must satisfy ^[a-z0-9_-]+$; at most 32 choices.
func WithFlagChoices(name string, choices ...string) option.Option[regConfig] {
	return option.OptionFunc[regConfig](func(c *regConfig) error {
		if len(choices) > 32 {
			return fmt.Errorf("cli: WithFlagChoices: too many choices (%d > 32)", len(choices))
		}
		for _, ch := range choices {
			if !choiceRe.MatchString(ch) {
				return fmt.Errorf("cli: WithFlagChoices: choice %q does not match ^[a-z0-9_-]+$", ch)
			}
		}
		if c.choices == nil {
			c.choices = make(map[string][]string)
		}
		c.choices[name] = choices
		return nil
	})
}

// WithFlagDeprecated marks the named flag as deprecated.
// When a deprecated flag is used, a warning is logged before the handler is called.
func WithFlagDeprecated(name string, message string) option.Option[regConfig] {
	return option.OptionFunc[regConfig](func(c *regConfig) error {
		if c.deprecated == nil {
			c.deprecated = make(map[string]string)
		}
		c.deprecated[name] = message
		return nil
	})
}

// WithFlagValidate registers a validation function for the named flag.
// The function is called after all flags are bound and env vars applied, before Run.
func WithFlagValidate(name string, fn func(string) error) option.Option[regConfig] {
	return option.OptionFunc[regConfig](func(c *regConfig) error {
		if fn == nil {
			return fmt.Errorf("cli: WithFlagValidate: fn is nil for flag %q", name)
		}
		if c.validate == nil {
			c.validate = make(map[string]func(string) error)
		}
		c.validate[name] = fn
		return nil
	})
}
