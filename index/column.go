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

type indexColumn struct {
	data []*indexItem
}

func (ic *indexColumn) size() int {
	return len(ic.data)
}

func (ic *indexColumn) get(idx int) *indexItem {
	return ic.data[idx]
}

func (ic *indexColumn) set(idx int, it *indexItem) {
	ic.data[idx] = it
}

func (ic *indexColumn) extend(appendSize int) {
	ic.data = append(ic.data, make([]*indexItem, appendSize)...)
}

// slice removes spare array capacity
func (ic *indexColumn) slice(rightIdx int) {
	ic.data = ic.data[:rightIdx]
}

func newIndexColumn(size int) *indexColumn {
	return &indexColumn{data: make([]*indexItem, size)}
}
