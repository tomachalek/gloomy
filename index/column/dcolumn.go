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
	"io"
	"log"
	"os"
	"path/filepath"
)

var _ = log.Print

type AttrVal int

type AttrValColumn interface {
	Name() string

	Seek(file *os.File, positions int)
	Get(idx int) AttrVal
	Set(idx int, it AttrVal)
	Save(dirPath string) error
	Extend(appendSize int)
	Resize(rightIdx int)

	Size() int

	// StoredSize returns a total number of items
	// the column possesses on disk
	StoredSize() int

	// UnitSize specifies a number of bytes needed
	// to store a single record to disk
	UnitSize() int

	// LoadChunk loads a partial data starting from
	// index fromIdx (incl.) up to toIdx (incl.)
	// Please note that while this function always
	// guarantees that required data are loaded,
	// it may also load some additional items.
	// I.e. loading (1, 5) does not imply Size() == 5
	LoadChunk(fromIdx int, toIdx int)

	ForEach(func(int, interface{}))

	DataPath() string

	ReadItem(reader io.Reader, idx int)
}

type Metadata struct {
	cols  []AttrValColumn
	attrs map[string]int
}

func (m *Metadata) Get(idx int) []AttrVal {
	ans := make([]AttrVal, len(m.cols))
	for i, v := range m.cols {
		ans[i] = v.Get(idx)
	}
	return ans
}

func (m *Metadata) LoadChunk(fromIdx int, toIdx int) {
	for _, v := range m.cols {
		v.LoadChunk(fromIdx, toIdx)
	}
}

func (m *Metadata) Extend(appendSize int) {
	for _, v := range m.cols {
		v.Extend(appendSize)
	}
}

func (m *Metadata) Size() int {
	if len(m.cols) > 0 {
		return m.cols[0].Size() // we expect all the columns to have the same len/cap
	}
	return 0
}

func (m *Metadata) Resize(rightIdx int) {
	for _, v := range m.cols {
		v.Resize(rightIdx)
	}
}

func (m *Metadata) Save(dirPath string) error {
	for _, v := range m.cols {
		err := v.Save(dirPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewMetadata(attrs map[string]string) *Metadata {
	cols := make([]AttrValColumn, len(attrs))
	attrMap := make(map[string]int)
	i := 0
	var err error
	for k, v := range attrs {
		cols[i], err = NewMetadataColumn(k, v, 0) // TODO size
		if err != nil {
			panic(err)
		}
		attrMap[k] = i
	}
	return &Metadata{cols: cols, attrs: attrMap}
}

func LoadMetadata(dirPath string) *Metadata {
	attrs := make(map[string]int) // TODO apply conf
	cols := make([]AttrValColumn, len(attrs))
	return &Metadata{cols: cols, attrs: attrs}
}

// ----------------------------------------------------------------------------

func createColumnPath(colIdent string, dirPath string) string {
	return filepath.Join(dirPath, fmt.Sprintf("column_%s.idx", colIdent))
}

func saveAttrColumn(col AttrValColumn, dirPath string) (string, error) {
	dstPath := createColumnPath(col.Name(), dirPath)
	f, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664)
	defer f.Close()
	if err != nil {
		return dstPath, err
	}
	fw := bufio.NewWriter(f)
	defer fw.Flush()

	// write num of items
	binary.Write(fw, binary.LittleEndian, int64(col.Size()))

	// write item length in bytes and some spare zeros
	binary.Write(fw, binary.LittleEndian, []int8{int8(col.UnitSize() * 8), 0, 0, 0, 0, 0, 0, 0})

	// write data
	col.ForEach(func(i int, v interface{}) {
		binary.Write(fw, binary.LittleEndian, v)
	})
	return dstPath, nil
}

// return actual offset (it is different from 'fromIdx')
func loadAttrColumnChunk(col AttrValColumn, fromIdx int, toIdx int) int {
	f, err := os.Open(col.DataPath())
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if fromIdx > 0 {
		fromIdx-- // we must know 'upTo' value of previous index item
	}
	col.Seek(f, fromIdx)
	newLength := toIdx + 1 - fromIdx

	rawData := make([]byte, newLength*col.UnitSize())
	_, err = f.Read(rawData)
	if err != nil {
		panic(err)
	}
	indexData := bytes.NewReader(rawData)

	lengthDiff := newLength - col.Size()
	if lengthDiff > 0 {
		col.Extend(lengthDiff)

	} else if lengthDiff < 0 {
		col.Resize(newLength)
	}
	for i := fromIdx; i <= toIdx; i++ {
		col.ReadItem(indexData, i-fromIdx)
	}
	return fromIdx
}

// ----------------------------------------------------------------------------

type Column8 struct {
	data     []uint8
	dataPath string
	fullSize int
	offset   int
	name     string
}

func (c *Column8) Name() string {
	return c.name
}

func (c *Column8) Get(idx int) AttrVal {
	return AttrVal(c.data[idx-c.offset])
}

func (c *Column8) Set(idx int, it AttrVal) {
	c.data[idx-c.offset] = uint8(it)
}

func (c *Column8) Size() int {
	return len(c.data)
}

func (c *Column8) StoredSize() int {
	return c.fullSize
}

func (c *Column8) UnitSize() int {
	return 1
}

// Resize removes spare array items
func (c *Column8) Resize(rightIdx int) {
	c.data = c.data[:rightIdx]
}

func (c *Column8) Extend(appendSize int) {
	c.data = append(c.data, make([]uint8, appendSize)...)
}

func (c *Column8) ForEach(fn func(int, interface{})) {
	for i, v := range c.data {
		fn(i, v)
	}
}

func (c *Column8) DataPath() string {
	return c.dataPath
}

func (c *Column8) Seek(file *os.File, numPos int) {
	file.Seek(int64(numPos+16), os.SEEK_SET)
}

func (c *Column8) Save(dirPath string) error {
	dstPath, err := saveAttrColumn(c, dirPath)
	if err != nil {
		return err
	}
	c.dataPath = dstPath
	return nil
}

func (c *Column8) LoadChunk(fromIdx int, toIdx int) {
	c.offset = loadAttrColumnChunk(c, fromIdx, toIdx)
}

func (c *Column8) ReadItem(reader io.Reader, idx int) {
	var v uint8
	binary.Read(reader, binary.LittleEndian, &v)
	c.data[idx] = v
}

// -------------------------------------------------------

type Column32 struct {
	data     []uint32
	dataPath string
	fullSize int
	offset   int
	name     string
}

func (c *Column32) Name() string {
	return c.name
}

func (c *Column32) Get(idx int) AttrVal {
	return AttrVal(c.data[idx-c.offset])
}

func (c *Column32) Set(idx int, it AttrVal) {
	c.data[idx-c.offset] = uint32(it)
}

func (c *Column32) Size() int {
	return len(c.data)
}

func (c *Column32) StoredSize() int {
	return c.fullSize
}

func (c *Column32) UnitSize() int {
	return 4
}

// Resize removes spare array items
func (c *Column32) Resize(rightIdx int) {
	c.data = c.data[:rightIdx]
}

func (c *Column32) Extend(appendSize int) {
	c.data = append(c.data, make([]uint32, appendSize)...)
}

func (c *Column32) ForEach(fn func(int, interface{})) {
	for i, v := range c.data {
		fn(i, v)
	}
}

func (c *Column32) DataPath() string {
	return c.dataPath
}

func (c *Column32) Seek(file *os.File, numPos int) {
	file.Seek(int64(numPos*c.UnitSize()), os.SEEK_SET)
}

func (c *Column32) Save(dirPath string) error {
	dstPath, err := saveAttrColumn(c, dirPath)
	if err != nil {
		return err
	}
	c.dataPath = dstPath
	return nil
}

func (c *Column32) LoadChunk(fromIdx int, toIdx int) {
	c.offset = loadAttrColumnChunk(c, fromIdx, toIdx)
}

func (c *Column32) ReadItem(reader io.Reader, idx int) {
	var v uint32
	binary.Read(reader, binary.LittleEndian, &v)
	c.data[idx] = v
}

// ---------------------------------------------------------

func NewMetadataColumn(ident string, typeIdent string, size int) (AttrValColumn, error) {
	switch typeIdent {
	case "col8":
		return &Column8{name: ident, data: make([]uint8, size), fullSize: size}, nil
	case "col32":
		return &Column32{name: ident, data: make([]uint32, size), fullSize: size}, nil
	default:
		return nil, fmt.Errorf("Unknown column type %s", typeIdent)
	}
}

//
// TODO rename to NewBoundMetadataColumn
func LoadMetadataColumn(ident string, dirPath string) (AttrValColumn, error) {
	f, err := os.Open(createColumnPath(ident, dirPath))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var colLen int64
	binary.Read(f, binary.LittleEndian, &colLen)
	flags := make([]int8, 8)
	binary.Read(f, binary.LittleEndian, flags)
	var ans AttrValColumn
	var ansErr error
	switch flags[0] {
	case 8:
		ans = &Column8{fullSize: int(colLen), dataPath: f.Name()}
	case 32:
		ans = &Column32{fullSize: int(colLen), dataPath: f.Name()}
	default:
		ansErr = fmt.Errorf("Cannot load metadata column, unsupported item length %d", flags[0])
	}
	return ans, ansErr
}

func LoadCountsColumn(dirPath string) (AttrValColumn, error) {
	return LoadMetadataColumn("_counts", dirPath)
}

func NewCountsColumn(size int) AttrValColumn {
	return &Column32{name: "_counts", data: make([]uint32, size)}
}
