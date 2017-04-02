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
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/tomachalek/gloomy/wstore"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// MaxNgramSize specifies the largest n-gram
	// (1-gram, 2-gram,..., n-gram) size Gloomy supports
	MaxNgramSize = 10

	// ColumnIndexFilenameMask specifies a filename for
	// an n-gram column (= data for all the variants at
	// a specific position within an n-gram)
	ColumnIndexFilenameMask = "idx_col_%d.idx"
)

type ngramResultItem struct {
	next  *ngramResultItem
	ngram []int
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
func (nsr *NgramSearchResult) Next() []int {
	ans := nsr.curr
	if ans == nil {
		return nil
	}
	nsr.curr = nsr.curr.next
	return ans.ngram
}

func (nsr *NgramSearchResult) addValue(ngram []int) {
	item := &ngramResultItem{ngram: ngram}
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

type indexItem struct {
	index int
	upTo  int
}

type indexColumn []*indexItem

// NgramIndex is a low-level implementation
// of a n-gram index.
type NgramIndex struct {
	values []indexColumn
}

// GetInfo returns a human readable overview
// of the index
func (n *NgramIndex) GetInfo() string {
	sizes := make([]string, len(n.values))
	for i, v := range n.values {
		sizes[i] = fmt.Sprintf("%d", len(v))
	}
	return fmt.Sprintf("NgramIndex, num cols: %d, sizes %s", len(n.values), strings.Join(sizes, ", "))
}

// GetNgramsAt returns all the ngrams where the first word index equals position
func (n *NgramIndex) GetNgramsAt(position int) *NgramSearchResult {
	result := &NgramSearchResult{}
	n.getNextTokenRecords(0, position, position, make([]int, 0), result)
	result.ResetCursor()
	return result
}

func (n *NgramIndex) getNextTokenRecords(colIdx int, fromRow int, toRow int, prevTokens []int, result *NgramSearchResult) {
	col := n.values[colIdx]
	for i := fromRow; i <= toRow; i++ {
		idx := col[i]
		currNgram := append(prevTokens, idx.index)
		if colIdx == len(n.values)-1 {
			result.addValue(currNgram)

		} else {
			nextFromIdx := 0
			if fromRow > 0 {
				nextFromIdx = col[i-1].upTo + 1
			}
			nextToIdx := idx.upTo
			n.getNextTokenRecords(colIdx+1, nextFromIdx, nextToIdx, currNgram, result)
		}
	}
}

// NewNgramIndex creates a new empty instance of NgramIndex
func NewNgramIndex(ngramSize int, initialLength int) *NgramIndex {
	ans := &NgramIndex{
		values: make([]indexColumn, ngramSize),
	}
	for i := range ans.values {
		ans.values[i] = make(indexColumn, initialLength)
	}
	return ans
}

// ----------------------------------------------------------------------------

// SearchableIndex is a higher-level representation
// of ngram-index with some functions allowing searching
type SearchableIndex struct {
	index  *NgramIndex
	wstore *wstore.WordIndex
}

// GetNgramsOf returns all the n-grams with first word
// equal to the 'word' argument
func (si *SearchableIndex) GetNgramsOf(word string) *NgramSearchResult {
	var ans *NgramSearchResult
	w := si.wstore.Find(word)
	col0Idx := sort.Search(len(si.index.values[0]), func(i int) bool {
		return si.index.values[0][i].index >= w
	})
	if col0Idx == len(si.index.values[0]) {
		return ans
	}
	ans = si.index.GetNgramsAt(col0Idx)
	return ans
}

// OpenSearchableIndex creates a instance of SearchableIndex
// based on internal NgramIndex instance and WordIndex instance
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

// NewDynamicNgramIndex creates a new instance of DynamicNgramIndex
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

// GetIndex returns internal index structure
func (nib *DynamicNgramIndex) GetIndex() *NgramIndex {
	return nib.index
}

// GetInfo returns a brief human-readable
// information about the index
func (nib *DynamicNgramIndex) GetInfo() string {
	return nib.index.GetInfo()
}

// GetNgramsAt returns all the ngrams where the first word index equals position
func (nib *DynamicNgramIndex) GetNgramsAt(position int) *NgramSearchResult {
	return nib.index.GetNgramsAt(position)
}

// AddNgram adds a new n-gram represented as an array
// of indices to the index
func (nib *DynamicNgramIndex) AddNgram(ngram []int) {
	addedNewPos := true
	for i := len(ngram) - 1; i >= 0; i-- {
		addedNewPos = nib.addValue(i, ngram[i], addedNewPos)
	}
}

// Finish should be called once adding of n-grams
// is done. The method frees up some memory preallocated
// for new n-grams.
func (nib *DynamicNgramIndex) Finish() {
	for i, v := range nib.index.values {
		nib.index.values[i] = v[:nib.cursors[i]]
	}
}

func (nib *DynamicNgramIndex) addValue(tokenPos int, index int, nextPosChanged bool) bool {

	col := nib.index.values[tokenPos]
	if nib.cursors[tokenPos] >= len(col)-1 {
		nib.index.values[tokenPos] = append(col, make(indexColumn, nib.initialLength/2)...)
		col = nib.index.values[tokenPos]
	}
	upTo := 0
	if tokenPos < len(nib.cursors)-1 {
		upTo = nib.cursors[tokenPos+1]
	}
	addedNewPos := false
	if nib.cursors[tokenPos] == -1 || nib.index.values[tokenPos][nib.cursors[tokenPos]].index != index {
		nib.cursors[tokenPos]++
		col[nib.cursors[tokenPos]] = &indexItem{index: index, upTo: upTo}
		addedNewPos = true

	} else if nextPosChanged {
		col[nib.cursors[tokenPos]].upTo++
	}
	return addedNewPos

}

func createColIdxPath(colIdx int, dirPath string) string {
	return filepath.Join(dirPath, fmt.Sprintf(ColumnIndexFilenameMask, colIdx))
}

func (nib *DynamicNgramIndex) saveIndexColumn(colIdx int, dirPath string) error {
	dstPath := createColIdxPath(colIdx, dirPath)
	f, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY, 0664)
	defer f.Close()
	if err != nil {
		return err
	}
	fw := bufio.NewWriter(f)
	defer fw.Flush()
	data := nib.index.values[colIdx]
	binary.Write(fw, binary.LittleEndian, int64(len(data)))
	for _, idx := range data {
		err = binary.Write(fw, binary.LittleEndian, int64(idx.index))
		if err != nil {
			os.Remove(dstPath) // try to clean but don't care much
			return err
		}
		binary.Write(fw, binary.LittleEndian, int64(idx.upTo))
	}
	return nil
}

func loadIndexColumn(indexPath string) indexColumn {
	f, err := os.Open(indexPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	fr := bufio.NewReader(f)
	var colLength int64
	binary.Read(fr, binary.LittleEndian, &colLength)
	ans := make(indexColumn, colLength)
	for i := 0; i < int(colLength); i++ {
		var index, upTo int64
		binary.Read(fr, binary.LittleEndian, &index)
		binary.Read(fr, binary.LittleEndian, &upTo)
		ans[i] = &indexItem{index: int(index), upTo: int(upTo)}
	}
	return ans
}

// LoadNgramIndex loads index data from within
// a specified directory.
func LoadNgramIndex(dirPath string) *NgramIndex {
	colIdxPaths := make([]string, MaxNgramSize)
	ans := &NgramIndex{}
	for i := 0; i < MaxNgramSize; i++ {
		tmp := createColIdxPath(i, dirPath)
		if _, err := os.Stat(tmp); os.IsNotExist(err) {
			colIdxPaths = colIdxPaths[:i]
			break
		}
		colIdxPaths[i] = tmp
	}
	ans.values = make([]indexColumn, len(colIdxPaths))
	for i := range ans.values {
		ans.values[i] = loadIndexColumn(colIdxPaths[i])
	}
	return ans
}

// Save stores current index data to bunch of files
// within the provided directory.
func (nib *DynamicNgramIndex) Save(dirPath string) error {
	var err error
	for i := range nib.index.values {
		if err = nib.saveIndexColumn(i, dirPath); err != nil {
			return err
		}
	}
	return err
}
