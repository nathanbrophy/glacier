// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"os"
	"strings"
)

// positionals returns the non-flag arguments from os.Args[1:] that come after
// the command tokens. skip is the number of command-name tokens to skip (1 for
// top-level commands, 2 for sub-sub-commands like "new command").
//
// Example: os.Args = [glacier, explain, 66] → positionals(1) = ["66"]
// Example: os.Args = [glacier, new, command, pause] → positionals(2) = ["pause"]
func positionals(skip int) []string {
	var result []string
	skipped := 0
	for _, a := range os.Args[1:] {
		if strings.HasPrefix(a, "-") {
			continue
		}
		skipped++
		if skipped <= skip {
			continue // skip command name tokens
		}
		result = append(result, a)
	}
	return result
}

// firstPositional returns the first positional argument after the command
// tokens, or "" if there is none.
func firstPositional(skip int) string {
	pos := positionals(skip)
	if len(pos) == 0 {
		return ""
	}
	return pos[0]
}
