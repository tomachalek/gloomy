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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRTNodeAddEdge(t *testing.T) {
	node := newRTNode()
	assert.Equal(t, 0, len(node.edges))
	edge := newRTEdge("foo", 20, newRTNode())
	node.addEdge(edge)
	assert.Equal(t, 1, len(node.edges))
	assert.Equal(t, edge, node.edges[0])
}

func TestCommonPrefixLen(t *testing.T) {
	assert.Equal(t, 4, commonPrefixLen("atom", "atomius"))
	assert.Equal(t, 4, commonPrefixLen("atomax", "atomius"))
	assert.Equal(t, 2, commonPrefixLen("at", "at"))
	assert.Equal(t, 0, commonPrefixLen("bat", "at"))
	assert.Equal(t, 1, commonPrefixLen("bat", "bet"))
	assert.Equal(t, 0, commonPrefixLen("", ""))

}

//
//  initial state:
//   o ------------------> o ------------> o
//   n1  e12[sunflower]    n2  e23[foo]    n3
//
//   after the split:
//   o ------------> o ----------------> o -------------> o
//   n1  e12[sun]    nX  e12a[flower]    n2   e23[foo]    n3
//
func TestRTEdgeSplit(t *testing.T) {
	node1 := newRTNode()
	node2 := newRTNode()
	edge12 := newRTEdge("sunflower", 37, node2)
	node1.addEdge(edge12)

	node3 := newRTNode()
	edge23 := newRTEdge("foo", 52, node3)
	node2.addEdge(edge23)

	edge12a := edge12.split("sun", 14)

	assert.Equal(t, node1.edges[0], edge12)
	assert.Equal(t, "sun", edge12.value)
	assert.Equal(t, 14, edge12.idx)

	assert.Equal(t, edge12.node.edges[0], edge12a)
	assert.Equal(t, "flower", edge12a.value)
	assert.Equal(t, 37, edge12a.idx)
	assert.Equal(t, edge12a.node, node2)

	assert.Equal(t, node2.edges[0], edge23)
	assert.Equal(t, edge23.node, node3)
	assert.Equal(t, "foo", edge23.value)
	assert.Equal(t, 52, edge23.idx)
}

func TestAddSubstringToExisting(t *testing.T) {
	rt := NewRadixTree()
	rt.Add("sunflower", 11)
	assert.Equal(t, "sunflower", rt.root.edges[0].value)
	assert.Equal(t, 11, rt.root.edges[0].idx)
	assert.Equal(t, 1, len(rt.root.edges))

	edge := rt.Add("sun", 12)

	assert.Equal(t, 1, len(rt.root.edges))
	assert.Equal(t, edge, rt.root.edges[0].node.edges[0])
	assert.Equal(t, 11, rt.root.edges[0].node.edges[0].idx)
	assert.Equal(t, 12, rt.root.edges[0].idx)
	assert.Equal(t, "flower", edge.value)
}

func TestAddSuperStringToExisting(t *testing.T) {
	rt := NewRadixTree()
	rt.Add("sun", 11)
	assert.Equal(t, "sun", rt.root.edges[0].value)
	assert.Equal(t, 11, rt.root.edges[0].idx)
	rt.Add("sunflower", 12)
	assert.Equal(t, "flower", rt.root.edges[0].node.edges[0].value)
	assert.Equal(t, 12, rt.root.edges[0].node.edges[0].idx)
}

func TestAddTwoWithCommonPrefix(t *testing.T) {
	rt := NewRadixTree()
	rt.Add("romane", 12)
	assert.Equal(t, 12, rt.root.edges[0].idx)
	rt.Add("romanus", 13)
	assert.Equal(t, -1, rt.root.edges[0].idx) // common prefix edge does not refer to any actual index
	assert.Equal(t, 1, len(rt.root.edges))
	assert.Equal(t, "roman", rt.root.edges[0].value)
	tmp := rt.root.edges[0].node
	assert.Equal(t, 2, len(tmp.edges))
	assert.Equal(t, "e", tmp.edges[0].value)
	assert.Equal(t, 12, tmp.edges[0].idx)
	assert.Equal(t, "us", tmp.edges[1].value)
	assert.Equal(t, 13, tmp.edges[1].idx)
}

func createTestingTree() *RadixTree {
	rt := NewRadixTree()
	rt.Add("romane", 11)
	rt.Add("romanus", 12)
	rt.Add("romulus", 13)
	rt.Add("rubens", 14)
	rt.Add("ruber", 15)
	rt.Add("rubicon", 16)
	rt.Add("voltron", 17)
	return rt
}

func TestSearch(t *testing.T) {
	rt := createTestingTree()
	var srch *RTEdge

	srch = rt.find("romane")
	assert.Equal(t, "e", srch.value)
	assert.Equal(t, 11, srch.idx)

	srch = rt.find("romanus")
	assert.Equal(t, "us", srch.value)
	assert.Equal(t, 12, srch.idx)

	srch = rt.find("romulus")
	assert.Equal(t, "ulus", srch.value)
	assert.Equal(t, 13, srch.idx)

	srch = rt.find("rubens")
	assert.Equal(t, "ns", srch.value)
	assert.Equal(t, 14, srch.idx)

	srch = rt.find("ruber")
	assert.Equal(t, "r", srch.value)
	assert.Equal(t, 15, srch.idx)

	srch = rt.find("rubicon")
	assert.Equal(t, "icon", srch.value)
	assert.Equal(t, 16, srch.idx)

	srch = rt.find("rub")
	assert.Equal(t, -1, srch.idx)

	srch = rt.find("xen")
	assert.Nil(t, srch)

	srch = rt.find("rombom")
	assert.Nil(t, srch)
}

func TestApiFound(t *testing.T) {
	rt := createTestingTree()
	var idx int

	idx = rt.Find("romane")
	assert.Equal(t, 11, idx)

	idx = rt.Find("xen")
	assert.Equal(t, -1, idx)
}

func TestFindByPrefix(t *testing.T) {
	rt := createTestingTree()
	ans := rt.FindByPrefix("rom")
	assert.Equal(t, "romane", ans[0])
	assert.Equal(t, "romanus", ans[1])
	assert.Equal(t, "romulus", ans[2])
	assert.Equal(t, 3, len(ans))
}

func TestFindWholeWordSingleEdge(t *testing.T) {
	rt := createTestingTree()
	ans := rt.FindByPrefix("voltron")
	assert.Equal(t, "voltron", ans[0])
}

func TestFindByPrefixNonExisting(t *testing.T) {
	rt := createTestingTree()
	ans := rt.FindByPrefix("romix")
	assert.Equal(t, 0, len(ans))
}

func TestFindIndicesByPrefixNonExisting(t *testing.T) {
	rt := createTestingTree()
	ans := rt.FindIndicesByPrefix("rom")
	assert.Equal(t, 11, ans[0])
	assert.Equal(t, 12, ans[1])
	assert.Equal(t, 13, ans[2])
	assert.Equal(t, 3, len(ans))
}
