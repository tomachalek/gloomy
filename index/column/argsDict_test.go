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

func TestNewArgsDictWriter(t *testing.T) {
	adw := NewArgsDictWriter("foo")
	assert.Equal(t, "foo", adw.Name())
	assert.Equal(t, 0, adw.counter)
	assert.NotNil(t, adw.index)
	assert.Equal(t, 0, len(adw.index))
}

func TestArgsDictWriterAddValue(t *testing.T) {
	adw := NewArgsDictWriter("foo")
	adw.AddValue("xxx")

	_, ok := adw.index["xxx"]
	assert.True(t, ok)
	assert.Equal(t, 1, len(adw.index))
}
