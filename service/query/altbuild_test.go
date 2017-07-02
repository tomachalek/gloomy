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

package query

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRuneBuild(t *testing.T) {
	a := newState()
	c := a.addRune('x')
	assert.Equal(t, "x", c.asString())
}

func TestAddState(t *testing.T) {
	a := newState()
	a2 := a.addState()
	assert.True(t, a.children[0] == a2)
}

func TestAppendState(t *testing.T) {
	a := newState()
	a2 := newState()
	a3 := a.appendState(a2)
	assert.True(t, a.children[0] == a2)
	assert.True(t, a3 == a2)
}

func TestAddAlternativeChunks(t *testing.T) {
	a := newState()
	r := a.addRune('f')
	a2 := r.addState()
	r21 := a2.addRune('o')
	r21.addRune('x')
	r31 := a2.addRune('O')
	r31.addRune('X')
	assert.Equal(t, "o", a2.children[0].asString())
	assert.Equal(t, "O", a2.children[1].asString())
	assert.Equal(t, 2, len(a2.children))
}

func TestGetEnd(t *testing.T) {
	a := newState()
	x1 := a.addRune('a')
	x2 := a.addRune('b')
	x3 := a.addRune('c')
	b := newState()
	x1.appendState(b)
	x2.appendState(b)
	x3.appendState(b)

	assert.Equal(t, b, a.getLast())
}

func TestExport(t *testing.T) {
	a := newState()
	r := a.addRune('f')

	a2 := r.addState()
	r21 := a2.addRune('o')
	r22 := r21.addRune('x')
	a3 := r22.addState()
	a3.addRune('1')

	a3.addRune('2')

	a5 := a2.addState()
	r51 := a5.addRune('O')
	r51.addRune('X')

	alts := a.getAll()
	assert.Equal(t, 3, len(alts))
	assert.Equal(t, "fox1", alts[0])
	assert.Equal(t, "fox2", alts[1])
	assert.Equal(t, "fOX", alts[2])
}

func TestRemoveChild(t *testing.T) {
	a := newState()
	x1 := a.addState()
	x2 := a.addState()
	x3 := a.addState()
	x4 := a.addState()
	x5 := a.addState()
	assert.Equal(t, 5, len(a.children))

	a.removeChild(x3)
	assert.Equal(t, 4, len(a.children))
	assert.Equal(t, x1, a.children[0])
	assert.Equal(t, x2, a.children[1])
	assert.Equal(t, x4, a.children[2])
	assert.Equal(t, x5, a.children[3])
}
