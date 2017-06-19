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
}

func (ch *chunk) addAlternative() *alternative {
	if ch.next != nil {
		panic("Cannot add alternative - already present")
	}
	ch.next = &alternative{children: make([]*chunk, 0, 10)}
	return ch.next
}

func (ch *chunk) addRune(v rune) {
	if cap(ch.value) == len(ch.value) {
		ch.value = append(ch.value, make([]rune, 1, 10)...)

	} else {
		ch.value = ch.value[:len(ch.value)+1]
	}
	ch.value[len(ch.value)-1] = v
}

func (ch *chunk) asString() string {
	return string(ch.value)
}

func (ch *chunk) asRunes() []rune {
	return ch.value
}

func (ch *chunk) getAll() []string {
	if !ch.hasNext() {
		return []string{ch.asString()}
	}
	return ch.next.getAlternatives(ch.asRunes())
}

func (ch *chunk) hasNext() bool {
	return ch.next != nil
}

func newChunk() *chunk {
	return &chunk{value: make([]rune, 0, 10)}
}

// ------------------------------------------------------------------

type alternative struct {
	children []*chunk
}

func (a *alternative) addChunk() *chunk {
	if cap(a.children) == len(a.children) {
		a.children = append(a.children, make([]*chunk, 1, 10)...)

	} else {
		a.children = a.children[:len(a.children)+1]
	}
	nc := newChunk()
	a.children[len(a.children)-1] = nc
	return nc
}

func (a *alternative) isLeave() bool {
	for _, v := range a.children {
		if v.hasNext() {
			return false
		}
	}
	return true
}

func (a *alternative) getAlternatives(prefix []rune) []string {
	var dfs func(*alternative, []rune)
	alts := make([]string, 50)
	i := 0
	dfs = func(n *alternative, curr []rune) {
		for _, c := range n.children {
			v := append(curr, c.asRunes()...)
			if c.hasNext() {
				dfs(c.next, v)

			} else {
				if i >= len(alts) {
					alts = append(alts, make([]string, 10)...)
				}
				alts[i] = string(v)
				i++
			}
		}
	}
	dfs(a, prefix)
	return alts[:i]
}

func (a *alternative) getLeaves() []*alternative {
	var dfs func(*alternative)
	alts := make([]*alternative, 50)
	i := 0
	dfs = func(n *alternative) {
		if n.isLeave() {
			if i >= len(alts) {
				alts = append(alts, make([]*alternative, 10)...)
			}
			alts[i] = n
			i++

		} else {
			for _, c := range n.children {
				if c.hasNext() {
					dfs(c.next)

				}
			}
		}
	}
	dfs(a)
	return alts[:i]
}
