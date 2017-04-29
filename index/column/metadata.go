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

// MetadataWriter is used for writing metadata attributes
// during indexing. It collects individual metadata
// columns into a single object providing similar methods
// (Get, Set, Extend, Size, Resize, Save).
type MetadataWriter struct {
	dicts ArgsWriterList
	cols  []AttrValColumn
}

func (mw *MetadataWriter) ForEachArg(fn func(i int, v *ArgsDictWriter, col AttrValColumn)) {
	for i := 0; i < len(mw.dicts); i++ { // we expect len(dicts) == len(cols)
		fn(i, mw.dicts[i], mw.cols[i])
	}
}

func (mw *MetadataWriter) NumCols() int {
	return len(mw.dicts)
}

func (mw *MetadataWriter) Get(idx int) []AttrVal {
	ans := make([]AttrVal, len(mw.cols))
	for i, v := range mw.cols {
		ans[i] = v.Get(idx)
	}
	return ans
}

func (mw *MetadataWriter) Set(idx int, val []AttrVal) {
	for i := 0; i < len(mw.cols); i++ {
		mw.cols[i].Set(idx, val[i])
	}
}

func (mw *MetadataWriter) Extend(appendSize int) {
	for _, v := range mw.cols {
		v.Extend(appendSize)
	}
}

func (mw *MetadataWriter) Size() int {
	if len(mw.cols) > 0 {
		return mw.cols[0].Size() // we expect all the columns to have the same len/cap
	}
	return 0
}

func (mw *MetadataWriter) Shrink(rightIdx int) {
	for _, v := range mw.cols {
		v.Shrink(rightIdx)
	}
}

func (mw *MetadataWriter) Save(dirPath string) error {
	for _, avc := range mw.cols {
		err := avc.Save(dirPath)
		if err != nil {
			return err
		}
	}
	for _, d := range mw.dicts {
		err := d.Save(dirPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewMetadataWriter(attrs map[string]string) *MetadataWriter {
	cols := make([]AttrValColumn, len(attrs))
	dicts := make([]*ArgsDictWriter, len(attrs))
	i := 0
	var err error
	for k, v := range attrs {
		cols[i], err = NewMetadataColumn(k, v, 0) // TODO size
		if err != nil {
			panic(err)
		}
		dicts[i] = NewArgsDictWriter(k)
		i++
	}
	return &MetadataWriter{cols: cols, dicts: dicts}
}

// -----------------------------------------------------------------------

// MetadataReader is used for reading indexed
// metadata attributes when in search mode.
// It collects individual metadata columns
// into a single object.
type MetadataReader struct {
	dicts ArgsReaderList
	cols  []AttrValColumn
}

func (mr *MetadataReader) LoadChunk(fromIdx int, toIdx int) {
	for _, v := range mr.cols {
		v.LoadChunk(fromIdx, toIdx)
	}
}

func (mr *MetadataReader) Get(idx int) []string {
	ans := make([]string, len(mr.cols))
	for i, v := range mr.cols {
		ans[i] = mr.dicts[i].index[v.Get(idx)]
	}
	return ans
}

func LoadMetadataReader(dirPath string, attrNames []string) (*MetadataReader, error) {
	cols := make([]AttrValColumn, len(attrNames))
	dicts := make([]*ArgsDictReader, len(attrNames))
	for i, attrName := range attrNames {
		tmp, err := LoadMetadataColumn(attrName, dirPath)
		if err != nil {
			return nil, err
		}
		cols[i] = tmp
		var err2 error
		dicts[i], err2 = LoadArgsDict(dirPath, attrName)
		if err2 != nil {
			return nil, err2
		}
	}

	return &MetadataReader{cols: cols, dicts: dicts}, nil
}
