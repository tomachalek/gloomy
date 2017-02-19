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

type StackItem struct {
	value *VerticalMetaLine
	prev  *StackItem
}

type Stack struct {
	last *StackItem
}

func NewStack() *Stack {
	return &Stack{}
}

func (s *Stack) Push(value *VerticalMetaLine) {
	item := &StackItem{value: value, prev: s.last}
	s.last = item
}

func (s *Stack) Pop() *VerticalMetaLine {
	item := s.last
	s.last = item.prev
	return item.value
}

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
