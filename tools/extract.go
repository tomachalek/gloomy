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

package tools

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tomachalek/gloomy/index/gconf"
	"github.com/tomachalek/gloomy/vertical"
)

// NgramExtractor is used for extracting n-ngrams
// from a vertical file to a raw form
// TODO: specify format
type NgramExtractor struct {
	ngramSize   int
	buffer      *vertical.NgramBuffer
	stopWords   []string
	ignoreWords []string
	counter     map[string]int
}

func (n *NgramExtractor) isStopWord(w string) bool {
	for _, w2 := range n.stopWords {
		if w == w2 {
			return true
		}
	}
	return false
}

func (n *NgramExtractor) isIgnoreWord(w string) bool {
	for _, w2 := range n.ignoreWords {
		if w == w2 {
			return true
		}
	}
	return false
}

func (n *NgramExtractor) SaveNgrams(outPath string) error {
	out, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err) // TODO what can we do here?
	}
	for k, v := range n.counter {
		if v > 1 {
			_, werr := out.WriteString(fmt.Sprintf("%s %d\n", k, v))
			if werr != nil {
				panic(werr)
			}
		}
	}
	return nil
}

// ProcessLine processes a parsed vertical file line
func (n *NgramExtractor) ProcessLine(vline *vertical.Token) {
	if vline != nil {
		wordLC := vline.WordLC()
		if n.isStopWord(wordLC) {
			n.buffer.Reset()

		} else if !n.isIgnoreWord(wordLC) {
			n.buffer.AddToken(wordLC)
			if n.buffer.IsValid() {
				n.counter[n.buffer.Stringer()]++
			}
		}

	} else { // parser encoutered a structure
		n.buffer.Reset()
	}
}

// -----------------------------------------------------------

// ExtractNgrams runs the n-gram extraction process
func ExtractNgrams(conf *gconf.IndexBuilderConf, ngramSize int) {
	baseIndexPath := filepath.Join(conf.OutDirectory,
		fmt.Sprintf("ngrams-%s", filepath.Base(conf.VerticalFilePath)))
	extractor := &NgramExtractor{
		ngramSize:   ngramSize,
		buffer:      vertical.NewNgramBuffer(ngramSize),
		stopWords:   conf.NgramStopStrings,
		ignoreWords: conf.NgramIgnoreStrings,
		counter:     make(map[string]int),
	}
	vertical.ParseVerticalFile(conf.GetParserConf(), extractor)
	extractor.SaveNgrams(baseIndexPath)
}
