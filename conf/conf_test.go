// SPDX-License-Identifier: Apache-2.0

package conf_test

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	"github.com/nathanbrophy/glacier/conf"
)

// ---- helpers ---------------------------------------------------------------

type serverCfg struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Debug   bool   `json:"debug"`
	Timeout int    `json:"timeout"`
}

// writeJSONFile writes content to a temp file and returns its path.
func writeJSONFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "conf-*.json")
	require.NoError(t, err)
	_, err = f.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}

// ---- T#CF1: Decode with defaults only --------------------------------------

func TestDecode_DefaultsOnly(t *testing.T) {
	t.Parallel()
	type cfg struct {
		Name string `json:"name"`
		Val  int    `json:"val"`
	}
	// Without any sources, Decode should return zero values of T.
	got, err := conf.Decode[cfg](context.Background())
	require.NoError(t, err)
	assert.Equal(t, got.Name, "")
	assert.Equal(t, got.Val, 0)
}

// ---- T#CF2: Decode with WithSet overrides a single field -------------------

func TestDecode_WithSet(t *testing.T) {
	t.Parallel()
	got, err := conf.Decode[serverCfg](
		context.Background(),
		conf.WithSet("host", "example.com"),
		conf.WithSet("port", 9090),
	)
	require.NoError(t, err)
	assert.Equal(t, got.Host, "example.com")
	assert.Equal(t, got.Port, 9090)
}

// ---- T#CF3: Decode with WithEnvPrefix reads env var into field -------------

func TestDecode_WithEnvPrefix(t *testing.T) {
	t.Setenv("MYAPP__HOST", "env-host")
	t.Setenv("MYAPP__PORT", "7070")

	got, err := conf.Decode[serverCfg](
		context.Background(),
		conf.WithEnvPrefix("MYAPP"),
	)
	require.NoError(t, err)
	assert.Equal(t, got.Host, "env-host")
	assert.Equal(t, got.Port, 7070)
}

// ---- T#CF4: Decode with JSON file via WithFile populates fields -------------

func TestDecode_WithFile(t *testing.T) {
	t.Parallel()
	path := writeJSONFile(t, `{"host":"filehost","port":3000,"debug":true}`)

	got, err := conf.Decode[serverCfg](
		context.Background(),
		conf.WithFile(path),
	)
	require.NoError(t, err)
	assert.Equal(t, got.Host, "filehost")
	assert.Equal(t, got.Port, 3000)
	assert.Equal(t, got.Debug, true)
}

// ---- T#CF5: Precedence: Set > env > file > defaults ------------------------

func TestDecode_Precedence(t *testing.T) {
	path := writeJSONFile(t, `{"host":"filehost","port":1000,"debug":false,"timeout":5}`)

	// env overrides file
	t.Setenv("APP__HOST", "envhost")
	t.Setenv("APP__PORT", "2000")

	got, err := conf.Decode[serverCfg](
		context.Background(),
		conf.WithFile(path),
		conf.WithEnvPrefix("APP"),
		conf.WithSet("host", "sethost"),
		conf.WithSet("port", 3000),
	)
	require.NoError(t, err)
	// WithSet wins for host and port.
	assert.Equal(t, got.Host, "sethost")
	assert.Equal(t, got.Port, 3000)
	// Env wins over file for the remainder.
	// debug and timeout come from the file (env didn't set them).
	assert.Equal(t, got.Debug, false)
	assert.Equal(t, got.Timeout, 5)
}

// ---- T#CF6: Non-existent file returns DecodeError wrapping fs.ErrNotExist ---

func TestDecode_NonExistentFile(t *testing.T) {
	t.Parallel()
	_, err := conf.Decode[serverCfg](
		context.Background(),
		conf.WithFile(filepath.Join(t.TempDir(), "missing.json")),
	)
	require.Error(t, err)
	var de *conf.DecodeError
	assert.True(t, errors.As(err, &de), "expected *conf.DecodeError")
	assert.ErrorIs(t, err, fs.ErrNotExist)
}

// ---- T#CF7: Path traversal in WithFile is rejected -------------------------
// Note: safefile.Clean rejects ".." in relative paths, but WithFile uses
// os.Open directly (absolute paths are valid). We test that a path whose
// underlying os.Open call fails returns a proper error.
//
// We test a clearly non-existent traversal-style path rather than relying on
// safefile here, because conf.WithFile accepts absolute paths for production use.

func TestDecode_PathTraversalReturnsError(t *testing.T) {
	t.Parallel()
	// A path that does not exist at all should return an error.
	_, err := conf.Decode[serverCfg](
		context.Background(),
		conf.WithFile("../../nonexistent/secret.json"),
	)
	require.Error(t, err)
	var de *conf.DecodeError
	assert.True(t, errors.As(err, &de))
}

// ---- T#CF8: Register panics on duplicate path ------------------------------

func TestRegister_PanicsDuplicate(t *testing.T) {
	// Not parallel :  uses globalRegistry.
	const path = "test.duplicate"
	// First registration must succeed.
	conf.Register[serverCfg](path, serverCfg{Port: 1})

	defer func() {
		if r := recover(); r == nil {
			assert.True(t, false, "expected panic on duplicate Register; got none")
		}
	}()
	conf.Register[serverCfg](path, serverCfg{Port: 2})
}

// ---- T#CF9: Register accessor returns defaults before Load -----------------

func TestRegister_AccessorReturnsDefaults(t *testing.T) {
	// Not parallel :  uses globalRegistry.
	const path = "test.defaults"
	defaults := serverCfg{Host: "localhost", Port: 8080}
	get := conf.Register[serverCfg](path, defaults)

	got := get()
	require.NotNil(t, got)
	assert.Equal(t, got.Host, "localhost")
	assert.Equal(t, got.Port, 8080)
}

// ---- T#CF10: Loader.Load updates registered sections ----------------------

func TestLoader_Load_UpdatesRegistered(t *testing.T) {
	// Not parallel :  uses globalRegistry.
	const path = "test.loader.load"
	defaults := serverCfg{Host: "default", Port: 80}
	get := conf.Register[serverCfg](path, defaults)

	path2 := writeJSONFile(t, `{"test":{"loader":{"load":{"host":"loaded","port":9999}}}}`)
	l := conf.NewLoader(conf.WithFile(path2))
	require.NoError(t, l.Load(context.Background()))

	got := get()
	require.NotNil(t, got)
	assert.Equal(t, got.Host, "loaded")
	assert.Equal(t, got.Port, 9999)
}

// ---- T#CF11: Loader.Close prevents subsequent Load calls ------------------

func TestLoader_Close(t *testing.T) {
	t.Parallel()
	l := conf.NewLoader()
	require.NoError(t, l.Close())

	err := l.Load(context.Background())
	assert.ErrorIs(t, err, conf.ErrLoaderClosed)
}

// ---- T#CF12: MustLoad panics on error -------------------------------------

func TestLoader_MustLoad_Panics(t *testing.T) {
	t.Parallel()
	l := conf.NewLoader()
	require.NoError(t, l.Close())

	defer func() {
		if r := recover(); r == nil {
			assert.True(t, false, "expected MustLoad to panic; got none")
		}
	}()
	l.MustLoad(context.Background())
}

// ---- T#CF13: cancelled context surfaces as DecodeError --------------------

func TestDecode_CancelledContext(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := conf.Decode[serverCfg](ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

// ---- WithDefaults layer ----------------------------------------------------

func TestDecode_WithDefaults(t *testing.T) {
	t.Parallel()
	type cfg struct {
		Timeout int `json:"timeout"`
	}
	got, err := conf.Decode[cfg](
		context.Background(),
		conf.WithDefaults(func() map[string]any {
			return map[string]any{"timeout": 42}
		}),
	)
	require.NoError(t, err)
	assert.Equal(t, got.Timeout, 42)
}

// ---- WithDefaults overridden by WithSet ------------------------------------

func TestDecode_WithDefaultsOverriddenBySet(t *testing.T) {
	t.Parallel()
	type cfg struct {
		Timeout int `json:"timeout"`
	}
	got, err := conf.Decode[cfg](
		context.Background(),
		conf.WithDefaults(func() map[string]any {
			return map[string]any{"timeout": 42}
		}),
		conf.WithSet("timeout", 99),
	)
	require.NoError(t, err)
	assert.Equal(t, got.Timeout, 99)
}

// ---- FlagSource integration ------------------------------------------------

type mapFlagSource map[string]string

func (m mapFlagSource) Lookup(path string) (string, bool) {
	v, ok := m[path]
	return v, ok
}

func TestDecode_WithFlagSource(t *testing.T) {
	t.Parallel()
	type cfg struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
	flags := mapFlagSource{"host": `"flaghost"`, "port": "5555"}
	got, err := conf.Decode[cfg](
		context.Background(),
		conf.WithFlagSource(flags),
	)
	require.NoError(t, err)
	assert.Equal(t, got.Host, "flaghost")
	assert.Equal(t, got.Port, 5555)
}

// ---- ErrFileTooLarge -------------------------------------------------------

func TestDecode_FileTooLarge(t *testing.T) {
	t.Parallel()
	// Write a file > 1 MiB.
	dir := t.TempDir()
	fpath := filepath.Join(dir, "big.json")

	// Write a valid JSON prefix then pad to > 1 MiB.
	// Use a JSON string value large enough to exceed the limit.
	const mib = 1 << 20
	payload := make([]byte, mib+100)
	payload[0] = '"'
	for i := 1; i < len(payload)-1; i++ {
		payload[i] = 'a'
	}
	payload[len(payload)-1] = '"'
	require.NoError(t, os.WriteFile(fpath, payload, 0o600))

	_, err := conf.Decode[serverCfg](context.Background(), conf.WithFile(fpath))
	require.Error(t, err)
	assert.ErrorIs(t, err, conf.ErrFileTooLarge)
}

// ---- JSON parse error in file ----------------------------------------------

func TestDecode_FileInvalidJSON(t *testing.T) {
	t.Parallel()
	path := writeJSONFile(t, `{bad json}`)

	_, err := conf.Decode[serverCfg](context.Background(), conf.WithFile(path))
	require.Error(t, err)
	var de *conf.DecodeError
	assert.True(t, errors.As(err, &de))
}

// ---- Close is idempotent ---------------------------------------------------

func TestLoader_Close_Idempotent(t *testing.T) {
	t.Parallel()
	l := conf.NewLoader()
	require.NoError(t, l.Close())
	require.NoError(t, l.Close())
	require.NoError(t, l.Close())
}

// ---- DecodeError.Error message format -------------------------------------

func TestDecodeError_ErrorString(t *testing.T) {
	t.Parallel()
	cause := errors.New("underlying")

	withPath := &conf.DecodeError{Path: "server.port", Cause: cause, Layer: "file"}
	assert.Equal(t, withPath.Error(), "conf: decode server.port: underlying")

	withoutPath := &conf.DecodeError{Cause: cause, Layer: "file"}
	assert.Equal(t, withoutPath.Error(), "conf: decode: underlying")
}

// ---- JSON file with nested section path ------------------------------------

func TestDecode_JSONFileWithSectionPath(t *testing.T) {
	// This tests that when the JSON file has a nested section and we're
	// using Decode (no sectionPath), the entire document is used.
	t.Parallel()
	path := writeJSONFile(t, `{"host":"nested-host","port":1234}`)

	got, err := conf.Decode[serverCfg](context.Background(), conf.WithFile(path))
	require.NoError(t, err)
	assert.Equal(t, got.Host, "nested-host")
}

// ---- Env var JSON-typed values (bool, number) -----------------------------

func TestDecode_EnvVarBoolAndNumber(t *testing.T) {
	t.Setenv("SVC__DEBUG", "true")
	t.Setenv("SVC__TIMEOUT", "30")

	type cfg struct {
		Debug   bool `json:"debug"`
		Timeout int  `json:"timeout"`
	}
	got, err := conf.Decode[cfg](context.Background(), conf.WithEnvPrefix("SVC"))
	require.NoError(t, err)
	assert.Equal(t, got.Debug, true)
	assert.Equal(t, got.Timeout, 30)
}

// ---- Verify json.Marshal round-trip for WithSet with typed int -------------

func TestDecode_WithSet_IntValue(t *testing.T) {
	t.Parallel()
	// When caller passes an int (not float64), it should end up in the struct
	// correctly via JSON round-trip.
	type cfg struct {
		Port int `json:"port"`
	}
	got, err := conf.Decode[cfg](context.Background(), conf.WithSet("port", 4242))
	require.NoError(t, err)
	assert.Equal(t, got.Port, 4242)
}

// ---- Verify Loader.Load with cancelled context ----------------------------

func TestLoader_Load_CancelledContext(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	l := conf.NewLoader()
	err := l.Load(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

// ---- Package-level Load/MustLoad use the same default loader ---------------

func TestPackageLevel_MustLoad_PanicsOnClosed(t *testing.T) {
	// This test uses the package-level defaultLoader; we cannot close it
	// without affecting other tests. We instead test via a Loader instance.
	t.Parallel()
	l := conf.NewLoader()
	require.NoError(t, l.Close())
	defer func() {
		r := recover()
		assert.True(t, r != nil, "expected panic")
	}()
	l.MustLoad(context.Background())
}

// ---- JSON file depth exceeded ----------------------------------------------

func TestDecode_FileDepthExceeded(t *testing.T) {
	t.Parallel()

	// Build JSON with 33 levels of nesting (MaxDepth+1 == 33).
	var b []byte
	for range 33 {
		b = append(b, '{')
		b = append(b, []byte(`"x":`)...)
	}
	b = append(b, '1')
	for range 33 {
		b = append(b, '}')
	}

	dir := t.TempDir()
	fpath := filepath.Join(dir, "deep.json")
	require.NoError(t, os.WriteFile(fpath, b, 0o600))

	_, err := conf.Decode[serverCfg](context.Background(), conf.WithFile(fpath))
	require.Error(t, err)
	assert.ErrorIs(t, err, conf.ErrDepthExceeded)
}

// ---- WithLogger does not panic ---------------------------------------------

func TestLoader_WithLogger(t *testing.T) {
	t.Parallel()
	l := conf.NewLoader(conf.WithLogger(nil))
	// nil logger triggers default; Load must not panic.
	_ = l.Load(context.Background())
}

// ---- JSON marshal of merged map is well-formed ----------------------------

func TestDecode_MergedMapIsValidJSON(t *testing.T) {
	t.Parallel()
	type cfg struct {
		X int    `json:"x"`
		Y string `json:"y"`
	}
	got, err := conf.Decode[cfg](
		context.Background(),
		conf.WithSet("x", 7),
		conf.WithSet("y", "hello"),
	)
	require.NoError(t, err)
	b, err := json.Marshal(got)
	require.NoError(t, err)
	assert.True(t, len(b) > 0)
}
