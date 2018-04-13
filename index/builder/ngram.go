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

import "strings"

// StdNgramBuffer is used for continuous "circular" inserting
// of tokens and their export as n-grams
type StdNgramBuffer struct {
	begin int
	write int
	Size  int
	data  []string
}

// AddToken add a token to the buffer
func (n *StdNgramBuffer) AddToken(token string) {
	n.write = (n.write + 1) % n.Size
	n.data[n.write] = token
	n.begin = (n.begin + 1) % n.Size
}

// GetValue return current
func (n *StdNgramBuffer) GetValue() []string {
	ans := make([]string, n.Size)
	for i := range n.data {
		ans[i] = n.data[(n.begin+i)%n.Size]
	}
	return ans
}

// IsValid returns true if all the n-grams
// positions are non-empty. This can be used
// to filter out incomplete n-grams from the
// beginning of a sentence
func (n *StdNgramBuffer) IsValid() bool {
	for _, v := range n.data {
		if v == "" {
			return false
		}
	}
	return true
}

// Reset clears out all the values
// and also internal pointers to start
// generating n-grams from scratch.
func (n *StdNgramBuffer) Reset() {
	n.begin = 0
	n.write = -1
	for i := range n.data {
		n.data[i] = ""
	}
}

// Stringer produces a user-friendly overview
// of buffer where tokens are separated by
// spaces. Please note that this works also
// on non-valid tokens. I.e. to be sure,
// IsValid must be called.
func (n *StdNgramBuffer) Stringer() string {
	return strings.Join(n.GetValue(), " ")
}

// NewStdNgramBuffer is a factory function
// which creates a properly initialized
// buffer.
func NewStdNgramBuffer(size int) *StdNgramBuffer {
	return &StdNgramBuffer{
		Size:  size,
		begin: 0,
		write: -1,
		data:  make([]string, size),
	}
}

type DummyNgramBuffer struct {
}

func (n *DummyNgramBuffer) AddToken(token string) {
}

func (n *DummyNgramBuffer) GetValue() []string {
	return []string{}
}

func (n *DummyNgramBuffer) IsValid() bool {
	return false
}

func (n *DummyNgramBuffer) Reset() {
}

func (n *DummyNgramBuffer) Stringer() string {
	return ""
}
