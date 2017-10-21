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

package tokenizer

import (
	"github.com/stretchr/testify/assert"
	"github.com/tomachalek/vertigo"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	loremData = []string{
		"Lorem", "ipsum", "dolor", "sit", "amet", "consectetur", "adipiscing", "elit", "sed", "do",
		"eiusmod", "tempor", "incididunt", "ut", "labore", "et", "dolore", "magna", "aliqua", "Ut",
		"enim", "ad", "minim", "veniam", "quis", "nostrud", "exercitation", "ullamco", "laboris", "nisi",
		"ut", "aliquip", "ex", "ea", "commodo", "consequat", "Duis", "aute", "irure", "dolor",
		"in", "reprehenderit", "in", "voluptate", "velit", "esse", "cillum", "dolore", "eu", "fugiat",
		"nulla", "pariatur", "Excepteur", "sint", "occaecat", "cupidatat", "non", "proident", "sunt", "in",
		"culpa", "qui", "officia", "deserunt", "mollit", "anim", "id", "est", "laborum",
	}
)

type TestingProc struct {
	i      int
	tokens []string
}

func (tp *TestingProc) ProcToken(token *vertigo.Token) {
	tp.tokens[tp.i] = token.Word
	tp.i++
}

func (tp *TestingProc) ProcStruct(strc *vertigo.Structure) {

}

func (tp *TestingProc) ProcStructClose(strc *vertigo.StructureClose) {

}

func TestParseText(t *testing.T) {
	loremPath := filepath.Join(filepath.Dir(os.Getenv("GOPATH")), "go/src/github.com/tomachalek/gloomy/testdata/loremipsum.txt")
	conf := &vertigo.ParserConf{
		InputFilePath: loremPath,
		Encoding:      "utf-8",
	}
	testingProc := &TestingProc{
		tokens: make([]string, 69),
	}
	ParseFile(conf, testingProc)
	for i, s := range testingProc.tokens {
		assert.Equal(t, strings.ToLower(loremData[i]), s)
	}
}

func TestParseEmpty(t *testing.T) {
	s := ""
	r := strings.NewReader(s)
	st, _ := newSimpleTokenizer("utf-8")
	testingProc := &TestingProc{
		tokens: make([]string, 5),
	}
	st.parseSource(r, testingProc)
	assert.Equal(t, testingProc.tokens[0], "")
	assert.Equal(t, testingProc.i, 0)
}

func TestNonUTF8Encoding(t *testing.T) {
	textPath := filepath.Join(filepath.Dir(os.Getenv("GOPATH")), "go/src/github.com/tomachalek/gloomy/testdata/cs-win1250.txt")
	r, _ := os.Open(textPath)
	st, _ := newSimpleTokenizer("windows-1250")
	testingProc := &TestingProc{
		tokens: make([]string, 5),
	}
	st.parseSource(r, testingProc)
	tst := []string{"žluťoučký", "kůň", "úpěl", "ďábelské", "ódy"}
	for i, s := range tst {
		assert.Equal(t, s, testingProc.tokens[i])
	}
}

func TestInvalidEncoding(t *testing.T) {
	st, err := newSimpleTokenizer("foo")
	assert.Nil(t, st)
	assert.Error(t, err)
}
