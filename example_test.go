// SPDX-License-Identifier: Apache-2.0

package glacier_test

import (
	"fmt"

	"github.com/nathanbrophy/glacier/errs"
)

// Example shows the umbrella package's role: it is a documentation anchor
// for the framework. Users import the leaf packages directly (errs, conf,
// cli, mock, httpmock, term, fixture, ...). The root package itself has no
// callable surface; see specs/0002-framework-shape.md for the layered
// dependency map.
func Example() {
	err := errs.Sentinel("example: synthetic")
	fmt.Println(err.Error())
	// Output: example: synthetic
}
