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
	node  *RTNode
	value string
}

func (rte *RTEdge) split(substr string) *RTEdge {
	newNode := NewRTNode()

	tmpNode := rte.node
	rte.node = newNode
	newEdge := NewRTEdge(rte.value[len(substr):], tmpNode)
	rte.value = rte.value[:len(substr)]
	newNode.edges = []*RTEdge{newEdge}

	return newEdge
}

func NewRTEdge(value string, node *RTNode) *RTEdge {
	return &RTEdge{node: node, value: value}
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

func traverseTree(fromNode *RTNode, srch string) *RTEdge {
	for _, edge := range fromNode.edges {
		if srch == edge.value {
			return edge

		} else if strings.HasPrefix(srch, edge.value) {
			return traverseTree(edge.node, srch[len(edge.value):])

		} else if strings.HasPrefix(edge.value, srch) {
			return edge.split(srch)

		} else if fromNode.isLeaf() {
		}
		return edge
	}
	newEdge := NewRTEdge(srch, NewRTNode())
	fromNode.addEdge(newEdge)
	return newEdge
}

func (rt *RadixTree) Add(word string) *RTEdge {
	return traverseTree(rt.root, word)
}

func NewRadixTree() *RadixTree {
	return &RadixTree{root: NewRTNode()}
}
