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
	"bufio"
	"github.com/tomachalek/gloomy/index/builder/files"
	"github.com/tomachalek/gloomy/index/gconf"
	"github.com/tomachalek/vertigo"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"io"
	"regexp"
	"strings"
)

const (
	channelChunkSize = 250000 // changing the value affects performance (TODO how much?)
)

type simpleTokenizer struct {
	lineRegexp *regexp.Regexp
	charset    *charmap.Charmap
}

func (st *simpleTokenizer) parseLine(s string) []string {
	tmp := strings.TrimSpace(importString(s, st.charset))
	return st.lineRegexp.Split(tmp, -1)
}

// ParseSource parses a text provided via a specified reader object
// (typically a file) and charset.
func (st *simpleTokenizer) parseSource(source io.Reader, lproc vertigo.LineProcessor) error {
	brd := bufio.NewScanner(source)

	ch := make(chan []interface{})
	chunk := make([]interface{}, channelChunkSize)
	go func() {
		i := 0
		for brd.Scan() {
			line := st.parseLine(brd.Text())
			for _, token := range line {
				if token != "" {
					chunk[i] = &vertigo.Token{Word: token}
					i++
					if i == channelChunkSize {
						i = 0
						ch <- chunk
						chunk = make([]interface{}, channelChunkSize)
					}
				}
			}
		}
		if i > 0 {
			ch <- chunk[:i]
		}
		close(ch)
	}()

	for tokens := range ch {
		for _, token := range tokens {
			switch token.(type) {
			case *vertigo.Token:
				tk := token.(*vertigo.Token)
				lproc.ProcToken(tk)
			}
		}
	}
	return nil
}

func importString(s string, ch *charmap.Charmap) string {
	if ch == nil { // we assume utf-8 here (default Gloomy encoding)
		return strings.ToLower(s)
	}
	ans, _, _ := transform.String(ch.NewDecoder(), s)
	return strings.ToLower(ans)
}

func newSimpleTokenizer(charsetName string) *simpleTokenizer {
	charset, err := vertigo.GetCharmapByName(charsetName)
	if err != nil {
		panic(err)
	}
	cr := regexp.MustCompile("[,\\.\\s;\\?\\!:]+")
	return &simpleTokenizer{
		lineRegexp: cr,
		charset:    charset,
	}
}

// ParseFile parses a file described (path + encoding) in a provided configuration file
// passing the data to LineProcessor.
func ParseFile(conf *gconf.IndexBuilderConf, lproc vertigo.LineProcessor) error {
	st := newSimpleTokenizer(conf.Encoding)
	rd, err := files.NewReader(conf.VerticalFilePath)
	if err == nil {
		return st.parseSource(rd, lproc)
	}
	return err
}
