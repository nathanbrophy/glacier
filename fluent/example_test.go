// SPDX-License-Identifier: Apache-2.0

package fluent_test

import (
	"fmt"

	"github.com/nathanbrophy/glacier/fluent"
)

func ExampleFrom() {
	seq := fluent.From([]int{1, 2, 3})
	for v := range seq {
		fmt.Println(v)
	}
	// Output:
	// 1
	// 2
	// 3
}

func ExampleMap() {
	src := fluent.From([]int{1, 2, 3, 4, 5})
	doubled := fluent.Map(src, func(v int) int { return v * 2 })
	for v := range doubled {
		fmt.Println(v)
	}
	// Output:
	// 2
	// 4
	// 6
	// 8
	// 10
}

func ExampleFilter() {
	src := fluent.From([]int{1, 2, 3, 4, 5, 6})
	evens := fluent.Filter(src, func(v int) bool { return v%2 == 0 })
	for v := range evens {
		fmt.Println(v)
	}
	// Output:
	// 2
	// 4
	// 6
}

func ExampleToSlice() {
	src := fluent.From([]string{"a", "b", "c"})
	s := fluent.ToSlice(src)
	fmt.Println(s)
	// Output:
	// [a b c]
}

func ExampleReduce() {
	src := fluent.From([]int{1, 2, 3, 4, 5})
	sum := fluent.Reduce(src, 0, func(acc, v int) int { return acc + v })
	fmt.Println(sum)
	// Output:
	// 15
}

func ExampleSort() {
	src := fluent.From([]int{5, 3, 1, 4, 2})
	for v := range fluent.Sort(src) {
		fmt.Println(v)
	}
	// Output:
	// 1
	// 2
	// 3
	// 4
	// 5
}

func ExampleDistinct() {
	src := fluent.From([]int{1, 2, 1, 3, 2, 4})
	for v := range fluent.Distinct(src) {
		fmt.Println(v)
	}
	// Output:
	// 1
	// 2
	// 3
	// 4
}
