// SPDX-License-Identifier: Apache-2.0

// Package require mirrors every assertion in the parent assert package with
// t.Fatal semantics so a failing assertion immediately halts the test. Import
// assert for "continue on failure" and require for "halt on failure," matching
// the testify split familiar to most Go contributors. Every signature in this
// package is identical to its counterpart in assert except that failure calls
// t.FailNow instead of t.Errorf.
package require
