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
	left  *NgramNode
	right *NgramNode
	ngram []string
}

type NgramList struct {
	root *NgramNode
}

func dfsWalkthruRecursive(node *NgramNode, fn func(n *NgramNode)) {
	if node.left != nil {
		dfsWalkthruRecursive(node.left, fn)
	}
	fn(node)
	if node.right != nil {
		dfsWalkthruRecursive(node.right, fn)
	}
}

func (n *NgramList) DFSWalkthru(fn func(n *NgramNode)) {
	dfsWalkthruRecursive(n.root, fn)
}

func (n *NgramList) Add(ngram []string) {
	if n.root == nil {
		n.root = &NgramNode{ngram: ngram}

	} else {
		item := n.root
		for item != nil {
			if ngramsCmp(ngram, item.ngram) <= 0 {
				if item.left != nil {
					item = item.left

				} else {
					item.left = &NgramNode{ngram: ngram}
					break
				}

			} else {
				if item.right != nil {
					item = item.right

				} else {
					item.right = &NgramNode{ngram: ngram}
					break
				}
			}
		}
	}

}
