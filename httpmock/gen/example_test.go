// SPDX-License-Identifier: Apache-2.0

package gen_test

import (
	"fmt"

	"github.com/nathanbrophy/glacier/httpmock/gen"
)

// ExampleGenerate_patternRequired shows that Generate returns an error when no
// pattern is supplied.
func ExampleGenerate_patternRequired() {
	err := gen.Generate(gen.Options{})
	fmt.Println(err != nil)
	// Output: true
}
