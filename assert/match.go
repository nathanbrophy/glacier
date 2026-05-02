// SPDX-License-Identifier: Apache-2.0

package assert

import (
	"regexp"
	"strings"
)

// Match reports whether got matches pattern. Default mode is glob: * matches
// any run of characters, ? matches a single character, pattern is anchored
// at both ends. MatchRegex() switches to regexp.Compile semantics.
// MatchIgnoreCase() makes the match case-insensitive.
//
// On failure reports the pattern, mode, and input via t.Errorf. On regex
// compilation failure, reports the compile error and returns false.
//
// Preconditions: t is non-nil; pattern is non-empty (empty pattern never
// matches any non-empty string; does match an empty string in glob mode).
// Concurrency: goroutine-safe. Compiled regexps are not cached globally;
// callers with hot-path regex matches should compile and reuse themselves.
//
// §21.4 F4, E14, E15, E16
func Match(t TB, got, pattern string, opts ...MatchOption) bool {
	t.Helper()
	cfg := applyMatchOptions(opts)

	if cfg.useRegex {
		return matchRegex(t, got, pattern, cfg.ignoreCase)
	}
	return matchGlob(t, got, pattern, cfg.ignoreCase)
}

func matchRegex(t TB, got, pattern string, ignoreCase bool) bool {
	t.Helper()
	if ignoreCase {
		pattern = "(?i)" + pattern
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		t.Errorf("Match failed: regex compilation error for pattern %q: %s.", pattern, err.Error())
		return false
	}
	if re.MatchString(got) {
		return true
	}
	t.Errorf("Match failed: %q does not match regex %q.", got, pattern)
	return false
}

func matchGlob(t TB, got, pattern string, ignoreCase bool) bool {
	t.Helper()
	cmpGot := got
	cmpPat := pattern
	if ignoreCase {
		cmpGot = strings.ToLower(cmpGot)
		cmpPat = strings.ToLower(cmpPat)
	}
	if globMatch(cmpPat, cmpGot) {
		return true
	}
	t.Errorf("Match failed: %q does not match glob %q.", got, pattern)
	return false
}

// globMatch returns true if pattern matches s.
// Supports * (any run of characters) and ? (single character).
// Pattern is implicitly anchored at both ends.
func globMatch(pattern, s string) bool {
	// dp-style recursive matching with memoisation via simple recursion.
	// For the sizes seen in tests this is fast enough.
	return globMatchRec(pattern, s)
}

func globMatchRec(pat, s string) bool {
	for len(pat) > 0 {
		switch pat[0] {
		case '*':
			// Skip consecutive stars.
			for len(pat) > 0 && pat[0] == '*' {
				pat = pat[1:]
			}
			// Empty star at end matches everything.
			if len(pat) == 0 {
				return true
			}
			// Try matching pat at every position in s.
			for i := range len(s) + 1 {
				if globMatchRec(pat, s[i:]) {
					return true
				}
			}
			return false
		case '?':
			if len(s) == 0 {
				return false
			}
			pat = pat[1:]
			s = s[1:]
		default:
			if len(s) == 0 || pat[0] != s[0] {
				return false
			}
			pat = pat[1:]
			s = s[1:]
		}
	}
	return len(s) == 0
}
