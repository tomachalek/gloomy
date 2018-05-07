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

// Package index represents an n-gram index as a read-only structure
// providing both low level methods for accessing the internal ngram tree
// and higher level methods for searching a specific word.
package index

import "log"

type NgramResultItem struct {
	next     *NgramResultItem
	Ngram    []int
	Count    int
	Metadata []string
}

// NgramSearchResult is a low level result
// representation where n-grams are just
// arrays of integers (i.e. no translation
// to actual words yet).
//
// The result is implemented to behave
// as a kind of iterator (using HasNext(), Next()
// methods) rather than copying all the result
// data into an array.
type NgramSearchResult struct {
	first *NgramResultItem
	curr  *NgramResultItem
	last  *NgramResultItem
	size  int
}

func (nsr *NgramSearchResult) Append(other *NgramSearchResult) {
	if nsr.last != nil {
		nsr.last.next = other.first

	} else {
		nsr.first = other.first
		nsr.curr = nsr.first
	}
	nsr.last = other.last
	nsr.size += other.size
}

// RemoveNext removes the following item to the 'v' one.
// In case v is nil, first item is removed.
// The function call resets iterator to the first item.
func (nsr *NgramSearchResult) RemoveNext(v *NgramResultItem) *NgramResultItem {
	var rmitem *NgramResultItem
	if v == nil { // remove first
		rmitem = nsr.first
		nsr.first = nsr.first.next
		nsr.curr = nsr.first

	} else {
		rmitem = v.next
		v.next = rmitem.next
		if nsr.last == rmitem {
			nsr.last = v
		}
	}
	nsr.size--
	rmitem.next = nil
	return rmitem
}

func (nsr *NgramSearchResult) Filter(fn func(*NgramResultItem) bool) {
	var prev, curr *NgramResultItem
	curr = nsr.first
	for curr != nil {
		if !fn(curr) {
			curr = curr.next
			nsr.RemoveNext(prev)

		} else {
			prev, curr = curr, curr.next
		}
	}
}

// Slice slices internal list preserving items starting
// from leftIdx (including) up to rightIdx (excluding).
// If an actual slice has been performed then true is
// returned, otherwise false is returned. Slice is
// performed only if rightIdx is strictly greater than
// leftIdx.
func (nsr *NgramSearchResult) Slice(leftIdx int, rightIdx int) bool {
	if leftIdx < 0 || rightIdx >= nsr.Size() {
		log.Panicf("Invalid slice arguments (%d, %d)", leftIdx, rightIdx)
	}
	if leftIdx >= rightIdx {
		return false
	}
	curr := nsr.first
	for i := 1; i <= leftIdx; i++ {
		curr = curr.next
	}
	nsr.first = curr
	nsr.curr = curr
	nsr.size = 1
	for i := leftIdx + 1; i < rightIdx; i++ {
		curr = curr.next
		nsr.size++
	}
	nsr.last = curr
	curr.next = nil
	return true
}

// Size returns a size of the result
// (this is an O(1) operation)
func (nsr *NgramSearchResult) Size() int {
	return nsr.size
}

// ResetCursor moves a pointer pointing
// to the current result item back to the
// first result item.
func (nsr *NgramSearchResult) ResetCursor() {
	nsr.curr = nsr.first
}

// HasNext tests whether the result
// has at least one item left (i.e. whether
// it is possible to call Next() and get
// a valid row)
func (nsr *NgramSearchResult) HasNext() bool {
	return nsr.curr != nil
}

// Next returs a following result item.
func (nsr *NgramSearchResult) Next() *NgramResultItem {
	ans := nsr.curr
	if ans == nil {
		return nil
	}
	nsr.curr = nsr.curr.next
	return ans
}

func (nsr *NgramSearchResult) addValue(ngram []int, count int, metadata []string) {
	item := &NgramResultItem{Ngram: ngram, Count: count, Metadata: metadata}
	if nsr.first == nil {
		nsr.first = item
	}
	if nsr.curr != nil {
		nsr.curr.next = item
	}
	nsr.curr = item
	nsr.last = item
	nsr.size++
}
