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
	"bytes"
	"encoding/binary"
	"os"
)

const (
	recNumberSizeBytes = 8
)

type indexItem struct {
	index int
	upTo  int
}

type indexColumn struct {
	data     []*indexItem
	fullSize int
	dataPath string
	offset   int
}

func (ic *indexColumn) size() int {
	return len(ic.data)
}

func (ic *indexColumn) get(idx int) *indexItem {
	return ic.data[idx-ic.offset]
}

func (ic *indexColumn) set(idx int, it *indexItem) {
	ic.data[idx-ic.offset] = it
}

func (ic *indexColumn) extend(appendSize int) {
	ic.data = append(ic.data, make([]*indexItem, appendSize)...)
}

// slice removes spare array items
func (ic *indexColumn) resize(rightIdx int) {
	ic.data = ic.data[:rightIdx]
}

func (ic *indexColumn) loadWholeChunk() {
	f, err := os.Open(ic.dataPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	fr := bufio.NewReader(f)
	var colLength int64
	binary.Read(fr, binary.LittleEndian, &colLength)
	ic.fullSize = int(colLength)

	if len(ic.data) != cap(ic.data) {
		ic.data = ic.data[:cap(ic.data)]
	}
	lengthDiff := ic.fullSize - len(ic.data)
	if lengthDiff > 0 {
		ic.data = append(ic.data, make([]*indexItem, lengthDiff)...)

	} else {
		ic.data = ic.data[:ic.fullSize]
	}

	for i := 0; i < ic.fullSize; i++ {
		var index, upTo int64
		binary.Read(fr, binary.LittleEndian, &index)
		binary.Read(fr, binary.LittleEndian, &upTo)
		ic.data[i] = &indexItem{index: int(index), upTo: int(upTo)}
	}
}

func (ic *indexColumn) loadChunk(fromIdx int, toIdx int) {
	f, err := os.Open(ic.dataPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var colLen int64
	binary.Read(f, binary.LittleEndian, &colLen)
	ic.fullSize = int(colLen)

	if fromIdx > 0 {
		fromIdx-- // we must know 'upTo' value of previous index item
	}

	f.Seek(int64(fromIdx*recNumberSizeBytes*2+recNumberSizeBytes), os.SEEK_SET)
	newLength := toIdx + 1 - fromIdx

	rawData := make([]byte, newLength*recNumberSizeBytes*2)
	_, err = f.Read(rawData)
	if err != nil {
		panic(err)
	}
	indexData := bytes.NewReader(rawData)

	if len(ic.data) != cap(ic.data) {
		ic.data = ic.data[:cap(ic.data)]
	}
	lengthDiff := newLength - len(ic.data)
	if lengthDiff > 0 {
		ic.data = append(ic.data, make([]*indexItem, lengthDiff)...)

	} else if lengthDiff < 0 {
		ic.data = ic.data[:newLength]
	}
	for i := 0; i < newLength; i++ {
		var index, upTo int64
		binary.Read(indexData, binary.LittleEndian, &index)
		binary.Read(indexData, binary.LittleEndian, &upTo)
		ic.data[i] = &indexItem{index: int(index), upTo: int(upTo)}
	}
	ic.offset = fromIdx
}

func newBoundIndexColumn(dataPath string) *indexColumn {
	return &indexColumn{data: make([]*indexItem, 0), dataPath: dataPath}
}

func newIndexColumn(size int) *indexColumn {
	return &indexColumn{data: make([]*indexItem, size)}
}
