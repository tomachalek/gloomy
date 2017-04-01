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
	return r
}

func TestNewNgramIndex(t *testing.T) {
	d := NewNgramIndex(3, 5)
	assert.Equal(t, 3, len(d.values))
}

func TestNgramSearchResultEmpty(t *testing.T) {
	r := &NgramSearchResult{}
	ans := r.Next()
	assert.Nil(t, ans)
}

func TestNgramSearchResultIter2(t *testing.T) {
	r := createSimpleResult()
	tst := make([]int, 3)

	for i := 0; r.HasNext(); i++ {
		tst[i] = r.Next()[0]
	}
	assert.Equal(t, []int{0, 1, 2}, tst)
}
