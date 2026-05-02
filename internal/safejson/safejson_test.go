// SPDX-License-Identifier: Apache-2.0

package safejson_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	safejson "github.com/nathanbrophy/glacier/internal/safejson"
)

// buildNestedJSON builds a JSON string with depth levels of object nesting.
func buildNestedJSON(depth int) string {
	var b strings.Builder
	for range depth {
		b.WriteString(`{"x":`)
	}
	b.WriteString("1")
	for range depth {
		b.WriteString("}")
	}
	return b.String()
}

func TestDecode_Happy(t *testing.T) {
	t.Parallel()
	input := `{"name":"glacier","version":1}`
	var dst map[string]any
	require.NoError(t, safejson.Decode(strings.NewReader(input), &dst))
	assert.Equal(t, dst["name"], any("glacier"))
	assert.Equal(t, dst["version"], any(float64(1)))
}

func TestDecode_TooLarge(t *testing.T) {
	t.Parallel()
	// Build a byte slice slightly larger than MaxFileSize.
	data := make([]byte, safejson.MaxFileSize+1)
	// Make it look like a JSON string so the size guard fires before parse.
	data[0] = '"'
	for i := int64(1); i < safejson.MaxFileSize; i++ {
		data[i] = 'a'
	}
	data[safejson.MaxFileSize] = '"'
	var dst any
	err := safejson.Decode(bytes.NewReader(data), &dst)
	assert.ErrorIs(t, err, safejson.ErrTooLarge)
}

func TestDecode_DepthExceeded(t *testing.T) {
	t.Parallel()
	// MaxDepth+1 levels of nesting must be rejected.
	deep := buildNestedJSON(safejson.MaxDepth + 1)
	var dst any
	err := safejson.Decode(strings.NewReader(deep), &dst)
	assert.ErrorIs(t, err, safejson.ErrDepthExceeded)
}

func TestDecode_MaxDepthAllowed(t *testing.T) {
	t.Parallel()
	// Exactly MaxDepth levels of nesting must be accepted.
	at := buildNestedJSON(safejson.MaxDepth)
	var dst any
	require.NoError(t, safejson.Decode(strings.NewReader(at), &dst))
}

func TestDecode_InvalidJSON(t *testing.T) {
	t.Parallel()
	var dst map[string]any
	err := safejson.Decode(strings.NewReader(`{bad}`), &dst)
	assert.Error(t, err)
	// Must NOT be a safejson sentinel — it should be a standard json error.
	assert.True(t, !errors.Is(err, safejson.ErrTooLarge))
	assert.True(t, !errors.Is(err, safejson.ErrDepthExceeded))
}

func TestDecodeStrict_RejectsUnknownFields(t *testing.T) {
	t.Parallel()
	type cfg struct {
		Name string `json:"name"`
	}
	input := `{"name":"x","unknown":true}`
	var dst cfg
	err := safejson.DecodeStrict(strings.NewReader(input), &dst)
	assert.Error(t, err)
}

func TestDecodeStrict_Happy(t *testing.T) {
	t.Parallel()
	type cfg struct {
		Name string `json:"name"`
	}
	input := `{"name":"glacier"}`
	var dst cfg
	require.NoError(t, safejson.DecodeStrict(strings.NewReader(input), &dst))
	assert.Equal(t, dst.Name, "glacier")
}

func TestDecodeStrict_TooLarge(t *testing.T) {
	t.Parallel()
	data := make([]byte, safejson.MaxFileSize+2)
	data[0] = '"'
	for i := int64(1); i < safejson.MaxFileSize; i++ {
		data[i] = 'a'
	}
	data[safejson.MaxFileSize] = '"'
	data[safejson.MaxFileSize+1] = 0
	var dst any
	err := safejson.DecodeStrict(bytes.NewReader(data), &dst)
	assert.ErrorIs(t, err, safejson.ErrTooLarge)
}

func TestDecodeStrict_DepthExceeded(t *testing.T) {
	t.Parallel()
	deep := buildNestedJSON(safejson.MaxDepth + 1)
	var dst any
	err := safejson.DecodeStrict(strings.NewReader(deep), &dst)
	assert.ErrorIs(t, err, safejson.ErrDepthExceeded)
}

func ExampleDecode() {
	type Config struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
	input := `{"host":"localhost","port":8080}`
	var cfg Config
	if err := safejson.Decode(strings.NewReader(input), &cfg); err != nil {
		panic(err)
	}
	// cfg.Host == "localhost", cfg.Port == 8080
}
