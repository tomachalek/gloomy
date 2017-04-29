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
	"github.com/stretchr/testify/assert"
	"testing"
)

func createColumSize3() (*IndexColumn, []*IndexItem) {
	ic := NewIndexColumn(3)
	values := make([]*IndexItem, 3)
	for i := 0; i < 3; i++ {
		values[i] = &IndexItem{Index: i, UpTo: 0}
		ic.data[i] = values[i]
	}
	return ic, values
}

func TestNewIndexColumn(t *testing.T) {
	ic := NewIndexColumn(0)
	assert.NotNil(t, ic.data)
	assert.Equal(t, 0, ic.offset)
	assert.Equal(t, "", ic.dataPath)
	assert.Equal(t, 0, ic.fullSize)
}

func TestIndexColumnSize(t *testing.T) {
	size := 10
	ic := NewIndexColumn(size)
	assert.Equal(t, size, ic.Size())
}

func TestIndexColumnGet(t *testing.T) {
	ic, values := createColumSize3()

	assert.Equal(t, values[0], ic.Get(0))
	assert.Equal(t, values[1], ic.Get(1))
	assert.Equal(t, values[2], ic.Get(2))
}

func TestIndexColumnSet(t *testing.T) {
	ic, _ := createColumSize3()

	vNew := &IndexItem{Index: 10, UpTo: 100}
	ic.Set(0, vNew)
	assert.Equal(t, vNew, ic.Get(0))
	assert.Equal(t, 3, ic.Size())
}

func TestIndexColumnExtend(t *testing.T) {
	ic, _ := createColumSize3()
	ic.Extend(2)
	assert.Equal(t, 5, ic.Size())
	assert.Nil(t, ic.Get(3))
	assert.Nil(t, ic.Get(4))
}

func TestIndexColumnShrink(t *testing.T) {
	ic, _ := createColumSize3()
	ic.Shrink(1)
	assert.Equal(t, 1, ic.Size())
}

func TestIndexColumnShrinkOverflow(t *testing.T) {
	assert.Panics(t, func() {
		ic, _ := createColumSize3()
		ic.Shrink(10)
	})
}
