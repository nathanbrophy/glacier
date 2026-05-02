// SPDX-License-Identifier: Apache-2.0

package fluent_test

import (
	"bytes"
	"testing"

	"github.com/nathanbrophy/glacier/fluent"
)

func FuzzWords(f *testing.F) {
	f.Add([]byte("hello world"))
	f.Add([]byte(""))
	f.Add([]byte("  \t\n  "))
	f.Add([]byte("one\ttwo\nthree"))
	f.Fuzz(func(t *testing.T, data []byte) {
		seq := fluent.Words(bytes.NewReader(data))
		for range seq {
		}
	})
}
