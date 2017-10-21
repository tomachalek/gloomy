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
	"github.com/stretchr/testify/assert"
	"testing"
)

func createSimpleResult() *NgramSearchResult {
	r := &NgramSearchResult{}
	item1 := &NgramResultItem{}
	item1.Ngram = []int{0}
	r.first = item1
	r.curr = item1
	item2 := &NgramResultItem{}
	item2.Ngram = []int{1}
	item1.next = item2
	item3 := &NgramResultItem{}
	item3.Ngram = []int{2}
	item2.next = item3
	item4 := &NgramResultItem{}
	item4.Ngram = []int{3}
	item3.next = item4
	item5 := &NgramResultItem{}
	item5.Ngram = []int{4}
	item4.next = item5
	r.last = item5
	r.size = 5
	return r
}

func createAnotherResult() *NgramSearchResult {
	r := &NgramSearchResult{}
	item1 := &NgramResultItem{}
	item1.Ngram = []int{3}
	r.first = item1
	r.curr = item1
	item2 := &NgramResultItem{}
	item2.Ngram = []int{4}
	item1.next = item2
	item3 := &NgramResultItem{}
	item3.Ngram = []int{5}
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

	assert.Equal(t, 2, r.Size())
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
	tst := make([]int, 5)

	for i := 0; r.HasNext(); i++ {
		tst[i] = r.Next().Ngram[0]
	}
	assert.Equal(t, []int{0, 1, 2, 3, 4}, tst)
}

func TestNgramSearchResultAppend(t *testing.T) {
	r1 := createSimpleResult()
	r2 := createAnotherResult()
	r1.Append(r2)

	assert.Equal(t, 8, r1.Size())
	assert.True(t, r1.first.next.next.next.next.next == r2.first)
	assert.True(t, r1.last == r2.last)
	assert.True(t, r1.curr == r1.first)
}

func TestNgramAppendNonemptyToEmpty(t *testing.T) {
	r1 := NgramSearchResult{}
	r2 := createSimpleResult()
	r1.Append(r2)
	assert.Equal(t, 5, r1.Size())
	assert.True(t, r1.last == r2.last)
	assert.True(t, r1.first == r2.first)
	assert.True(t, r1.curr == r1.first)
}

func TestNgramSearchResultSlice(t *testing.T) {
	r := &NgramSearchResult{}
	for i := 0; i < 20; i++ {
		r.addValue([]int{i}, 1, []string{})
	}
	r.Slice(10, 15)

	assert.Equal(t, 5, r.Size())
	assert.Equal(t, 10, r.first.Ngram[0])
	assert.True(t, r.last == r.first.next.next.next.next)
	assert.Nil(t, r.first.next.next.next.next.next)
	assert.Equal(t, 14, r.first.next.next.next.next.Ngram[0])
	assert.True(t, r.first == r.curr)
}

func TestNgramSearchResultSliceZero(t *testing.T) {
	r := &NgramSearchResult{}
	for i := 0; i < 10; i++ {
		r.addValue([]int{i}, 1, []string{})
	}
	ok := r.Slice(5, 5)
	assert.False(t, ok)
	assert.Equal(t, 10, r.Size())
}

func TestNgramSearchResultSliceNegativeRight(t *testing.T) {
	r := &NgramSearchResult{}
	for i := 0; i < 10; i++ {
		r.addValue([]int{i}, 1, []string{})
	}
	ok := r.Slice(5, -1)
	assert.False(t, ok)
	assert.Equal(t, 10, r.Size())
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

func TestNgramSearchResultRemoveFirst(t *testing.T) {
	r := createSimpleResult()
	r1 := r.Next()
	r2 := r.Next()
	r3 := r.Next()
	r4 := r.Next()
	r5 := r.Next()
	r.ResetCursor()
	rRemoved := r.RemoveNext(nil)
	assert.True(t, r1 == rRemoved)
	assert.Nil(t, rRemoved.next)
	assert.Equal(t, 4, r.Size())
	assert.Equal(t, r2, r.first)
	assert.Equal(t, r2, r.Next())
	assert.Equal(t, r3, r.Next())
	assert.Equal(t, r4, r.Next())
	assert.Equal(t, r5, r.Next())
	assert.Equal(t, r5, r.last)
}

func TestNgramSearchResultRemoveTwoMiddle(t *testing.T) {
	r := createSimpleResult()
	r1 := r.Next()
	r.Next()
	r.Next()
	r4 := r.Next()
	r.RemoveNext(r1)
	r.RemoveNext(r1)
	assert.Equal(t, 3, r.Size())
	assert.True(t, r1.next == r4)
	count := 0
	r.ResetCursor()
	for r.HasNext() {
		r.Next()
		count++
	}
	assert.Equal(t, 3, count)
}

func TestNgramSearchResultRemoveLast(t *testing.T) {
	r := createSimpleResult()
	r.Next()
	r.Next()
	r.Next()
	r4 := r.Next()
	r.RemoveNext(r4)
	assert.Equal(t, 4, r.Size())
	assert.Nil(t, r4.next)
	assert.True(t, r.last == r4)
}

func TestNgramSearchResultFilterRejectAll(t *testing.T) {
	r := createSimpleResult()
	r.Filter(func(v *NgramResultItem) bool {
		return false
	})
	assert.Equal(t, 0, r.Size())
}

func TestNgramSearchResultFilterAcceptAll(t *testing.T) {
	r := createSimpleResult()
	r.Filter(func(v *NgramResultItem) bool {
		return true
	})
	assert.Equal(t, 5, r.Size())
}

func TestNgramSearchResultFilterAcceptEven(t *testing.T) {
	r := createSimpleResult()
	r.Filter(func(v *NgramResultItem) bool {
		return v.Ngram[0]%2 == 0
	})
	assert.Equal(t, 3, r.Size())
	counter := 0
	r.ResetCursor()
	for r.HasNext() {
		v := r.Next()
		assert.Equal(t, counter*2, v.Ngram[0])
		counter++
	}
	assert.Equal(t, 3, counter)
}

func TestNgramSearchResultFilterAcceptBlock(t *testing.T) {
	r := createSimpleResult()
	r.Filter(func(v *NgramResultItem) bool {
		return v.Ngram[0] == 1 || v.Ngram[0] == 2
	})
	assert.Equal(t, 2, r.Size())
	counter := 1
	r.ResetCursor()
	for r.HasNext() {
		v := r.Next()
		assert.Equal(t, counter, v.Ngram[0])
		counter++
	}
	assert.Equal(t, 3, counter)
}

/*
func TestLoadNgramIndex(t *testing.T) {
	tmpdir := os.Getenv("TMP") // typically Windows
	if tmpdir == "" {
		tmpdir = "/tmp" //  we assume Linux/Unix
	}
	fmt.Println("TMPDIR", tmpdir)
	idx := LoadNgramIndex("/foo/bar", []string{})
	assert.NotNil(t, idx)
}
*/
