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

		assert.False(t, bytes.Contains(data, []byte("DefaultTransport")),
			"httpmock production file "+f+" references http.DefaultTransport")
		assert.False(t, bytes.Contains(data, []byte("net.Dial")),
			"httpmock production file "+f+" references net.Dial")
		assert.False(t, bytes.Contains(data, []byte("\"net\"")),
			"httpmock production file "+f+" directly imports the 'net' package")
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

		assert.False(t, bytes.Contains(data, []byte("glacier/fluent")),
			"httpmock production file "+f+" imports fluent (tier violation)")
	}
}
