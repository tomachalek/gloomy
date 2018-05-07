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

// Package index represents an n-gram index as a read-only structure
// providing both low level methods for accessing the internal ngram tree
// and higher level methods for searching a specific word.
package index

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/tomachalek/gloomy/index/column"
	"github.com/tomachalek/gloomy/wdict"
)

const (
	// MaxNgramSize specifies the largest n-gram
	// (1-gram, 2-gram,..., n-gram) size Gloomy supports
	MaxNgramSize = 10
)

// --------------------------------------------------------------------

// NgramIndex is a low-level implementation
// of a n-gram index.
type NgramIndex struct {
	values   []*column.IndexColumn
	counts   column.AttrValColumn
	metadata *column.MetadataReader
}

// GetInfo returns a human readable overview
// of the index
func (n *NgramIndex) GetInfo() string {
	sizes := make([]string, len(n.values))
	for i, v := range n.values {
		sizes[i] = fmt.Sprintf("%d", v.Size())
	}
	return fmt.Sprintf("NgramIndex, num cols: %d, sizes %s", len(n.values), strings.Join(sizes, ", "))
}

// LoadRange loads data for all the configured
// n-gram and metadata columns delimited by
// interval [fromPos, toPos] applied
// on the zero-th n-gram column (e.g. 100-200 on
// 0th column means 1700-3500 on the 1st, 7000-9000
// on 2th column which is calculated automatically).
//
// Both interval ends are included.
func (n *NgramIndex) LoadRange(fromPos int, toPos int) {
	n.loadData(fromPos, toPos)
}

// GetNgramsAt returns all the ngrams where the first word
// index equals position
func (n *NgramIndex) GetNgramsAt(position int) *NgramSearchResult {
	result := &NgramSearchResult{}
	n.getNextTokenRecords(0, position, position, make([]int, 0), result)
	result.ResetCursor()
	return result
}

func (n *NgramIndex) findLoadRange(colIdx int, fromRow int, toRow int) (int, int) {
	leftIdx := fromRow
	if fromRow > 0 {
		leftIdx = n.values[colIdx].Get(fromRow-1).UpTo + 1
	}
	rightIdx := n.values[colIdx].Get(toRow).UpTo
	return leftIdx, rightIdx
}

func (n *NgramIndex) loadData(fromRow int, toRow int) {
	left := fromRow
	right := toRow
	for i := 0; i < len(n.values)-1; i++ {
		left, right = n.findLoadRange(i, left, right)
		n.values[i+1].LoadChunk(left, right)
	}
	n.counts.LoadChunk(left, right)
	n.metadata.LoadChunk(left, right)
}

func (n *NgramIndex) getNextTokenRecords(colIdx int, fromRow int, toRow int, prevTokens []int, result *NgramSearchResult) {
	col := n.values[colIdx]
	for i := fromRow; i <= toRow; i++ {
		idx := col.Get(i)
		currNgram := append(prevTokens, idx.Index)
		if colIdx == len(n.values)-1 {
			result.addValue(currNgram, int(n.counts.Get(i)), n.metadata.Get(i))

		} else {
			nextFromIdx := 0
			if fromRow > 0 {
				nextFromIdx = col.Get(i-1).UpTo + 1
			}
			nextToIdx := idx.UpTo
			n.getNextTokenRecords(colIdx+1, nextFromIdx, nextToIdx, currNgram, result)
		}
	}
}

// NewNgramIndex creates a new empty instance of NgramIndex
func NewNgramIndex(ngramSize int, initialLength int, attrMap map[string]string) *NgramIndex {
	countsCol := column.NewCountsColumn(initialLength)
	ans := &NgramIndex{
		values:   make([]*column.IndexColumn, ngramSize),
		counts:   countsCol,
		metadata: nil,
	}
	for i := range ans.values {
		ans.values[i] = column.NewIndexColumn(initialLength)
	}
	return ans
}

// ----------------------------------------------------------------------------

// SearchableIndex is a higher-level representation
// of ngram-index with some functions allowing searching
//
// Please note that SearchableIndex does not handle data
// loading automatically. It provides method LoadRange
// to load a specified part of column data but the logic
// is up to a search routine (which decides which words
// we are actually looking for by parsing a query),
type SearchableIndex struct {
	index  *NgramIndex
	wstore *wdict.WordDictReader
}

// GetNgramsOf returns all the n-grams with first word
// equal to the 'word' argument
func (si *SearchableIndex) GetNgramsOf(word string) *NgramSearchResult {
	var ans *NgramSearchResult
	w := si.wstore.Find(word)
	if w == -1 {
		return &NgramSearchResult{}
	}
	col0Idx := sort.Search(si.index.values[0].Size(), func(i int) bool {
		return si.index.values[0].Get(i).Index >= w
	})
	if col0Idx == si.index.values[0].Size() {
		return ans
	}
	si.LoadRange(col0Idx, col0Idx)
	ans = si.index.GetNgramsAt(col0Idx)
	return ans
}

// LoadRange loads column data starting from fromIdx
// up to toIdx
func (si *SearchableIndex) LoadRange(fromIdx int, toIdx int) {
	si.index.LoadRange(fromIdx, toIdx)
}

// GetCol0Idx returns an index within zero column
// of provided word identied by an index within
// word dictionary
func (si *SearchableIndex) GetCol0Idx(widx int) int {
	ans := sort.Search(si.index.values[0].Size(), func(i int) bool {
		return si.index.values[0].Get(i).Index >= widx
	})
	if si.index.values[0].Get(ans).Index == widx {
		return ans
	}
	return -1
}

// GetNgramsOfColIdx returns all the n-grams with the first word identified
// by its index within zero column
func (si *SearchableIndex) GetNgramsOfColIdx(idx int) *NgramSearchResult {
	var ans *NgramSearchResult
	if idx >= si.index.values[0].Size() {
		return ans
	}
	ans = si.index.GetNgramsAt(idx)
	return ans
}

// GetNgramsOfWidx returns all the n-grams with the first word identified
// by its word dictionary index value
func (si *SearchableIndex) GetNgramsOfWidx(idx int) *NgramSearchResult {
	var ans *NgramSearchResult
	col0Idx := sort.Search(si.index.values[0].Size(), func(i int) bool {
		return si.index.values[0].Get(i).Index >= idx
	})
	if col0Idx == si.index.values[0].Size() {
		return ans
	}
	ans = si.index.GetNgramsAt(col0Idx)
	return ans
}

// OpenSearchableIndex creates a instance of SearchableIndex
// based on internal NgramIndex instance and WordIndex instance
func OpenSearchableIndex(index *NgramIndex, wstore *wdict.WordDictReader) *SearchableIndex {
	return &SearchableIndex{index: index, wstore: wstore}
}

// ---------------------------------------------------------------------

// LoadNgramIndex loads index data from within
// a specified directory.
func LoadNgramIndex(dirPath string, attrs []string) *NgramIndex {
	colIdxPaths := make([]string, MaxNgramSize)
	ans := &NgramIndex{}
	for i := 0; i < MaxNgramSize; i++ {
		tmp := column.CreateColIdxPath(i, dirPath)
		if _, err := os.Stat(tmp); os.IsNotExist(err) {
			colIdxPaths = colIdxPaths[:i]
			break
		}
		colIdxPaths[i] = tmp
	}
	var err2 error
	ans.counts, err2 = column.LoadCountsColumn(dirPath)
	if err2 != nil {
		panic(err2)
	}
	var err3 error
	ans.metadata, err3 = column.LoadMetadataReader(dirPath, attrs)
	if err3 != nil {
		panic(err3)
	}
	ans.values = make([]*column.IndexColumn, len(colIdxPaths))
	for i := range ans.values {
		ans.values[i] = column.NewBoundIndexColumn(colIdxPaths[i])
		if i == 0 {
			ans.values[i].LoadWholeChunk() // TODO
		}
	}
	return ans
}
