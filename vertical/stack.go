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

package vertical

import (
	"log"
)

type ElmParser interface {
	Begin(value *VerticalMetaLine)
	End(name string) *VerticalMetaLine
	GetAttrs() map[string]string
	Size() int
}

type stackItem struct {
	value *VerticalMetaLine
	prev  *stackItem
}

// ---------------------------------------------------------

// Stack represents a data structure used to keep
// vertical file (xml-like) metadata. It is implemented
// as a simple linked list
type Stack struct {
	last *stackItem
}

// NewStack creates a new Stack instance
func NewStack() *Stack {
	return &Stack{}
}

// Push adds an item at the beginning
func (s *Stack) Begin(value *VerticalMetaLine) {
	item := &stackItem{value: value, prev: s.last}
	s.last = item
}

// Pop takes the first element
func (s *Stack) End(name string) *VerticalMetaLine {
	if name != s.last.value.Name {
		log.Printf("Tag nesting problem. Expected: %s, found %s", s.last.value.Name, name)
	}
	item := s.last
	s.last = item.prev
	return item.value
}

// Size returns a size of the stack
func (s *Stack) Size() int {
	size := 0
	item := s.last
	for {
		if item != nil {
			size++
			item = item.prev

		} else {
			break
		}
	}
	return size
}

// GetAttrs returns all the actual structural attributes
// and their values found on stack.
// Elements are encoded as follows:
// [struct_name].[attr_name]=[value]
// (e.g. doc.author="Isaac Asimov")
func (s *Stack) GetAttrs() map[string]string {
	ans := make(map[string]string)
	curr := s.last
	for curr != nil {
		for k, v := range curr.value.Attrs {
			ans[curr.value.Name+"."+k] = v
		}
		curr = curr.prev
	}
	return ans
}
