// SPDX-License-Identifier: Apache-2.0

package term_test

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	"github.com/nathanbrophy/glacier/term"
)

func TestDownloadProgressRead(t *testing.T) {
	t.Parallel()
	data := []byte("hello world this is a download")
	r := bytes.NewReader(data)
	dp := term.NewDownloadProgress(r, int64(len(data)), "test download")
	assert.NotNil(t, dp.Source, "DownloadProgress.Source must not be nil")

	// Read 10 bytes.
	buf := make([]byte, 10)
	n, err := dp.Read(buf)
	assert.True(t, err == nil || err == io.EOF, "Read error: "+fmt.Sprintf("%v", err))
	assert.Equal(t, n, 10)

	lines, done := dp.Render()
	assert.False(t, done, "DownloadProgress not done after partial read")
	assert.True(t, len(lines) > 0, "Render() returned empty lines")
}

func TestDownloadProgressReadToEOF(t *testing.T) {
	t.Parallel()
	data := []byte("abc")
	r := bytes.NewReader(data)
	dp := term.NewDownloadProgress(r, int64(len(data)), "")

	// Read everything.
	_, err := io.ReadAll(dp)
	require.NoError(t, err, "ReadAll")

	_, done := dp.Render()
	assert.True(t, done, "DownloadProgress.Render() done=false after full read to EOF")
}

func TestDownloadProgressIndeterminate(t *testing.T) {
	t.Parallel()
	data := []byte("xyz")
	r := bytes.NewReader(data)
	dp := term.NewDownloadProgress(r, -1, "unknown size")
	lines, done := dp.Render()
	assert.False(t, done, "indeterminate DownloadProgress done=true before EOF")
	assert.True(t, len(lines) > 0, "Render() returned no lines for indeterminate progress")
}

func TestDownloadProgressReadAfterDone(t *testing.T) {
	t.Parallel()
	// L-add-9: Read after Done() still passes through to Source.
	data := []byte("data")
	r := bytes.NewReader(data)
	dp := term.NewDownloadProgress(r, 100, "label")
	dp.Done()

	buf := make([]byte, 4)
	n, err := dp.Read(buf)
	// Source still has data; Read should work transparently.
	assert.True(t, err == nil || err == io.EOF, "Read after Done(): error "+fmt.Sprintf("%v", err))
	assert.True(t, n > 0 || err == io.EOF, "Read after Done() returned 0 bytes without EOF")
}

func TestNewDownloadProgressOptions(t *testing.T) {
	t.Parallel()
	r := bytes.NewReader([]byte("x"))
	dp := term.NewDownloadProgress(r, 100, "label",
		term.WithProgressShowSpeed(),
		term.WithProgressShowBytes(),
	)
	lines, _ := dp.Render()
	assert.True(t, len(lines) > 0, "Render() returned no lines")
}
