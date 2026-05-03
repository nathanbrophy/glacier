// SPDX-License-Identifier: Apache-2.0

package conf

import (
	"context"
	"encoding/json"
	"log/slog"
	"reflect"
	"sync"

	"github.com/nathanbrophy/glacier/option"
)

// Loader holds the mutable state for a configuration load session.
// Create one with NewLoader; call Load to populate all registered sections.
type Loader struct {
	mu       sync.Mutex
	isClosed bool
	closed   sync.Once
	defaults []LoadOption // options baked in at construction time
	logger   *slog.Logger
}

// NewLoader creates a Loader. Options provided here are treated as defaults
// and merged (lower priority) with any options passed to each Load call.
func NewLoader(opts ...LoadOption) *Loader {
	return &Loader{defaults: opts}
}

// Load applies configuration sources to all registered sections and
// atomically replaces every registered struct with its newly decoded value.
// Per-call opts take precedence over options passed to NewLoader.
func (l *Loader) Load(ctx context.Context, opts ...LoadOption) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.isClosed {
		return ErrLoaderClosed
	}
	if ctx.Err() != nil {
		return &DecodeError{Cause: ctx.Err(), Layer: "ctx"}
	}

	// Merge constructor defaults with per-call opts (per-call wins).
	allOpts := append(append([]LoadOption(nil), l.defaults...), opts...)
	cfg, err := option.Apply(allOpts)
	if err != nil {
		return &DecodeError{Cause: err, Layer: "options"}
	}
	if cfg.logger == nil {
		cfg.logger = slog.Default()
	}
	if cfg.envSliceSep == "" {
		cfg.envSliceSep = ","
	}

	// Snapshot the registry so we can release the registry lock before decoding.
	globalRegistry.mu.Lock()
	regs := make([]*registration, 0, len(globalRegistry.regs))
	for _, r := range globalRegistry.regs {
		regs = append(regs, r)
	}
	globalRegistry.mu.Unlock()

	type result struct {
		reg *registration
		val any // *T
	}
	results := make([]result, 0, len(regs))

	// Decode phase: produce new *T for every registration.
	for _, reg := range regs {
		newVal, decErr := decodeRegistration(ctx, cfg, reg)
		if decErr != nil {
			return decErr
		}
		results = append(results, result{reg: reg, val: newVal})
	}

	// Commit phase: atomically swap every registration.
	for _, r := range results {
		r.reg.store(r.val)
		cfg.logger.Debug("conf: committed section", "path", r.reg.path)
	}
	return nil
}

// decodeRegistration builds a new *T for reg by applying all configured sources.
func decodeRegistration(ctx context.Context, cfg loadConfig, reg *registration) (any, error) {
	if ctx.Err() != nil {
		return nil, &DecodeError{Cause: ctx.Err(), Layer: "ctx"}
	}

	defType := reflect.TypeOf(reg.defaults)
	if defType.Kind() == reflect.Ptr {
		defType = defType.Elem()
	}

	merged, err := buildMerged(cfg, reg.path, reg.defaults)
	if err != nil {
		return nil, err
	}

	// Layer 5: FlagSource :  requires the struct type for field enumeration.
	if cfg.flagSrc != nil {
		applyFlagSourceToMerged(cfg.flagSrc, reg.path, defType, merged)
	}

	// Layer 6: WithSet overrides.
	applyWithSet(merged, cfg.sets, reg.path)

	data, err := json.Marshal(merged)
	if err != nil {
		return nil, &DecodeError{Path: reg.path, Cause: err, Layer: "marshal"}
	}

	newVal := reflect.New(defType)
	if err := json.Unmarshal(data, newVal.Interface()); err != nil {
		return nil, &DecodeError{Path: reg.path, Cause: err, Layer: "unmarshal"}
	}
	return newVal.Interface(), nil
}

// MustLoad calls Load and panics if Load returns a non-nil error.
func (l *Loader) MustLoad(ctx context.Context, opts ...LoadOption) {
	if err := l.Load(ctx, opts...); err != nil {
		panic(err)
	}
}

// Close marks the Loader as closed. Subsequent Load calls return ErrLoaderClosed.
// Idempotent: multiple Close calls are safe.
func (l *Loader) Close() error {
	l.closed.Do(func() {
		l.mu.Lock()
		l.isClosed = true
		l.mu.Unlock()
	})
	return nil
}

// defaultLoader is the package-level Loader for programs that do not need
// multiple independent load sessions.
var defaultLoader = NewLoader()

// Load applies configuration to all registered sections using the default Loader.
func Load(ctx context.Context, opts ...LoadOption) error {
	return defaultLoader.Load(ctx, opts...)
}

// MustLoad calls the package-level Load and panics on error.
func MustLoad(ctx context.Context, opts ...LoadOption) {
	defaultLoader.MustLoad(ctx, opts...)
}

// Decode applies configuration sources to a single struct of type T without
// using or modifying the global registry. It is safe for concurrent use and
// preferred in tests.
func Decode[T any](ctx context.Context, opts ...LoadOption) (*T, error) {
	if ctx.Err() != nil {
		return nil, &DecodeError{Cause: ctx.Err(), Layer: "ctx"}
	}

	cfg, err := option.Apply(opts)
	if err != nil {
		return nil, &DecodeError{Cause: err, Layer: "options"}
	}
	if cfg.envSliceSep == "" {
		cfg.envSliceSep = ","
	}

	var zero T
	merged, err := buildMerged(cfg, "", &zero)
	if err != nil {
		return nil, err
	}

	if cfg.flagSrc != nil {
		applyFlagSourceToMerged(cfg.flagSrc, "", reflect.TypeOf(zero), merged)
	}

	// WithSet overrides at root level.
	applyWithSet(merged, cfg.sets, "")

	data, err := json.Marshal(merged)
	if err != nil {
		return nil, &DecodeError{Cause: err, Layer: "marshal"}
	}

	result := new(T)
	if err := json.Unmarshal(data, result); err != nil {
		return nil, &DecodeError{Cause: err, Layer: "unmarshal"}
	}
	return result, nil
}
