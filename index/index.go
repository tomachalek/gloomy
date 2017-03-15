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
	"log"
	"strings"
)

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

func (n *NgramIndex) GetNgramsAt(position int) []int {
	ans := make([]int, 10) // TODO (testing)
	ans[0] = n.values[0][position].index
	ans[1] = n.values[1][n.values[0][position].upTo].index
	return ans
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

func (nib *DynamicNgramIndex) GetInfo() string {
	return nib.index.GetInfo()
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
	upTo := -1
	if tokenPos < len(nib.cursors)-1 {
		upTo = nib.cursors[tokenPos+1]
	}
	if nib.cursors[tokenPos] == -1 || nib.index.values[tokenPos][nib.cursors[tokenPos]].index != index {
		nib.cursors[tokenPos]++
		col[nib.cursors[tokenPos]] = &IndexItem{index: index, upTo: upTo}
		log.Printf("adding pos %d, record [%d, %d]", tokenPos, index, upTo)

	} else {
		log.Println("Not moving cursor on ", index)
	}
}
