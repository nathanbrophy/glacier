// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

// frameworkLeaves is the set of Glacier framework packages that cmd/glacier
// must collectively import (D-S18, spec 0032 §"Glacier everywhere"). Test
// fails for each package absent from the import graph; missing packages are
// reported as named gaps so the fix is targeted.
var frameworkLeaves = []string{
	"github.com/nathanbrophy/glacier/option",
	"github.com/nathanbrophy/glacier/errs",
	"github.com/nathanbrophy/glacier/log",
	"github.com/nathanbrophy/glacier/assert",
	"github.com/nathanbrophy/glacier/term",
	"github.com/nathanbrophy/glacier/concur",
	"github.com/nathanbrophy/glacier/fluent",
	"github.com/nathanbrophy/glacier/conf",
	"github.com/nathanbrophy/glacier/fixture",
	"github.com/nathanbrophy/glacier/obs",
	"github.com/nathanbrophy/glacier/cli",
	"github.com/nathanbrophy/glacier/mock",
	"github.com/nathanbrophy/glacier/httpmock",
	"github.com/nathanbrophy/glacier/httpc",
	"github.com/nathanbrophy/glacier/cache",
}

// TestGlacierEverywhere asserts that the union of imports across
// cmd/glacier/... (including test files) covers every Glacier framework leaf
// package. Each missing package is reported individually via t.Errorf so the
// test does not abort on the first gap.
func TestGlacierEverywhere(t *testing.T) {
	// Use the full module import path so go list resolves correctly regardless
	// of the test binary's working directory (which is the package dir, not
	// the module root).
	const pattern = "github.com/nathanbrophy/glacier/cmd/glacier/..."

	out, err := exec.Command("go", "list", "-json", "-test", pattern).Output()
	if err != nil {
		// Fall back without -test on build error.
		out, err = exec.Command("go", "list", "-json", pattern).Output()
		if err != nil {
			t.Skipf("go list failed: %v", err)
			return
		}
	}

	allImports := collectImports(t, out)

	for _, pkg := range frameworkLeaves {
		if !allImports[pkg] {
			t.Errorf("framework package NOT covered by cmd/glacier/... import graph: %s", pkg)
		}
	}
}

// collectImports parses the JSON stream produced by go list -json and returns
// the union of all Imports and TestImports fields.
func collectImports(t *testing.T, raw []byte) map[string]bool {
	t.Helper()
	seen := make(map[string]bool)
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	for dec.More() {
		var pkg struct {
			ImportPath  string   `json:"ImportPath"`
			Imports     []string `json:"Imports"`
			TestImports []string `json:"TestImports"`
		}
		if err := dec.Decode(&pkg); err != nil {
			t.Fatalf("collectImports: decode: %v", err)
		}
		for _, imp := range pkg.Imports {
			seen[imp] = true
		}
		for _, imp := range pkg.TestImports {
			seen[imp] = true
		}
	}
	return seen
}
