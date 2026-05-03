// SPDX-License-Identifier: Apache-2.0

package assert

import "testing"

// §21.4 F4, E14, E15, E16

func TestMatchGlob(t *testing.T) {
	mt := &mockTB{}
	True(t, Match(mt, "hello world", "hello *"), "glob: 'hello *' matches 'hello world'")
}

func TestMatchGlobAnchors(t *testing.T) {
	mt := &mockTB{}
	// Glob is anchored: "abc" does NOT match "a" (no wildcard).
	False(t, Match(mt, "abc", "a"), "glob: 'a' does not match 'abc' (anchored)")
	mt.reset()
	// But "a*" matches "abc".
	True(t, Match(mt, "abc", "a*"), "glob: 'a*' matches 'abc'")
}

func TestMatchGlobSingleChar(t *testing.T) {
	mt := &mockTB{}
	True(t, Match(mt, "abc", "a?c"), "glob: 'a?c' matches 'abc'")
	mt.reset()
	False(t, Match(mt, "abbc", "a?c"), "glob: 'a?c' does not match 'abbc'")
}

func TestMatchRegex(t *testing.T) {
	mt := &mockTB{}
	True(t, Match(mt, "abc", `^[a-c]+$`, MatchRegex()), "regex: matches")
	mt.reset()
	False(t, Match(mt, "abc", `^[0-9]+$`, MatchRegex()), "regex: no match")
}

func TestMatchIgnoreCase(t *testing.T) {
	mt := &mockTB{}
	True(t, Match(mt, "Hello", "hello", MatchIgnoreCase()), "glob IgnoreCase: Hello matches hello")
	mt.reset()
	True(t, Match(mt, "user-12345", "USER-*", MatchIgnoreCase()), "glob IgnoreCase: user-12345 matches USER-*")
}

func TestMatchIgnoreCaseRegex(t *testing.T) {
	mt := &mockTB{}
	True(t, Match(mt, "Hello", "hello", MatchRegex(), MatchIgnoreCase()), "regex IgnoreCase: Hello matches hello")
}

func TestMatchInvalidRegex(t *testing.T) {
	mt := &mockTB{}
	False(t, Match(mt, "abc", `[invalid`, MatchRegex()), "invalid regex reports error")
	Equal(t, mt.errorfCalls, 1)
}

func TestMatchSpecialCharsGlob(t *testing.T) {
	mt := &mockTB{}
	// In glob mode, '.' is literal (not regex).
	// "a.b" glob pattern: a, then literal '.', then b.
	True(t, Match(mt, "a.b", "a.b"), "glob: literal dot matches literal dot")
	mt.reset()
	// "a.b" does not match "axb" in glob mode (dot is literal).
	False(t, Match(mt, "axb", "a.b"), "glob: literal dot does not match non-dot char")
}

// L-add-8: Match("", "") :  empty pattern matches empty string in glob mode.
func TestMatchEmptyPatternEmptyString(t *testing.T) {
	mt := &mockTB{}
	True(t, Match(mt, "", ""), "glob: empty pattern matches empty string")
}

func TestMatchEmptyPatternNonEmpty(t *testing.T) {
	mt := &mockTB{}
	False(t, Match(mt, "abc", ""), "glob: empty pattern does not match non-empty string")
}

func TestMatchUserIDExample(t *testing.T) {
	mt := &mockTB{}
	True(t, Match(mt, "user-12345", "user-*"), "glob: user-12345 matches user-*")
}
