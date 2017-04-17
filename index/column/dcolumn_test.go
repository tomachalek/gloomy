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
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func getFilePath() string {
	_, d, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(d))), "testdata")
}

func TestCreateColumnPath(t *testing.T) {
	ans := createColumnPath("doc_title", "/path/to/a/dir")
	assert.Equal(t, filepath.Join("/", "path", "to", "a", "dir", "column_doc_title.idx"), ans)
}

func TestCreateColumnPathSpareSlash(t *testing.T) {
	ans := createColumnPath("doc_title", "/path/to/a/dir/")
	assert.Equal(t, filepath.Join("/", "path", "to", "a", "dir", "column_doc_title.idx"), ans)
}

// Column 8

func TestMethodsColumn8(t *testing.T) {
	col, err := NewMetadataColumn("foo", "col8", 10)
	assert.Nil(t, err)

	assert.Equal(t, "foo", col.Name())
	assert.Equal(t, 1, col.UnitSize())
	assert.Equal(t, "", col.DataPath())
	assert.Equal(t, 10, col.Size())
	assert.Equal(t, 10, col.StoredSize())
}

func TestMethodsColumn8Extend(t *testing.T) {
	col, _ := NewMetadataColumn("foo", "col8", 10)
	col.Extend(5)

	assert.Equal(t, 15, col.Size())
	assert.Equal(t, 10, col.StoredSize())
}

func TestMethodsColumn8ExtendBound(t *testing.T) {
	col, _ := LoadMetadataColumn("10items", getFilePath())
	col.Extend(5)

	assert.Equal(t, 5, col.Size())
	assert.Equal(t, 10, col.StoredSize())
}

func TestMethodsColumn8Resize(t *testing.T) {
	col, _ := NewMetadataColumn("foo", "col8", 10)
	col.Resize(2)

	assert.Equal(t, 2, col.Size())
	assert.Equal(t, 10, col.StoredSize())
}

func TestMethodsColumn8ResizeOverflow(t *testing.T) {
	col, _ := NewMetadataColumn("foo", "col8", 10)

	assert.Panics(t, func() {
		col.Resize(20)
	})
}

func TestMethodsColumn8UnknownType(t *testing.T) {
	col, err := NewMetadataColumn("foo", "col100", 10)
	assert.Error(t, err)
	assert.Nil(t, col)
}

func TestMethodsColumn8Save(t *testing.T) {
	col, _ := NewMetadataColumn("foo", "col8", 10)
	for i := 0; i < 10; i++ {
		col.Set(i, AttrVal(i))
	}
	testdataPath := getFilePath()
	col.Save(testdataPath)

	indexPath := filepath.Join(testdataPath, "column_foo.idx")
	assert.Equal(t, indexPath, col.DataPath())

	f, err := os.Open(indexPath)
	assert.Nil(t, err)
	defer f.Close()

	var dSize int64
	colTypeEtc := make([]int8, 8)
	binary.Read(f, binary.LittleEndian, &dSize)
	assert.Equal(t, int64(10), dSize)
	binary.Read(f, binary.LittleEndian, colTypeEtc)
	assert.Equal(t, int8(8), colTypeEtc[0])
}

func TestColumn8LoadChunk(t *testing.T) {
	col, _ := LoadMetadataColumn("10items", getFilePath())
	assert.Equal(t, filepath.Join(getFilePath(), "column_10items.idx"), col.DataPath())
	assert.Equal(t, 0, col.Size())

	col.LoadChunk(1, 5)
	// NOTE - cannot test Size() here, see LoadChunk specs.
	for i := 1; i < 5; i++ {
		assert.Equal(t, AttrVal(i), col.Get(i))
	}

}
