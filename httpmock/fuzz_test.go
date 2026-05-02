// SPDX-License-Identifier: Apache-2.0

package httpmock_test

import (
	"testing"

	"github.com/nathanbrophy/glacier/httpmock"
)

func FuzzFixture(f *testing.F) {
	// Seed corpus: valid fixture, malformed, oversized path, traversal attempt.
	f.Add("basic")
	f.Add("malformed")
	f.Add("../etc/passwd")
	f.Add("")
	f.Add("a/b/c")

	f.Fuzz(func(t *testing.T, name string) {
		tb := &silentTB{t: t}
		rt := httpmock.New()
		// LoadFixtures must not panic for any input — it must always return
		// a typed error and call tb.Errorf.
		_ = rt.LoadFixtures(tb, name)
	})
}

// silentTB is a TB that swallows output to avoid noisy fuzz logs.
type silentTB struct {
	t *testing.T
}

func (s *silentTB) Helper()                   {}
func (s *silentTB) Errorf(_ string, _ ...any) {}
func (s *silentTB) Fatalf(_ string, _ ...any) { s.t.FailNow() }
func (s *silentTB) FailNow()                  { s.t.FailNow() }
func (s *silentTB) Cleanup(fn func())         { s.t.Cleanup(fn) }
func (s *silentTB) Name() string              { return s.t.Name() }
