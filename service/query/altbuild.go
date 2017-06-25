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

import "fmt"

type atnstate struct {
	children []*atnstate
	value    rune
}

func (a *atnstate) String() string {
	if a.isRune() {
		return fmt.Sprintf("RUNE[v: %c, num children: %d]", a.value, len(a.children))
	}
	return fmt.Sprintf("STATE[num children: %d]", len(a.children))
}

func (a *atnstate) asString() string {
	return string(a.value)
}

func (a *atnstate) getLast() *atnstate {
	// we expect every atnstate to be properly merged into a single final state
	curr := a
	for !curr.isLeaf() {
		curr = curr.children[0]
	}
	return curr
}

func (a *atnstate) addState() *atnstate {
	return a.appendState(newState())
}

func (a *atnstate) appendState(a2 *atnstate) *atnstate {
	if a.value != '\u0000' && len(a.children) > 0 {
		panic(fmt.Sprintf("Rune-like state cannot have multiple outgoing states. Value: %c", a.value))
	}
	a.children = append(a.children, a2)
	return a2
}

func (a *atnstate) addRune(value rune) *atnstate {
	return a.appendState(&atnstate{
		children: make([]*atnstate, 0, 1),
		value:    value,
	})
}

func (a *atnstate) isLeaf() bool {
	return len(a.children) == 0
}

func (a *atnstate) isRune() bool {
	return a.value != '\u0000'
}

func (a *atnstate) hasChild(a2 *atnstate) bool {
	for _, c := range a.children {
		if c == a2 {
			return true
		}
	}
	return false
}

func (a *atnstate) removeChild(child *atnstate) {
	for i, c := range a.children {
		if c == child {
			fmt.Println("REMOVE CHILD ", i)
			copy(a.children[i:], a.children[i+1:])
			a.children[len(a.children)-1] = nil
			a.children = a.children[:len(a.children)-1]
			break
		}
	}
}

func (a *atnstate) mergeAlternatives() *atnstate {
	var dfs func(*atnstate)
	leaves := make([]*atnstate, 0, 10)
	dfs = func(n *atnstate) {
		if n.isLeaf() {
			leaves = append(leaves, n)
		}
		for _, c := range n.children {
			dfs(c)
		}
	}
	dfs(a)
	newItem := &atnstate{}
	for _, item := range leaves {
		item.appendState(newItem)
	}
	return newItem
}

func (a *atnstate) getAll() []string {
	return a.getAlternatives([]rune{})
}

func (a *atnstate) getAlternatives(prefix []rune) []string {
	var dfs func(*atnstate, []rune)
	alts := make([]string, 0, 50)

	dfs = func(n *atnstate, prev []rune) {
		var v []rune
		if n.value != '\u0000' {
			v = append(prev, n.value)

		} else {
			v = prev
		}
		if n.isLeaf() {
			alts = append(alts, string(v))
		}
		for _, c := range n.children {
			dfs(c, v)
		}
	}
	dfs(a, prefix)
	return alts
}

func newState() *atnstate {
	return &atnstate{
		children: make([]*atnstate, 0, 10),
		value:    '\u0000',
	}
}

func newRune(value rune) *atnstate {
	return &atnstate{
		children: make([]*atnstate, 0, 1),
		value:    value,
	}
}

// ------------------------------------------------------------------

type stackItem struct {
	value *atnstate
	prev  *stackItem
}

type altStack struct {
	last *stackItem
}

// newStack creates a new Stack instance
func newAltStack() *altStack {
	return &altStack{}
}

func (s *altStack) isEmpty() bool {
	return s.last == nil
}

// Push adds an item at the beginning
func (s *altStack) Push(value *atnstate) {
	item := &stackItem{value: value, prev: s.last}
	s.last = item
}

func (s *altStack) PeekOrCreate() *atnstate {
	if s.last == nil {
		s.Push(&atnstate{})
	}
	return s.Peek()
}

// Pop takes the first element
func (s *altStack) Pop() *atnstate {
	item := s.last
	s.last = item.prev
	return item.value
}

func (s *altStack) Peek() *atnstate {
	if s.last != nil {
		return s.last.value
	}
	return nil
}
