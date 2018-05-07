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

import "github.com/tomachalek/gloomy/index/column"

// ------------------------------------------------

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

// MetadataWriter provides access to attached
// metadata index writer
func (nib *DynamicNgramIndex) MetadataWriter() *column.MetadataWriter {
	return nib.metadataWriter
}

// GetNgramsAt returns all the ngrams where the first word index equals position
func (nib *DynamicNgramIndex) GetNgramsAt(position int) *NgramSearchResult {
	return nib.index.GetNgramsAt(position)
}

// AddNgram adds a new n-gram represented as an array
// of indices to the index
func (nib *DynamicNgramIndex) AddNgram(ngram []int, count int, metadataList *column.MetadataList) {
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
	nib.metadataWriter.Set(lastPos, metadataList)
	// TODO add metadata
}

// findSplitPosition returns a position within an n-gram (i.e. value from 0...n-1)
// where the currently stored n-gram "tree" should split to create a new branch.
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
