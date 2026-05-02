// SPDX-License-Identifier: Apache-2.0

package fluent_test

import (
	"bytes"
	"testing"

	"github.com/nathanbrophy/glacier/fluent"
)

func FuzzLines(f *testing.F) {
	f.Add([]byte("hello\nworld\n"))
	f.Add([]byte(""))
	f.Add([]byte("\r\n\r\n"))
	f.Add([]byte("no newline at end"))
	f.Fuzz(func(t *testing.T, data []byte) {
		seq := fluent.Lines(bytes.NewReader(data))
		for range seq {
		}
	})
}
