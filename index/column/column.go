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

// This file contains an implementation of IndexColumn which is
// a storage for n-ngram tree nodes with the same depth (= position
// within an n-gram)

package column

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

const (
	recNumberSizeBytes = 8

	// ColumnIndexFilenameMask specifies a filename for
	// an n-gram column (= data for all the variants at
	// a specific position within an n-gram)
	columnIndexFilenameMask = "idx_col_%d.idx"
)

// ----------------------------------------------------------------------------

type IndexItem struct {
	Index int
	UpTo  int
}

type IndexColumn struct {
	data     []*IndexItem
	fullSize int
	dataPath string
	offset   int
}

func (ic *IndexColumn) Size() int {
	return len(ic.data)
}

func (ic *IndexColumn) Get(idx int) *IndexItem {
	return ic.data[idx-ic.offset]
}

func (ic *IndexColumn) Set(idx int, it *IndexItem) {
	ic.data[idx-ic.offset] = it
}

func (ic *IndexColumn) Extend(appendSize int) {
	ic.data = append(ic.data, make([]*IndexItem, appendSize)...)
}

// Resize removes spare array items
func (ic *IndexColumn) Shrink(rightIdx int) {
	if rightIdx >= len(ic.data) {
		panic("Cannot shrink to a larger column")
	}
	ic.data = ic.data[:rightIdx]
}

func (ic *IndexColumn) Save(colIdx int, dirPath string) error {
	dstPath := CreateColIdxPath(colIdx, dirPath)
	f, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664)
	defer f.Close()
	if err != nil {
		return err
	}
	fw := bufio.NewWriter(f)
	defer fw.Flush()
	binary.Write(fw, binary.LittleEndian, int64(len(ic.data)))
	for _, idx := range ic.data {
		err = binary.Write(fw, binary.LittleEndian, int64(idx.Index))
		if err != nil {
			os.Remove(dstPath) // try to clean but don't care much
			return err
		}
		binary.Write(fw, binary.LittleEndian, int64(idx.UpTo))
	}
	ic.dataPath = dstPath
	return nil
}

func (ic *IndexColumn) LoadWholeChunk() {
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
		ic.data = append(ic.data, make([]*IndexItem, lengthDiff)...)

	} else {
		ic.data = ic.data[:ic.fullSize]
	}

	for i := 0; i < ic.fullSize; i++ {
		var index, upTo int64
		binary.Read(fr, binary.LittleEndian, &index)
		binary.Read(fr, binary.LittleEndian, &upTo)
		ic.data[i] = &IndexItem{Index: int(index), UpTo: int(upTo)}
	}
}

func (ic *IndexColumn) LoadChunk(fromIdx int, toIdx int) {
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
		ic.data = append(ic.data, make([]*IndexItem, lengthDiff)...)

	} else if lengthDiff < 0 {
		ic.data = ic.data[:newLength]
	}
	for i := 0; i < newLength; i++ {
		var index, upTo int64
		binary.Read(indexData, binary.LittleEndian, &index)
		binary.Read(indexData, binary.LittleEndian, &upTo)
		ic.data[i] = &IndexItem{Index: int(index), UpTo: int(upTo)}
	}
	ic.offset = fromIdx
}

// ----------------------------------------------------------------------------

func NewBoundIndexColumn(dataPath string) *IndexColumn {
	return &IndexColumn{data: make([]*IndexItem, 0), dataPath: dataPath}
}

func NewIndexColumn(size int) *IndexColumn {
	return &IndexColumn{data: make([]*IndexItem, size)}
}

func CreateColIdxPath(colIdx int, dirPath string) string {
	return filepath.Join(dirPath, fmt.Sprintf(columnIndexFilenameMask, colIdx))
}
