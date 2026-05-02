// SPDX-License-Identifier: Apache-2.0

package assert

import "testing"

// §21.4 F12, F13, F14, F15, F16; §23.5

func TestIgnoreOrderOption(t *testing.T) {
	opt := IgnoreOrder()
	if opt == nil {
		t.Fatal("IgnoreOrder() returned nil")
	}
	cfg := applyEqualOptions([]EqualOption{opt})
	True(t, cfg.ignoreOrder, "IgnoreOrder: ignoreOrder flag set")
}

func TestIgnoreCaseOption(t *testing.T) {
	opt := IgnoreCase()
	cfg := applyEqualOptions([]EqualOption{opt})
	True(t, cfg.ignoreCase, "IgnoreCase: ignoreCase flag set")
}

func TestIgnoreWhitespaceOption(t *testing.T) {
	opt := IgnoreWhitespace()
	cfg := applyEqualOptions([]EqualOption{opt})
	True(t, cfg.ignoreWhitespace, "IgnoreWhitespace: ignoreWhitespace flag set")
}

func TestWithDeltaOption(t *testing.T) {
	opt := WithDelta(0.5)
	cfg := applyEqualOptions([]EqualOption{opt})
	True(t, cfg.deltaSet, "WithDelta: deltaSet flag set")
	True(t, cfg.delta == 0.5, "WithDelta: delta value set")
}

func TestWithDeltaNegativeRejected(t *testing.T) {
	mt := &mockTB{}
	// L-add-6: WithDelta(d < 0) reports error and returns false.
	result := Equal(mt, 1.0, 1.0, WithDelta(-0.1))
	False(t, result, "WithDelta(-0.1): should return false")
	True(t, mt.errorfCalls == 1, "WithDelta(-0.1): Errorf called once")
}

func TestIgnoreFieldsOption(t *testing.T) {
	opt := IgnoreFields("Foo", "Bar")
	cfg := applyEqualOptions([]EqualOption{opt})
	NotNil(t, cfg.ignoreFields, "IgnoreFields: map is non-nil")
	True(t, cfg.ignoreFields["Foo"], "IgnoreFields: Foo present")
	True(t, cfg.ignoreFields["Bar"], "IgnoreFields: Bar present")
}

func TestIgnoreFieldsEmptyNameSkipped(t *testing.T) {
	opt := IgnoreFields("", "Valid")
	cfg := applyEqualOptions([]EqualOption{opt})
	_, hasEmpty := cfg.ignoreFields[""]
	False(t, hasEmpty, "IgnoreFields: empty name is skipped")
	True(t, cfg.ignoreFields["Valid"], "IgnoreFields: Valid is present")
}

func TestIgnoreFieldsDeduplicated(t *testing.T) {
	opt := IgnoreFields("Foo", "Foo")
	cfg := applyEqualOptions([]EqualOption{opt})
	// map deduplicates naturally.
	True(t, cfg.ignoreFields["Foo"], "IgnoreFields: duplicate silently collapsed")
}

func TestOptionsCompose(t *testing.T) {
	cfg := applyEqualOptions([]EqualOption{IgnoreOrder(), IgnoreCase(), WithDelta(0.1)})
	True(t, cfg.ignoreOrder, "compose: ignoreOrder")
	True(t, cfg.ignoreCase, "compose: ignoreCase")
	True(t, cfg.deltaSet, "compose: deltaSet")
}

func TestMatchRegexOption(t *testing.T) {
	opt := MatchRegex()
	cfg := applyMatchOptions([]MatchOption{opt})
	True(t, cfg.useRegex, "MatchRegex: useRegex set")
}

func TestMatchIgnoreCaseOption(t *testing.T) {
	opt := MatchIgnoreCase()
	cfg := applyMatchOptions([]MatchOption{opt})
	True(t, cfg.ignoreCase, "MatchIgnoreCase: ignoreCase set")
}
