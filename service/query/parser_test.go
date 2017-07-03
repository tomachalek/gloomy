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
	alts := p.GetAllPrefixes()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(alts))
	assert.Equal(t, s, alts[0])
}

func TestCharEnum(t *testing.T) {
	p := NewParser()
	err := p.Parse("te[dxa]Z")
	assert.Nil(t, err)
	alts := p.GetAllPrefixes()
	assert.Equal(t, 3, len(alts))
	assert.Equal(t, "tedZ", alts[0])
	assert.Equal(t, "texZ", alts[1])
	assert.Equal(t, "teaZ", alts[2])
}

func TestIncorrectCharEnum(t *testing.T) {
	p := NewParser()
	err := p.Parse("[hxXH")
	assert.Error(t, err)
}

func TestCharEnumCannotInclParentheses(t *testing.T) {
	p := NewParser()
	err := p.Parse("[hx(foo)H]")
	assert.Error(t, err)
}

func TestParentheses(t *testing.T) {
	p := NewParser()
	err := p.Parse("(foo)")
	assert.Nil(t, err)
	alts := p.GetAllPrefixes()
	assert.Equal(t, 1, len(alts))
	assert.Equal(t, "foo", alts[0])
}

func TestParenthesesMissingLeft(t *testing.T) {
	p := NewParser()
	err := p.Parse("foo)")
	assert.Error(t, err)
}

func TestAlternatives(t *testing.T) {
	p := NewParser()
	err := p.Parse("(foo)|(bar)|(baz)")
	assert.Nil(t, err)
	alts := p.GetAllPrefixes()
	assert.Equal(t, 3, len(alts))
	assert.Equal(t, "foo", alts[0])
	assert.Equal(t, "bar", alts[1])
	assert.Equal(t, "baz", alts[2])
}

func TestAlternativesPriority(t *testing.T) {
	p := NewParser()
	err := p.Parse("foo|bar")
	assert.Nil(t, err)
	alts := p.GetAllPrefixes()
	assert.Equal(t, 2, len(alts))
	assert.Equal(t, "fooar", alts[0])
	assert.Equal(t, "fobar", alts[1])
}

// Combined stuff

func TestAlternatives2(t *testing.T) {
	p := NewParser()
	err := p.Parse("(foo)|([bB]ar)")
	assert.Nil(t, err)
	alts := p.GetAllPrefixes()
	assert.Equal(t, 3, len(alts))
	assert.Equal(t, "foo", alts[0])
	assert.Equal(t, "bar", alts[1])
	assert.Equal(t, "Bar", alts[2])
}

func TestAlternatives3(t *testing.T) {
	p := NewParser()
	err := p.Parse("abc?d")
	assert.Nil(t, err)
	alts := p.GetAllPrefixes()
	assert.Equal(t, 2, len(alts))
	assert.Equal(t, "abcd", alts[0])
	assert.Equal(t, "abd", alts[1])
}

func TestAlternatives4(t *testing.T) {
	p := NewParser()
	err := p.Parse("me(to)?dic")
	assert.Nil(t, err)
	alts := p.GetAllPrefixes()
	assert.Equal(t, 2, len(alts))
	assert.Equal(t, "metodic", alts[0])
	assert.Equal(t, "medic", alts[1])
}

func TestAlternatives5(t *testing.T) {
	p := NewParser()
	err := p.Parse("me(tada)?[Tt]a")
	assert.Nil(t, err)
	alts := p.GetAllPrefixes()
	assert.Equal(t, 4, len(alts))
	assert.Equal(t, "metadaTa", alts[0])
	assert.Equal(t, "metadata", alts[1])
	assert.Equal(t, "meTa", alts[2])
	assert.Equal(t, "meta", alts[3])
}

func TestAlternatives6(t *testing.T) {
	p := NewParser()
	err := p.Parse("me(tad[aA]x)?")
	assert.Nil(t, err)
	alts := p.GetAllPrefixes()
	assert.Equal(t, 3, len(alts))
	assert.Equal(t, "metadax", alts[0])
	assert.Equal(t, "metadAx", alts[1])
	assert.Equal(t, "me", alts[2])
}

func TestPlaceholder(t *testing.T) {
	p := NewParser()
	err := p.Parse("foo.+z")
	alts := p.GetAllPrefixes()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(alts))
	assert.Equal(t, "foo*", alts[0])
}

func TestPlaceholder2(t *testing.T) {
	p := NewParser()
	err := p.Parse("foo.*z")
	alts := p.GetAllPrefixes()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(alts))
	assert.Equal(t, "foo*", alts[0])
	assert.Equal(t, "fooz", alts[1])
}
