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
	node := NewRTNode()
	assert.Equal(t, 0, len(node.edges))
	edge := NewRTEdge("foo", NewRTNode())
	node.addEdge(edge)
	assert.Equal(t, 1, len(node.edges))
	assert.Equal(t, edge, node.edges[0])
}

func TestRTEdgeSplit(t *testing.T) {
	node1 := NewRTNode()
	node2 := NewRTNode()
	edge12 := NewRTEdge("sunflower", node2)
	node1.addEdge(edge12)

	node3 := NewRTNode()
	edge23 := NewRTEdge("foo", node3)
	node2.addEdge(edge23)

	edge12a := edge12.split("sun")

	assert.Equal(t, node1.edges[0], edge12)
	assert.Equal(t, edge12.node.edges[0], edge12a)
	assert.Equal(t, edge12a.node, node2)
	assert.Equal(t, node2.edges[0], edge23)
	assert.Equal(t, edge23.node, node3)

}
