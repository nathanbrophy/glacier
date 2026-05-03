// SPDX-License-Identifier: Apache-2.0

package gen

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

// checkDrift generates into an in-memory buffer, diffs against the on-disk
// file at path, and returns a descriptive error if stale.
// Returns nil if the on-disk file matches the generated content exactly.
func checkDrift(path string, generated []byte) error {
	onDisk, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("gen: check: generated file %q is missing (run cligen to regenerate)", path)
		}
		return fmt.Errorf("gen: check: read %q: %w", path, err)
	}

	if bytes.Equal(onDisk, generated) {
		return nil
	}

	diff := diffLines(string(onDisk), string(generated))
	return fmt.Errorf("gen: check: %q is stale :  rerun cligen to update:\n%s", path, diff)
}

// diffLines produces a simple line-by-line diff between a and b.
func diffLines(a, b string) string {
	aLines := strings.Split(a, "\n")
	bLines := strings.Split(b, "\n")

	var sb strings.Builder
	max := len(aLines)
	if len(bLines) > max {
		max = len(bLines)
	}

	changed := 0
	for i := range max {
		var al, bl string
		if i < len(aLines) {
			al = aLines[i]
		}
		if i < len(bLines) {
			bl = bLines[i]
		}
		if al != bl {
			changed++
			if changed <= 20 { // cap output
				sb.WriteString(fmt.Sprintf("-%s\n+%s\n", al, bl))
			}
		}
	}
	if changed > 20 {
		sb.WriteString(fmt.Sprintf("... (%d more differing lines)\n", changed-20))
	}
	return sb.String()
}
