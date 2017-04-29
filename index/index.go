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
	"github.com/tomachalek/gloomy/index/column"
	"github.com/tomachalek/gloomy/wdict"
	"os"
	"sort"
	"strings"
)

const (
	// MaxNgramSize specifies the largest n-gram
	// (1-gram, 2-gram,..., n-gram) size Gloomy supports
	MaxNgramSize = 10
)

type ngramResultItem struct {
	next     *ngramResultItem
	ngram    []int
	count    int
	metadata []string
}

type NgramResultValue struct {
	Ngram    []int
	Count    int
	Metadata []string
}

// NgramSearchResult is a low level result
// representation where n-grams are just
// arrays of integers (i.e. no translation
// to actual words yet).
//
// The result is implemented to behave
// as a kind of iterator (using HasNext(), Next()
// methods) rather than copying all the result
// data into an array.
type NgramSearchResult struct {
	first *ngramResultItem
	curr  *ngramResultItem
	size  int
}

// GetSize returns a size of the result
// (this is an O(1) operation)
func (nsr *NgramSearchResult) GetSize() int {
	return nsr.size
}

// ResetCursor moves a pointer pointing
// to the current result item back to the
// first result item.
func (nsr *NgramSearchResult) ResetCursor() {
	nsr.curr = nsr.first
}

// HasNext tests whether the result
// has at least one item left (i.e. whether
// it is possible to call Next() and get
// a valid row)
func (nsr *NgramSearchResult) HasNext() bool {
	return nsr.curr != nil
}

// Next returs a following result item.
func (nsr *NgramSearchResult) Next() *NgramResultValue {
	ans := nsr.curr
	if ans == nil {
		return nil
	}
	nsr.curr = nsr.curr.next
	return &NgramResultValue{
		Ngram:    ans.ngram,
		Count:    ans.count,
		Metadata: ans.metadata,
	}
}

func (nsr *NgramSearchResult) addValue(ngram []int, count int, metadata []string) {
	item := &ngramResultItem{ngram: ngram, count: count, metadata: metadata}
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

// GetNgramsAt returns all the ngrams where the first word index equals position
func (n *NgramIndex) GetNgramsAt(position int) *NgramSearchResult {
	n.loadData(position, position) // TODO multiple search phrases not supported yet
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
type SearchableIndex struct {
	index  *NgramIndex
	wstore *wdict.WordDictReader
}

// GetNgramsOf returns all the n-grams with first word
// equal to the 'word' argument
func (si *SearchableIndex) GetNgramsOf(word string) *NgramSearchResult {
	var ans *NgramSearchResult
	w := si.wstore.Find(word)
	col0Idx := sort.Search(si.index.values[0].Size(), func(i int) bool {
		return si.index.values[0].Get(i).Index >= w
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

// ----------------------------------------------------------------------------

// DynamicNgramIndex allows adding items to the index
type DynamicNgramIndex struct {
	index          *NgramIndex
	cursors        []int
	initialLength  int
	metadataWriter *column.MetadataWriter
}

// NewDynamicNgramIndex creates a new instance of DynamicNgramIndex
func NewDynamicNgramIndex(ngramSize int, initialLength int, attrMap map[string]string) *DynamicNgramIndex {
	cursors := make([]int, ngramSize)
	for i := range cursors {
		cursors[i] = -1
	}

	return &DynamicNgramIndex{
		initialLength:  initialLength,
		index:          NewNgramIndex(ngramSize, initialLength, attrMap),
		cursors:        cursors,
		metadataWriter: column.NewMetadataWriter(attrMap),
	}
}

// GetIndex returns internal index structure
func (nib *DynamicNgramIndex) GetIndex() *NgramIndex {
	return nib.index
}

// GetInfo returns a brief human-readable
// information about the index
func (nib *DynamicNgramIndex) GetInfo() string {
	return nib.index.GetInfo()
}

func (nib *DynamicNgramIndex) MetadataWriter() *column.MetadataWriter {
	return nib.metadataWriter
}

// GetNgramsAt returns all the ngrams where the first word index equals position
func (nib *DynamicNgramIndex) GetNgramsAt(position int) *NgramSearchResult {
	return nib.index.GetNgramsAt(position)
}

// AddNgram adds a new n-gram represented as an array
// of indices to the index
func (nib *DynamicNgramIndex) AddNgram(ngram []int, count int, metadata []column.AttrVal) {
	sp := nib.findSplitPosition(ngram)
	for i := 0; i < len(nib.index.values); i++ {
		col := nib.index.values[i]
		if nib.cursors[i] >= col.Size()-1 {
			col.Extend(nib.initialLength / 2)
		}

		if i == sp-1 {
			col.Get(nib.cursors[i]).UpTo++

		} else if i > sp-1 {
			nib.cursors[i]++
			upTo := 0
			if i < len(nib.cursors)-1 {
				upTo = nib.cursors[i+1] + 1
			}
			col.Set(nib.cursors[i], &column.IndexItem{Index: ngram[i], UpTo: upTo})
		}
	}
	lastPos := nib.cursors[len(nib.index.values)-1]
	if lastPos >= nib.index.counts.Size()-1 {
		nib.index.counts.Extend(nib.initialLength / 2)
	}
	nib.index.counts.Set(lastPos, column.AttrVal(count))
	if lastPos >= nib.metadataWriter.Size()-1 {
		nib.metadataWriter.Extend(nib.initialLength / 2)
	}
	nib.metadataWriter.Set(lastPos, metadata)
	// TODO add metadata
}

func (nib *DynamicNgramIndex) findSplitPosition(ngram []int) int {
	for i := 0; i < len(ngram); i++ {
		if nib.cursors[i] == -1 || ngram[i] != nib.index.values[i].Get(nib.cursors[i]).Index {
			return i
		}
	}
	return -1
}

// Finish should be called once adding of n-grams
// is done. The method frees up some memory preallocated
// for new n-grams.
func (nib *DynamicNgramIndex) Finish() {
	for i, v := range nib.index.values {
		v.Shrink(nib.cursors[i])
	}
	nib.metadataWriter.Shrink(nib.cursors[len(nib.index.values)-1])
}

// Save stores current index data to bunch of files
// within the provided directory.
func (nib *DynamicNgramIndex) Save(dirPath string) error {
	var err error
	for i, col := range nib.index.values {
		if err = col.Save(i, dirPath); err != nil {
			return err
		}
	}
	nib.index.counts.Save(dirPath)
	nib.metadataWriter.Save(dirPath)
	return err
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
