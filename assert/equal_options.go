// SPDX-License-Identifier: Apache-2.0

package assert

// equalConfig holds the resolved configuration for a single Equal call.
// Constructed by applying EqualOption values; zero value is the default
// (order-sensitive, case-sensitive, no delta, no ignored fields).
type equalConfig struct {
	ignoreOrder      bool
	ignoreCase       bool
	ignoreWhitespace bool
	deltaSet         bool
	// invariant: delta >= 0; negative values rejected by WithDelta constructor.
	delta float64
	// invariant: ignoreFields entries are non-empty field names; duplicates
	// are silently collapsed (field is ignored once regardless of count).
	ignoreFields map[string]bool
}

// EqualOption configures smart-equal semantics for Equal, NotEqual,
// Contains, JSONEq, and Subset. Options compose; conflicting options
// (e.g., IgnoreCase on a struct with no string fields) are silently
// no-ops for the irrelevant fields.
type EqualOption interface{ applyEqual(*equalConfig) }

// equalOptFunc is the internal adapter for EqualOption.
type equalOptFunc func(*equalConfig)

func (f equalOptFunc) applyEqual(c *equalConfig) { f(c) }

// applyEqualOptions applies opts to a zero-valued equalConfig and returns it.
func applyEqualOptions(opts []EqualOption) equalConfig {
	var c equalConfig
	for _, o := range opts {
		if o != nil {
			o.applyEqual(&c)
		}
	}
	return c
}

// IgnoreOrder returns an EqualOption that compares slices and arrays as
// multisets: every element in want must appear in got with the same count,
// regardless of position. Maps are always order-insensitive; this option
// does not affect them. Does not affect struct field order.
//
// Preconditions: none.
// Concurrency: option values are immutable; safe to share across goroutines.
func IgnoreOrder() EqualOption {
	return equalOptFunc(func(c *equalConfig) { c.ignoreOrder = true })
}

// IgnoreCase returns an EqualOption that compares strings using
// strings.EqualFold (Unicode-aware case folding). Applied to string values,
// map keys of type string, and struct fields of type string, recursively.
//
// Preconditions: none.
// Concurrency: immutable; goroutine-safe.
func IgnoreCase() EqualOption {
	return equalOptFunc(func(c *equalConfig) { c.ignoreCase = true })
}

// IgnoreWhitespace returns an EqualOption that normalizes strings before
// comparison: leading and trailing whitespace trimmed, internal runs of
// whitespace collapsed to a single space. Applied recursively to all string
// values in the comparison graph.
//
// Preconditions: none.
// Concurrency: immutable; goroutine-safe.
func IgnoreWhitespace() EqualOption {
	return equalOptFunc(func(c *equalConfig) { c.ignoreWhitespace = true })
}

// WithDelta returns an EqualOption that compares float32 and float64 values
// using absolute tolerance: |got - want| <= d. Applied to float fields inside
// structs and to slice elements. NaN values are never equal regardless of
// delta (Go semantics: NaN != NaN).
//
// Preconditions: d >= 0. Negative d is rejected; Equal reports a test error
// and returns false without comparing values.
// Concurrency: immutable; goroutine-safe.
func WithDelta(d float64) EqualOption {
	return equalOptFunc(func(c *equalConfig) {
		c.deltaSet = true
		c.delta = d
	})
}

// IgnoreFields returns an EqualOption that skips the named struct fields
// during comparison. Names match exported struct field names exactly
// (case-sensitive). Fields that do not exist on the compared type are
// silently ignored. Applied recursively: a field named "Created" is skipped
// at every nesting level where it appears.
//
// Preconditions: names must be non-empty. An empty name string is silently
// skipped (not an error).
// Concurrency: immutable; goroutine-safe.
func IgnoreFields(names ...string) EqualOption {
	return equalOptFunc(func(c *equalConfig) {
		if c.ignoreFields == nil {
			c.ignoreFields = make(map[string]bool)
		}
		for _, n := range names {
			if n != "" {
				c.ignoreFields[n] = true
			}
		}
	})
}

// matchConfig holds the resolved configuration for a single Match call.
// Zero value is glob mode, case-sensitive.
type matchConfig struct {
	useRegex   bool
	ignoreCase bool
}

// MatchOption configures Match semantics.
type MatchOption interface{ applyMatch(*matchConfig) }

// matchOptFunc is the internal adapter for MatchOption.
type matchOptFunc func(*matchConfig)

func (f matchOptFunc) applyMatch(c *matchConfig) { f(c) }

// applyMatchOptions applies opts to a zero-valued matchConfig and returns it.
func applyMatchOptions(opts []MatchOption) matchConfig {
	var c matchConfig
	for _, o := range opts {
		if o != nil {
			o.applyMatch(&c)
		}
	}
	return c
}

// MatchRegex returns a MatchOption that switches Match from glob to
// regexp syntax. The pattern is compiled once via regexp.Compile; a
// compilation failure is reported via t.Errorf and Match returns false.
//
// Concurrency: immutable; goroutine-safe.
func MatchRegex() MatchOption {
	return matchOptFunc(func(c *matchConfig) { c.useRegex = true })
}

// MatchIgnoreCase returns a MatchOption that makes the pattern match
// case-insensitively. For glob mode, both pattern and input are folded to
// lowercase before matching. For regex mode, the (?i) flag is prepended to
// the pattern.
//
// Concurrency: immutable; goroutine-safe.
func MatchIgnoreCase() MatchOption {
	return matchOptFunc(func(c *matchConfig) { c.ignoreCase = true })
}
