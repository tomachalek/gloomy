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

func TestChunkBuild(t *testing.T) {
	c := newChunk()
	c.addRune('x')
	c.addRune('y')
	c.addRune('z')
	assert.Equal(t, "xyz", c.asString())
}

func TestChunkResize(t *testing.T) {
	c := newChunk()
	for i := 0; i < 15; i++ {
		c.addRune('x')
	}
	assert.Equal(t, "xxxxxxxxxxxxxxx", c.asString())
}

func TestAddAlternative(t *testing.T) {
	c := newChunk()
	c.addRune('f')
	a := c.addAlternative()
	assert.True(t, a == c.next)
}

func TestAddAlternativeChunks(t *testing.T) {
	c := newChunk()
	c.addRune('f')
	a := c.addAlternative()
	c1 := a.addChunk()
	c1.addRune('o')
	c1.addRune('x')
	c2 := a.addChunk()
	c2.addRune('O')
	c2.addRune('X')
	assert.Equal(t, "ox", a.children[0].asString())
	assert.Equal(t, "OX", a.children[1].asString())
	assert.Equal(t, 2, len(a.children))
}

func TestExport(t *testing.T) {
	c := newChunk()
	c.addRune('f')
	a := c.addAlternative()
	c1 := a.addChunk()
	c1.addRune('o')
	c1.addRune('x')
	c2 := a.addChunk()
	c2.addRune('O')
	c2.addRune('X')
	ax := c1.addAlternative()
	c3 := ax.addChunk()
	c3.addRune('1')
	c4 := ax.addChunk()
	c4.addRune('2')
	alts := c.getAll()
	assert.Equal(t, 3, len(alts))
	assert.Equal(t, "fox1", alts[0])
	assert.Equal(t, "fox2", alts[1])
	assert.Equal(t, "fOX", alts[2])
}

//
// [c]---<a>---[c1]---<a1>
//        |            |----[c3]----<a3>
//        |            |-----[c4]----<a4>
//        |
//        |----[c2]---<a2>
//
func TestGetLeaves(t *testing.T) {
	c := newChunk()

	a := c.addAlternative()
	c1 := a.addChunk()
	a1 := c1.addAlternative()

	c2 := a.addChunk()
	a2 := c2.addAlternative()

	c3 := a1.addChunk()
	a3 := c3.addAlternative()

	c4 := a1.addChunk()
	a4 := c4.addAlternative()

	leaves := a.getLeaves()
	assert.Equal(t, 3, len(leaves))
	assert.Equal(t, a3, leaves[0])
	assert.Equal(t, a4, leaves[1])
	assert.Equal(t, a2, leaves[2])
}
