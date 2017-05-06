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

type RTEdge struct {

	// a radix tree node following this edge (i.e. the top-down direction)
	node *RTNode

	// a token part
	value string

	// position within index
	idx int
}

//
// o --------------> o --- ...
// ?   rte[aabb]     x
//
// o ------------> o ----------------> o --- ...
// ?   rte[aa]  newNode  newEdge[bb]   x
func (rte *RTEdge) split(substr string, idx int) *RTEdge {
	newNode := NewRTNode()

	tmpNode := rte.node
	rte.node = newNode
	newEdge := NewRTEdge(rte.value[len(substr):], rte.idx, tmpNode)
	rte.value = rte.value[:len(substr)]
	rte.idx = idx
	newNode.edges = []*RTEdge{newEdge}

	return newEdge
}

func NewRTEdge(value string, idx int, node *RTNode) *RTEdge {
	return &RTEdge{node: node, value: value, idx: idx}
}

// ----------------------------------------------------------------------------

type RTNode struct {
	edges []*RTEdge
}

func NewRTNode() *RTNode {
	return &RTNode{edges: make([]*RTEdge, 0, defaultEdgesCap)}
}

func (rtn *RTNode) isLeaf() bool {
	return len(rtn.edges) == 0
}

func (rtn *RTNode) addEdge(e *RTEdge) {
	for _, e2 := range rtn.edges {
		if e.value == e2.value {
			return
		}
	}
	rtn.edges = append(rtn.edges, e)
}

// ----------------------------------------------------------------------------

type RadixTree struct {
	root *RTNode
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

func writeTraversingTree(fromNode *RTNode, srch string, idx int) *RTEdge {
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
	newEdge := NewRTEdge(srch, idx, NewRTNode())
	fromNode.addEdge(newEdge)
	return newEdge
}

func traverseTree(fromNode *RTNode, srch string) *RTEdge {
	for _, edge := range fromNode.edges {
		if srch == edge.value {
			return edge

		} else if strings.HasPrefix(srch, edge.value) {
			return traverseTree(edge.node, srch[len(edge.value):])
		}
	}
	return nil
}

func (rt *RadixTree) Add(word string, idx int) *RTEdge {
	return writeTraversingTree(rt.root, word, idx)
}

func (rt *RadixTree) find(word string) *RTEdge {
	return traverseTree(rt.root, word)
}

func (rt *RadixTree) Find(word string) int {
	if srch := rt.find(word); srch != nil {
		return srch.idx
	}
	return -1
}

func NewRadixTree() *RadixTree {
	return &RadixTree{root: NewRTNode()}

}
