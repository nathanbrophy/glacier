// SPDX-License-Identifier: Apache-2.0

package fluent_test

import (
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/fluent"
)

func TestFrom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  []int
		want []int
	}{
		{"nil slice yields nothing", nil, []int{}},
		{"empty slice yields nothing", []int{}, []int{}},
		{"single element", []int{42}, []int{42}},
		{"multiple elements", []int{1, 2, 3, 4, 5}, []int{1, 2, 3, 4, 5}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := fluent.ToSlice(fluent.From(tc.src))
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestFromMap(t *testing.T) {
	t.Parallel()

	t.Run("nil map yields nothing", func(t *testing.T) {
		t.Parallel()
		got := fluent.ToMap(fluent.FromMap[string, int](nil))
		assert.Equal(t, len(got), 0)
	})

	t.Run("populated map yields all pairs", func(t *testing.T) {
		t.Parallel()
		m := map[string]int{"a": 1, "b": 2}
		got := fluent.ToMap(fluent.FromMap(m))
		assert.Equal(t, got, m)
	})
}

func TestFromChan(t *testing.T) {
	t.Parallel()

	t.Run("closed empty channel yields nothing", func(t *testing.T) {
		t.Parallel()
		ch := make(chan int)
		close(ch)
		got := fluent.ToSlice(fluent.FromChan(ch))
		assert.Equal(t, got, []int{})
	})

	t.Run("channel with values", func(t *testing.T) {
		t.Parallel()
		ch := make(chan int, 3)
		ch <- 10
		ch <- 20
		ch <- 30
		close(ch)
		got := fluent.ToSlice(fluent.FromChan(ch))
		assert.Equal(t, got, []int{10, 20, 30})
	})
}

func TestRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		start, stop, step int
		want              []int
	}{
		{"ascending", 0, 5, 1, []int{0, 1, 2, 3, 4}},
		{"ascending step 2", 0, 6, 2, []int{0, 2, 4}},
		{"descending", 5, 0, -1, []int{5, 4, 3, 2, 1}},
		{"empty: start==stop", 3, 3, 1, []int{}},
		{"empty: step mismatch direction", 0, 5, -1, []int{}},
		{"single", 7, 8, 1, []int{7}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := fluent.ToSlice(fluent.Range(tc.start, tc.stop, tc.step))
			assert.Equal(t, got, tc.want)
		})
	}

	t.Run("panics on step zero", func(t *testing.T) {
		t.Parallel()
		defer func() {
			r := recover()
			assert.True(t, r != nil, "expected panic")
		}()
		fluent.Range(0, 5, 0)
	})
}

func TestRepeat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		val   string
		n     int
		count int
	}{
		{"zero times", "x", 0, 0},
		{"negative n yields nothing", "x", -1, 0},
		{"three times", "hi", 3, 3},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := fluent.ToSlice(fluent.Repeat(tc.val, tc.n))
			assert.Equal(t, len(got), tc.count)
			for _, v := range got {
				assert.Equal(t, v, tc.val)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	t.Parallel()

	t.Run("first five squares", func(t *testing.T) {
		t.Parallel()
		i := 0
		seq := fluent.Take(fluent.Generate(func() (int, bool) {
			v := i * i
			i++
			return v, true
		}), 5)
		got := fluent.ToSlice(seq)
		assert.Equal(t, got, []int{0, 1, 4, 9, 16})
	})

	t.Run("terminates when fn returns false", func(t *testing.T) {
		t.Parallel()
		count := 0
		seq := fluent.Generate(func() (int, bool) {
			if count >= 3 {
				return 0, false
			}
			count++
			return count, true
		})
		got := fluent.ToSlice(seq)
		assert.Equal(t, got, []int{1, 2, 3})
	})
}

func TestLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"unix newlines", "hello\nworld\n", []string{"hello", "world"}},
		{"crlf stripped", "foo\r\nbar\r\n", []string{"foo", "bar"}},
		{"no trailing newline", "a\nb", []string{"a", "b"}},
		{"empty reader", "", []string{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := fluent.ToSlice(fluent.Lines(strings.NewReader(tc.input)))
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestWords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"space separated", "foo bar baz", []string{"foo", "bar", "baz"}},
		{"tabs and newlines", "one\ttwo\nthree", []string{"one", "two", "three"}},
		{"empty", "", []string{}},
		{"leading/trailing whitespace", "  hello  ", []string{"hello"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := fluent.ToSlice(fluent.Words(strings.NewReader(tc.input)))
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestSplit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    string
		sep  string
		want []string
	}{
		{"basic csv", "a,b,c", ",", []string{"a", "b", "c"}},
		{"no match", "abc", ",", []string{"abc"}},
		{"empty string", "", ",", []string{""}},
		{"multi-char sep", "a::b::c", "::", []string{"a", "b", "c"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := fluent.ToSlice(fluent.Split(tc.s, tc.sep))
			assert.Equal(t, got, tc.want)
		})
	}

	t.Run("panics on empty sep", func(t *testing.T) {
		t.Parallel()
		defer func() {
			r := recover()
			assert.True(t, r != nil, "expected panic for empty sep")
		}()
		fluent.Split("hello", "")
	})
}

func TestPairs(t *testing.T) {
	t.Parallel()

	t.Run("round-trip via Entries", func(t *testing.T) {
		t.Parallel()
		kvs := []fluent.KV[string, int]{
			{K: "a", V: 1},
			{K: "b", V: 2},
		}
		seq := fluent.From(kvs)
		got := fluent.ToSlice(fluent.Entries(fluent.Pairs(seq)))
		assert.Equal(t, got, kvs)
	})
}
