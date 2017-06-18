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

package query

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBasicString(t *testing.T) {
	p := NewParser()
	s := "žluťoučký"
	err := p.Parse(s)
	assert.Nil(t, err)
}

func TestCharEnum(t *testing.T) {
	p := NewParser()
	err := p.Parse("[hxXH]")
	assert.Nil(t, err)
}

func TestIncorrectCharEnum(t *testing.T) {
	p := NewParser()
	err := p.Parse("[hxXH")
	assert.Error(t, err)
}

func TestParentheses(t *testing.T) {
	p := NewParser()
	err := p.Parse("(foo)")
	assert.Nil(t, err)
}

func TestParenthesesMissingLeft(t *testing.T) {
	p := NewParser()
	err := p.Parse("foo)")
	assert.Error(t, err)
}

func TestAlternatives(t *testing.T) {
	p := NewParser()
	err := p.Parse("foo|bar")
	assert.Nil(t, err)
}

// Combined stuff

func TestAlternatives2(t *testing.T) {
	p := NewParser()
	err := p.Parse("foo|[bB]ar")
	assert.Nil(t, err)
}
