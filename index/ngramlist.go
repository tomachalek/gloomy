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

func ngramsCmp(n1 []string, n2 []string) int {
	if len(n1) != len(n2) {
		panic("Cannot compare ngrams of different sizes")
	}
	for i := 0; i < len(n1); i++ {
		if n1[i] > n2[i] {
			return 1

		} else if n1[i] < n2[i] {
			return -1
		}
	}
	return 0
}

type NgramNode struct {
	next  *NgramNode
	ngram []string
}

type SortedNgramList struct {
	firstNode *NgramNode
}

func (n *SortedNgramList) IndexOf(ngram []string) int {
	item := n.firstNode
	idx := 0
	for item != nil {
		if ngramsCmp(ngram, item.ngram) == 0 {
			return idx
		}
		item = item.next
	}
	return -1
}

func (n *SortedNgramList) Get(index int) *NgramNode {
	item := n.firstNode
	for i := 0; item != nil && i <= index; i++ {
		item = item.next
	}
	return item
}

func (n *SortedNgramList) GetRange(index1 int, index2 int) [][]string {
	first := n.Get(index1)
	if index1 >= index2 || first == nil {
		return [][]string{}
	}
	ans := make([][]string, index2-index1)
	item := first
	for i := 0; i < index2-index1; i, item = i+1, item.next {
		if item != nil {
			ans[i] = item.ngram

		} else {
			return ans[:i]
		}
	}
	return ans
}

func (n *SortedNgramList) Add(ngram []string) {
	item := n.firstNode

	if item == nil {
		n.firstNode = &NgramNode{ngram: ngram}

	} else if ngramsCmp(ngram, item.ngram) == -1 {
		n.firstNode = &NgramNode{ngram: ngram, next: item}

	} else {
		for item.next != nil && ngramsCmp(ngram, item.ngram) <= 0 {
			item = item.next
		}
		newNode := &NgramNode{ngram: ngram, next: item.next}
		item.next = newNode
	}
}
