// SPDX-License-Identifier: Apache-2.0

package conf

import (
	"log/slog"

	"github.com/nathanbrophy/glacier/option"
)

// loadConfig holds all resolved options for a Load or Decode call.
type loadConfig struct {
	filePath    string            // from WithFile; "" means no file
	envPrefix   string            // from WithEnvPrefix; "" means no env layer
	envSliceSep string            // from WithEnvSliceSep; defaults to ","
	flagSrc     FlagSource        // from WithFlagSource; nil means no flag layer
	sets        map[string]any    // from WithSet; applied last (highest priority)
	defaultsFns []func() map[string]any // from WithDefaults; applied after struct defaults
	logger      *slog.Logger
}

// FlagSource is implemented by callers that want to feed parsed flag values
// into conf.Load. The cli package provides a concrete implementation.
type FlagSource interface {
	// Lookup returns the string value for the given dot-separated field path
	// and reports whether the flag was explicitly set.
	Lookup(path string) (value string, ok bool)
}

// LoadOption is an option applied to Load or Decode.
type LoadOption = option.Option[loadConfig]

// WithFile configures the loader to read key-value pairs from the JSON file at
// the given path. The path is opened with os.Open; it may be absolute or
// relative to the process working directory.
func WithFile(path string) LoadOption {
	return option.OptionFunc[loadConfig](func(c *loadConfig) error {
		c.filePath = path
		return nil
	})
}

// WithEnvPrefix activates the environment-variable layer. Variables must be
// named <PREFIX>__<SECTION>__<FIELD> (all uppercase, double underscore
// separators).
func WithEnvPrefix(prefix string) LoadOption {
	return option.OptionFunc[loadConfig](func(c *loadConfig) error {
		c.envPrefix = prefix
		return nil
	})
}

// WithEnvSliceSep overrides the separator used when splitting an env-var value
// into a slice. Defaults to ",".
func WithEnvSliceSep(sep string) LoadOption {
	return option.OptionFunc[loadConfig](func(c *loadConfig) error {
		c.envSliceSep = sep
		return nil
	})
}

// WithFlagSource registers a FlagSource that is consulted after env vars but
// before WithSet overrides.
func WithFlagSource(fs FlagSource) LoadOption {
	return option.OptionFunc[loadConfig](func(c *loadConfig) error {
		c.flagSrc = fs
		return nil
	})
}

// WithSet overrides the value at the given dot-separated field path with value.
// WithSet is the highest-priority layer; it always wins.
func WithSet(path string, value any) LoadOption {
	return option.OptionFunc[loadConfig](func(c *loadConfig) error {
		if c.sets == nil {
			c.sets = make(map[string]any)
		}
		c.sets[path] = value
		return nil
	})
}

// WithDefaults registers an additional defaults function. The function returns
// a flat map of dot-separated paths to values; keys must include the section
// prefix (e.g., "server.port"). WithDefaults runs after the struct defaults but
// before the JSON file layer.
func WithDefaults(fn func() map[string]any) LoadOption {
	return option.OptionFunc[loadConfig](func(c *loadConfig) error {
		c.defaultsFns = append(c.defaultsFns, fn)
		return nil
	})
}

// WithLogger sets the slog.Logger used for debug-level load events.
// Defaults to slog.Default() when not set.
func WithLogger(l *slog.Logger) LoadOption {
	return option.OptionFunc[loadConfig](func(c *loadConfig) error {
		c.logger = l
		return nil
	})
}
