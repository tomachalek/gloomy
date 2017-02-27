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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNgramsCmp(t *testing.T) {
	ans := ngramsCmp([]string{"foo", "bar", "baz"}, []string{"foo", "bar", "baz"})
	assert.Equal(t, 0, ans)
}

func TestNgramsCmpFirstBigger(t *testing.T) {
	ans := ngramsCmp([]string{"zzz", "bar", "baz"}, []string{"foo", "bar", "baz"})
	assert.Equal(t, 1, ans)
}

func TestNgramsCmpFirstBigger2(t *testing.T) {
	ans := ngramsCmp([]string{"foo", "bar", "baz"}, []string{"foo", "bar", "bay"})
	assert.Equal(t, 1, ans)
}

func TestNgramsCmpFirstSmaller(t *testing.T) {
	ans := ngramsCmp([]string{"eon", "bar", "baz"}, []string{"foo", "bar", "baz"})
	assert.Equal(t, -1, ans)
}

func TestEmptyNgramsCmp(t *testing.T) {
	ans := ngramsCmp([]string{}, []string{})
	assert.Equal(t, 0, ans)
}

// --------

func TestNgramListAdd(t *testing.T) {
	nl := NgramList{}
	v := []string{"foo", "bar"}
	nl.Add(v)
	assert.Equal(t, 0, ngramsCmp(nl.root.ngram, v))
}

func TestNgramListAddMulti(t *testing.T) {
	n := NgramList{}
	v1 := []string{"foo", "bar"}
	n.Add(v1)
	v2 := []string{"boo", "bar"}
	n.Add(v2)
	v3 := []string{"moo", "bar"}
	n.Add(v3)
	v4 := []string{"zoo", "bar"}
	n.Add(v4)

	assert.Equal(t, v1[0], n.root.ngram[0])
	assert.Equal(t, v2[0], n.root.left.ngram[0])
	assert.Equal(t, v3[0], n.root.right.ngram[0])
	assert.Equal(t, v4[0], n.root.right.right.ngram[0])
}
