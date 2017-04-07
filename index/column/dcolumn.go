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
	"log"
	"os"
	"path/filepath"
)

type MetadataItem struct {
	Count uint32
	Flags uint64
	Date  int16
}

type MetadataColumn struct {
	data     []*MetadataItem
	dataPath string
	fullSize int
	offset   int
}

func (mc *MetadataColumn) Get(idx int) *MetadataItem {
	return mc.data[idx-mc.offset]
}

func (mc *MetadataColumn) Set(idx int, it *MetadataItem) {
	mc.data[idx-mc.offset] = it
}

func (mc *MetadataColumn) Save(dirPath string) error {
	dstPath := CreateMetadataIdxPath(dirPath)
	f, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664)
	defer f.Close()
	if err != nil {
		return err
	}
	fw := bufio.NewWriter(f)
	defer fw.Flush()

	binary.Write(fw, binary.LittleEndian, int64(len(mc.data)))
	for _, item := range mc.data {
		// TODO test for errors
		binary.Write(fw, binary.LittleEndian, item.Count)
		binary.Write(fw, binary.LittleEndian, item.Flags)
		binary.Write(fw, binary.LittleEndian, item.Date)
	}
	mc.dataPath = dstPath
	return nil
}

func (mc *MetadataColumn) Size() int {
	return len(mc.data)
}

// Resize removes spare array items
func (mc *MetadataColumn) Resize(rightIdx int) {
	mc.data = mc.data[:rightIdx]
}

func (mc *MetadataColumn) Extend(appendSize int) {
	mc.data = append(mc.data, make([]*MetadataItem, appendSize)...)
}

func (mc *MetadataColumn) LoadChunk(fromIdx int, toIdx int) {
	f, err := os.Open(mc.dataPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var colLen int64
	binary.Read(f, binary.LittleEndian, &colLen)
	mc.fullSize = int(colLen)

	if fromIdx > 0 {
		fromIdx-- // we must know 'upTo' value of previous index item
	}

	f.Seek(int64(fromIdx*(8+4+2)+8), os.SEEK_SET)
	newLength := toIdx + 1 - fromIdx

	rawData := make([]byte, newLength*(8+4+2))
	_, err = f.Read(rawData)
	if err != nil {
		panic(err)
	}
	indexData := bytes.NewReader(rawData)

	if len(mc.data) != cap(mc.data) {
		mc.data = mc.data[:cap(mc.data)]
	}
	lengthDiff := newLength - len(mc.data)
	if lengthDiff > 0 {
		mc.data = append(mc.data, make([]*MetadataItem, lengthDiff)...)

	} else if lengthDiff < 0 {
		mc.data = mc.data[:newLength]
	}
	for i := 0; i < newLength; i++ {
		var count uint32
		var flags uint64
		var date int16
		binary.Read(indexData, binary.LittleEndian, &count)
		binary.Read(indexData, binary.LittleEndian, &flags)
		binary.Read(indexData, binary.LittleEndian, &date)
		mc.data[i] = &MetadataItem{
			Count: count,
			Flags: flags,
			Date:  date,
		}
	}
	mc.offset = fromIdx
}

// -------------------------------------------------------

func NewBoundMetadataColumn(dataPath string) *MetadataColumn {
	return &MetadataColumn{data: make([]*MetadataItem, 0), dataPath: dataPath}
}

func NewMetadataColumn(size int) *MetadataColumn {
	return &MetadataColumn{data: make([]*MetadataItem, size)}
}

func CreateMetadataIdxPath(dirPath string) string {
	return filepath.Join(dirPath, "metadata.idx")
}
