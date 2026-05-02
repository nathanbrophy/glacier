// SPDX-License-Identifier: Apache-2.0

package httpmock_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
)

// TestNoNetworkImports verifies that httpmock production source files do not
// reference http.DefaultTransport or net.Dial. We check the source directly
// rather than via go list, because the 'net' package itself is always in the
// transitive closure of net/http (which httpmock must import for type definitions).
func TestNoNetworkImports(t *testing.T) {
	// Read all non-test .go files under httpmock/.
	files, err := filepath.Glob(filepath.Join(".", "*.go"))
	assert.NoError(t, err)

	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		data, err := os.ReadFile(f)
		assert.NoError(t, err)

		if bytes.Contains(data, []byte("DefaultTransport")) {
			t.Errorf("httpmock production file %s references http.DefaultTransport", f)
		}
		if bytes.Contains(data, []byte("net.Dial")) {
			t.Errorf("httpmock production file %s references net.Dial", f)
		}
		if bytes.Contains(data, []byte("\"net\"")) {
			t.Errorf("httpmock production file %s directly imports the 'net' package", f)
		}
	}
}

func TestImportSurfaceFluentInternal(t *testing.T) {
	// Verify fluent is not in the direct import set of httpmock production code.
	// Per tier constraints, httpmock (Tier 2 leaf) should not import fluent (Tier 1)
	// in production code.
	files, err := filepath.Glob(filepath.Join(".", "*.go"))
	assert.NoError(t, err)

	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		data, err := os.ReadFile(f)
		assert.NoError(t, err)

		if bytes.Contains(data, []byte("glacier/fluent")) {
			t.Errorf("httpmock production file %s imports fluent (tier violation)", f)
		}
	}
}
