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

package wdict

import (
	"strings"
)

const (
	defaultEdgesCap = 30
)

// RTEdge represents a Radix Tree edge
// which holds the actual string infix
type RTEdge struct {

	// a radix tree node following this edge (i.e. the top-down direction)
	node *rtNode

	// a token part
	value string

	// position within index
	idx int
}

// split splits an existing edge into two edges
// with a new node between them
//
// o --------------> o --- ...
// ?   rte[aabb]     x
//
// o ------------> o ----------------> o --- ...
// ?   rte[aa]  newNode  newEdge[bb]   x
//
func (rte *RTEdge) split(substr string, idx int) *RTEdge {
	newNode := newRTNode()

	tmpNode := rte.node
	rte.node = newNode
	newEdge := newRTEdge(rte.value[len(substr):], rte.idx, tmpNode)
	rte.value = rte.value[:len(substr)]
	rte.idx = idx
	newNode.edges = []*RTEdge{newEdge}

	return newEdge
}

func newRTEdge(value string, idx int, node *rtNode) *RTEdge {
	return &RTEdge{node: node, value: value, idx: idx}
}

// ----------------------------------------------------------------------------

type rtNode struct {
	edges []*RTEdge
}

func newRTNode() *rtNode {
	return &rtNode{edges: make([]*RTEdge, 0, defaultEdgesCap)}
}

func (rtn *rtNode) isLeaf() bool {
	return len(rtn.edges) == 0
}

func (rtn *rtNode) addEdge(e *RTEdge) {
	for _, e2 := range rtn.edges {
		if e.value == e2.value {
			return
		}
	}
	rtn.edges = append(rtn.edges, e)
}

// ----------------------------------------------------------------------------

// RadixTree is a simple implementation
// of Radix Tree data structure for searching
// strings by prefixes
// (https://en.wikipedia.org/wiki/Radix_tree)
type RadixTree struct {
	root *rtNode
}

func min(v1 int, v2 int) int {
	if v1 < v2 {
		return v1
	}
	return v2
}

func commonPrefixLen(s1 string, s2 string) int {
	var i int
	for i = 0; i < min(len(s1), len(s2)); i++ {
		if s1[i] != s2[i] {
			return i
		}
	}
	return i
}

// writeTraversingTree writes a new value to the tree by
// traversing it to the right location and by performing
// a required modication
func writeTraversingTree(fromNode *rtNode, srch string, idx int) *RTEdge {
	for _, edge := range fromNode.edges {
		if srch == edge.value {
			return edge

		} else if strings.HasPrefix(srch, edge.value) {
			return writeTraversingTree(edge.node, srch[len(edge.value):], idx)

		} else if strings.HasPrefix(edge.value, srch) {
			return edge.split(srch, idx)
		}
		prefixLen := commonPrefixLen(srch, edge.value)
		if prefixLen > 0 {
			edge.split(srch[:prefixLen], idx)
			edge.idx = -1
			return writeTraversingTree(edge.node, srch[prefixLen:], idx)
		}
	}
	newEdge := newRTEdge(srch, idx, newRTNode())
	fromNode.addEdge(newEdge)
	return newEdge
}

func traverseTree(fromNode *rtNode, srch string) *RTEdge {
	for _, edge := range fromNode.edges {
		if srch == edge.value {
			return edge

		} else if strings.HasPrefix(srch, edge.value) {
			return traverseTree(edge.node, srch[len(edge.value):])
		}
	}
	return nil
}

func collectWords(fromNode *rtNode, partWord string, ans []string) []string {
	for _, edge := range fromNode.edges {
		if edge.idx > -1 {
			ans = append(ans, partWord+edge.value)
		}
		ans = collectWords(edge.node, partWord+edge.value, ans)
	}
	return ans
}

func collectIndices(fromNode *rtNode, ans []int) []int {
	for _, edge := range fromNode.edges {
		if edge.idx > -1 {
			ans = append(ans, edge.idx)
		}
		ans = collectIndices(edge.node, ans)
	}
	return ans
}

func findWord(fromNode *rtNode, partWord string) *RTEdge {
	for _, edge := range fromNode.edges {
		if strings.HasPrefix(partWord, edge.value) && len(partWord) > len(edge.value) {
			return findWord(edge.node, partWord[len(edge.value):])

		} else if strings.HasPrefix(edge.value, partWord) {
			return edge
		}
	}
	return nil
}

// FindByPrefix finds all the matching strings with
// prefixes equal to 'prefix'.
func (rt *RadixTree) FindByPrefix(prefix string) []string {
	ans := make([]string, 0, 10)
	srchEdge := findWord(rt.root, prefix)
	if srchEdge != nil {
		if srchEdge.idx > -1 {
			ans = append(ans, prefix)
		}
		return collectWords(srchEdge.node, prefix, ans)
	}
	return []string{}
}

// FindIndicesByPrefix finds all the matching indices
// of strings with prefixes equal to 'prefix'.
func (rt *RadixTree) FindIndicesByPrefix(prefix string) []int {
	ans := make([]int, 0, 10)
	srchEdge := findWord(rt.root, prefix)
	if srchEdge != nil {
		if srchEdge.idx > -1 {
			ans = append(ans, srchEdge.idx)
		}
		return collectIndices(srchEdge.node, ans)
	}
	return []int{}
}

// Add adds a word and its dictionary index to the
// tree.
func (rt *RadixTree) Add(word string, idx int) *RTEdge {
	return writeTraversingTree(rt.root, word, idx)
}

func (rt *RadixTree) find(word string) *RTEdge {
	return traverseTree(rt.root, word)
}

// Find finds a matching word (exact) and returns
// its stored index.
func (rt *RadixTree) Find(word string) int {
	if srch := rt.find(word); srch != nil {
		return srch.idx
	}
	return -1
}

// NewRadixTree creates and returns an instance
// of an empty Radix Tree.
func NewRadixTree() *RadixTree {
	return &RadixTree{root: newRTNode()}
}
