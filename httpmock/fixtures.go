// SPDX-License-Identifier: Apache-2.0

package httpmock

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/internal/safefile"
)

const fixtureMaxBytes int64 = 16 * 1024 * 1024 // 16 MiB

// ErrFixtureTooLarge is returned when a fixture file exceeds fixtureMaxBytes.
var ErrFixtureTooLarge = errors.New("httpmock: fixture too large (max 16 MiB)")

var errDepthExceeded = errors.New("httpmock: fixture nesting depth exceeded (max 32)")

type fixtureDoc struct {
	Stubs []fixtureStub `json:"stubs"`
}

type fixtureStub struct {
	Method  string         `json:"method"`
	Path    string         `json:"path"`
	Respond fixtureRespond `json:"respond"`
}

type fixtureRespond struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    json.RawMessage   `json:"body"`
}

// LoadFixtures reads testdata/httpmock/<name>.json and registers stubs on t.
// Path traversal, file size (16 MiB), depth (32), UTF-8, and unknown fields
// are all validated. On any error, tb.Errorf is called and no stubs register.
func (t *Transport) LoadFixtures(tb assert.TB, name string) error {
	tb.Helper()
	cleanName, err := safefile.Clean(name + ".json")
	if err != nil {
		tb.Errorf("httpmock: LoadFixtures: invalid name %q: %s", name, err)
		return err
	}
	path := "testdata/httpmock/" + cleanName
	f, err := os.Open(path) //nolint:gosec
	if err != nil {
		tb.Errorf("httpmock: LoadFixtures: open %q: %s", path, err)
		return err
	}
	defer f.Close()

	limited := io.LimitReader(f, fixtureMaxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		tb.Errorf("httpmock: LoadFixtures: read %q: %s", path, err)
		return err
	}
	if int64(len(data)) > fixtureMaxBytes {
		tb.Errorf("httpmock: LoadFixtures: %q: %s", path, ErrFixtureTooLarge)
		return ErrFixtureTooLarge
	}
	if err := checkDepth(data); err != nil {
		tb.Errorf("httpmock: LoadFixtures: %q: %s", path, err)
		return err
	}
	if !utf8.Valid(data) {
		err := errors.New("httpmock: fixture contains invalid UTF-8")
		tb.Errorf("httpmock: LoadFixtures: %q: %s", path, err)
		return err
	}
	var doc fixtureDoc
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&doc); err != nil {
		tb.Errorf("httpmock: LoadFixtures: decode %q: %s", path, err)
		return err
	}
	for _, fs := range doc.Stubs {
		stub := t.OnRequest()
		if fs.Method != "" {
			stub.Method(fs.Method)
		}
		if fs.Path != "" {
			stub.Path(fs.Path)
		}
		status := fs.Respond.Status
		if status == 0 {
			status = http.StatusOK
		}
		var bodyBytes []byte
		if len(fs.Respond.Body) > 0 && string(fs.Respond.Body) != "null" {
			bodyBytes = []byte(fs.Respond.Body)
		}
		contentType := ""
		for k, v := range fs.Respond.Headers {
			if http.CanonicalHeaderKey(k) == "Content-Type" {
				contentType = v
				break
			}
		}
		stub.Respond(Body(status, bodyBytes, contentType))
	}
	return nil
}

func checkDepth(data []byte) error {
	dec := json.NewDecoder(strings.NewReader(string(data)))
	depth := 0
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch tok {
		case json.Delim('{'), json.Delim('['):
			depth++
			if depth > 32 {
				return errDepthExceeded
			}
		case json.Delim('}'), json.Delim(']'):
			depth--
		}
	}
	return nil
}
