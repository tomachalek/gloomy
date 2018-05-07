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

package builder

import (
	"fmt"

	"github.com/tomachalek/gloomy/index/column"
)

func ngramsCmp(n1 []string, n2 []string) int {
	if len(n1) != len(n2) {
		panic(fmt.Sprintf("Cannot compare ngrams of different sizes (%d vs. %d)", len(n1), len(n2)))
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

// ------------------------------------------------

type ngramNode struct {
	left     *ngramNode
	right    *ngramNode
	ngram    []string
	count    int
	metadata *column.MetadataList
}

func (n *ngramNode) GetCount() int {
	return n.count
}

func (n *ngramNode) GetNgram() []string {
	return n.ngram
}

func newNgramNode(ngram []string, metadata column.Metadata) *ngramNode {
	meta := &column.MetadataList{}
	meta.Add(metadata)
	return &ngramNode{
		ngram:    ngram,
		count:    1,
		metadata: meta,
	}
}

type RAMNgramList struct {
	root     *ngramNode
	numNodes int
}

func dfsWalkthruRecursive(node *ngramNode, fn func(n *NgramRecord)) {
	if node.left != nil {
		dfsWalkthruRecursive(node.left, fn)
	}
	fn(&NgramRecord{Ngram: node.ngram, Count: node.count, Metadata: node.metadata})
	if node.right != nil {
		dfsWalkthruRecursive(node.right, fn)
	}
}

func (n *RAMNgramList) ForEach(fn func(n *NgramRecord)) {
	if n.root != nil {
		dfsWalkthruRecursive(n.root, fn)
	}
}

func (n *RAMNgramList) Size() int {
	return n.numNodes
}

func (n *RAMNgramList) Add(ngram []string, metadata column.Metadata) {
	if n.root == nil {
		n.root = newNgramNode(ngram, metadata)
		n.numNodes = 1

	} else {
		item := n.root
		for item != nil {
			switch ngramsCmp(ngram, item.ngram) {
			case -1:
				if item.left != nil {
					item = item.left

				} else {
					item.left = newNgramNode(ngram, metadata)
					n.numNodes++
					item = nil // stop the iteration
				}
			case 1:
				if item.right != nil {
					item = item.right

				} else {
					item.right = newNgramNode(ngram, metadata)
					n.numNodes++
					item = nil // stop the iteration
				}
			case 0:
				item.metadata.Add(metadata)
				item.count++
				item = nil // stop the iteration
			}
		}
	}

}
