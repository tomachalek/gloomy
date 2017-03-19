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
	"fmt"
	"github.com/tomachalek/gloomy/wstore"
	"strings"
)

type NgramResultItem struct {
	next  *NgramResultItem
	ngram []int
}

type NgramSearchResult struct {
	first *NgramResultItem
	curr  *NgramResultItem
	size  int
}

func (nsr *NgramSearchResult) GetSize() int {
	return nsr.size
}

func (nsr *NgramSearchResult) ResetCursor() {
	nsr.curr = nsr.first
}

func (nsr *NgramSearchResult) HasNext() bool {
	return nsr.curr != nil
}

func (nsr *NgramSearchResult) Next() []int {
	ans := nsr.curr
	nsr.curr = nsr.curr.next
	return ans.ngram
}

func (nsr *NgramSearchResult) addValue(ngram []int) {
	item := &NgramResultItem{ngram: ngram}
	if nsr.first == nil {
		nsr.first = item
	}
	if nsr.curr != nil {
		nsr.curr.next = item
	}
	nsr.curr = item
	nsr.size++
}

// --------------------------------------------------------------------

type IndexItem struct {
	index int
	upTo  int
}

type IndexColumn []*IndexItem

type NgramIndex struct {
	values []IndexColumn
}

func (n *NgramIndex) GetInfo() string {
	sizes := make([]string, len(n.values))
	for i, v := range n.values {
		sizes[i] = fmt.Sprintf("%d", len(v))
	}
	return fmt.Sprintf("NgramIndex, num cols: %d, sizes %s", len(n.values), strings.Join(sizes, ", "))
}

func (n *NgramIndex) GetNgramsAt(position int) *NgramSearchResult {
	result := &NgramSearchResult{}
	n.getNextTokenRecords(position, 0, make([]int, 0), result)
	result.ResetCursor()
	return result
}

func (n *NgramIndex) getNextTokenRecords(position int, tokenIdx int, prevTokens []int, result *NgramSearchResult) {
	col := n.values[tokenIdx]
	from := position
	if tokenIdx > 1 {
		if position > 0 {
			from = col[position-1].upTo + 1

		} else {
			from = 0
		}
	}
	//log.Printf("tokenIdx: %d, token: %v, [%d, %d]", tokenIdx, col[position], from, position)
	for _, idx := range col[from : position+1] {
		currNgram := append(prevTokens, idx.index)
		if tokenIdx == len(n.values)-1 {
			result.addValue(currNgram)

		} else {
			n.getNextTokenRecords(idx.upTo, tokenIdx+1, currNgram, result)
		}
	}
}

func NewNgramIndex(ngramSize int, initialLength int) *NgramIndex {
	ans := &NgramIndex{
		values: make([]IndexColumn, ngramSize),
	}
	for i := range ans.values {
		ans.values[i] = make(IndexColumn, initialLength)
	}
	return ans
}

// ----------------------------------------------------------------------------

type SearchableIndex struct {
	index  *NgramIndex
	wstore *wstore.WordIndex
}

func (si *SearchableIndex) GetNgramsOf(word string) [][]string {
	w := si.wstore.Find(word)
	result := si.index.GetNgramsAt(w)
	ans := make([][]string, result.GetSize())
	for i := 0; result.HasNext(); i++ {
		tmp := result.Next()
		ans[i] = si.wstore.DecodeNgram(tmp)
	}
	return ans
}

func OpenSearchableIndex(index *NgramIndex, wstore *wstore.WordIndex) *SearchableIndex {
	return &SearchableIndex{index: index, wstore: wstore}
}

// ----------------------------------------------------------------------------

// DynamicNgramIndex allows adding items to the index
type DynamicNgramIndex struct {
	index         *NgramIndex
	cursors       []int
	initialLength int
}

func NewDynamicNgramIndex(ngramSize int, initialLength int) *DynamicNgramIndex {
	cursors := make([]int, ngramSize)
	for i := range cursors {
		cursors[i] = -1
	}
	return &DynamicNgramIndex{
		initialLength: initialLength,
		index:         NewNgramIndex(ngramSize, initialLength),
		cursors:       cursors,
	}
}

func (nib *DynamicNgramIndex) GetIndex() *NgramIndex {
	return nib.index
}

func (nib *DynamicNgramIndex) GetInfo() string {
	return nib.index.GetInfo()
}

func (nib *DynamicNgramIndex) GetNgramsAt(position int) *NgramSearchResult {
	return nib.index.GetNgramsAt(position)
}

func (nib *DynamicNgramIndex) AddNgram(ngram []int) {
	for i := len(ngram) - 1; i >= 0; i-- {
		nib.addValue(i, ngram[i])
	}
}

func (nib *DynamicNgramIndex) Finish() {
	for i, v := range nib.index.values {
		nib.index.values[i] = v[:nib.cursors[i]]
	}
}

func (nib *DynamicNgramIndex) addValue(tokenPos int, index int) {
	col := nib.index.values[tokenPos]
	if nib.cursors[tokenPos] >= len(col) {
		nib.index.values[tokenPos] = append(col, make(IndexColumn, nib.initialLength/2)...)
		col = nib.index.values[tokenPos]
	}
	upTo := 0
	if tokenPos < len(nib.cursors)-1 {
		upTo = nib.cursors[tokenPos+1]
	}
	if nib.cursors[tokenPos] == -1 || nib.index.values[tokenPos][nib.cursors[tokenPos]].index != index {
		nib.cursors[tokenPos]++
		col[nib.cursors[tokenPos]] = &IndexItem{index: index, upTo: upTo}
		//log.Printf("adding pos %d, record [%d, %d]", tokenPos, index, upTo)

	} else {
		//log.Println("Not moving cursor on ", index)
	}
}
