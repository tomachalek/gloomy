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

type chunk struct {
	value []rune
	next  *alternative
	curr  int
}

func (ch *chunk) addAlternative() *alternative {
	ch.next = &alternative{children: make([]*chunk, 10)}
	return ch.next
}

func (ch *chunk) addRune(v rune) {
	if ch.curr >= len(ch.value) {
		ch.value = append(ch.value, make([]rune, 10)...)
	}
	ch.value[ch.curr] = v
	ch.curr++
}

func (ch *chunk) asString() string {
	return string(ch.value[:ch.curr])
}

func (ch *chunk) asRunes() []rune {
	return ch.value[:ch.curr]
}

func (ch *chunk) getAll() []string {
	if !ch.hasNext() {
		return []string{ch.asString()}
	}
	return ch.next.getLeaves(ch.asRunes())
}

func (ch *chunk) hasNext() bool {
	return ch.next != nil
}

func newChunk() *chunk {
	return &chunk{value: make([]rune, 10), curr: 0}
}

// ------------------------------------------------------------------

type alternative struct {
	children []*chunk
	curr     int
}

func (a *alternative) addChunk() *chunk {
	if a.curr >= len(a.children) {
		a.children = append(a.children, make([]*chunk, 10)...)
	}
	a.children[a.curr] = newChunk()
	a.curr++
	return a.children[a.curr-1]
}

func (a *alternative) hasNext() bool {
	return false
}

func (a *alternative) getLeaves(prefix []rune) []string {
	var dfs func(*alternative, []rune)
	alts := make([]string, 50)
	i := 0
	dfs = func(n *alternative, curr []rune) {
		for ci := 0; ci < n.curr; ci++ {
			c := n.children[ci]
			v := append(curr, c.asRunes()...)
			if c.hasNext() {
				dfs(c.next, v)

			} else {
				alts[i] = string(v)
				i++
			}
		}
	}
	dfs(a, prefix)
	return alts[:i]
}
