// Copyright 2017 Tomas Machalek <tomas.machalek@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package index

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func createSimpleResult() *NgramSearchResult {
	r := &NgramSearchResult{}
	item1 := &ngramResultItem{}
	item1.ngram = []int{0}
	r.first = item1
	r.curr = item1
	item2 := &ngramResultItem{}
	item2.ngram = []int{1}
	item1.next = item2
	item3 := &ngramResultItem{}
	item3.ngram = []int{2}
	item2.next = item3
	r.last = item3
	r.size = 3
	return r
}

func createAnotherResult() *NgramSearchResult {
	r := &NgramSearchResult{}
	item1 := &ngramResultItem{}
	item1.ngram = []int{3}
	r.first = item1
	r.curr = item1
	item2 := &ngramResultItem{}
	item2.ngram = []int{4}
	item1.next = item2
	item3 := &ngramResultItem{}
	item3.ngram = []int{5}
	item2.next = item3
	r.last = item3
	r.size = 3
	return r
}

// ------------------------------------------------------------------

func TestNewNgramIndex(t *testing.T) {
	d := NewNgramIndex(3, 5, make(map[string]string))
	assert.Equal(t, 3, len(d.values))
}

func TestNgramSearchResultEmpty(t *testing.T) {
	r := &NgramSearchResult{}
	ans := r.Next()
	assert.Nil(t, ans)
}

func TestNgramSearchResultAddValue(t *testing.T) {
	r := &NgramSearchResult{}
	r.addValue([]int{0}, 1, []string{})
	r.addValue([]int{1}, 1, []string{})

	assert.Equal(t, 2, r.GetSize())
	assert.True(t, r.first.next == r.last)
	assert.True(t, r.first.next == r.curr)
}

func TestNGramSearchResultResetCursor(t *testing.T) {
	r := &NgramSearchResult{}
	r.addValue([]int{0}, 1, []string{})
	r.addValue([]int{1}, 1, []string{})
	r.ResetCursor()

	assert.True(t, r.first == r.curr)
}

func TestNgramSearchResultIter2(t *testing.T) {
	r := createSimpleResult()
	tst := make([]int, 3)

	for i := 0; r.HasNext(); i++ {
		tst[i] = r.Next().Ngram[0]
	}
	assert.Equal(t, []int{0, 1, 2}, tst)
}

func TestNgramSearchResultAppend(t *testing.T) {
	r1 := createSimpleResult()
	r2 := createAnotherResult()
	r1.Append(r2)

	assert.Equal(t, 6, r1.GetSize())
	assert.True(t, r1.first.next.next.next == r2.first)
	assert.True(t, r1.last == r2.last)
	assert.True(t, r1.curr == r1.first)
}

func TestNgramSearchResultSlice(t *testing.T) {
	r := &NgramSearchResult{}
	for i := 0; i < 20; i++ {
		r.addValue([]int{i}, 1, []string{})
	}
	r.Slice(10, 15)

	assert.Equal(t, 5, r.GetSize())
	assert.Equal(t, 10, r.first.ngram[0])
	assert.True(t, r.last == r.first.next.next.next.next)
	assert.Nil(t, r.first.next.next.next.next.next)
	assert.Equal(t, 14, r.first.next.next.next.next.ngram[0])
	assert.True(t, r.first == r.curr)
}

func TestNgramSearchResultSliceZero(t *testing.T) {
	r := &NgramSearchResult{}
	for i := 0; i < 10; i++ {
		r.addValue([]int{i}, 1, []string{})
	}
	ok := r.Slice(5, 5)
	assert.False(t, ok)
	assert.Equal(t, 10, r.GetSize())
}

func TestNgramSearchResultSliceNegativeRight(t *testing.T) {
	r := &NgramSearchResult{}
	for i := 0; i < 10; i++ {
		r.addValue([]int{i}, 1, []string{})
	}
	ok := r.Slice(5, -1)
	assert.False(t, ok)
	assert.Equal(t, 10, r.GetSize())
}

func TestNgramSearchResultSliceNegativeLeft(t *testing.T) {
	r := &NgramSearchResult{}
	for i := 0; i < 5; i++ {
		r.addValue([]int{i}, 1, []string{})
	}
	assert.Panics(t, func() {
		r.Slice(-1, 4)
	})
}

func TestNgramSearchResultSliceTooBigRight(t *testing.T) {
	r := &NgramSearchResult{}
	for i := 0; i < 5; i++ {
		r.addValue([]int{i}, 1, []string{})
	}
	assert.Panics(t, func() {
		r.Slice(-1, 10)
	})
}
