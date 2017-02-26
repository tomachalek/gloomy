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

package index

import (
	"bufio"
	"fmt"
	"github.com/tomachalek/gloomy/vertical"
	"os"
	"path/filepath"
	"strings"
)

type IndexBuilder struct {
	outDir        string
	baseIndexFile *os.File
	prevItem      *vertical.Token
	ngramSize     int
	ngramList     *NgramList
	stopWords     []string
	ignoreWords   []string
	buffer        *vertical.NgramBuffer
}

func (b *IndexBuilder) isStopWord(w string) bool {
	for _, w2 := range b.stopWords {
		if w == w2 {
			return true
		}
	}
	return false
}

func (b *IndexBuilder) isIgnoreWord(w string) bool {
	for _, w2 := range b.ignoreWords {
		if w == w2 {
			return true
		}
	}
	return false
}

func (b *IndexBuilder) ProcessLine(vline *vertical.Token) {
	if vline != nil {
		wordLC := vline.WordLC()
		if b.isStopWord(wordLC) {
			b.buffer.Reset()

		} else if !b.isIgnoreWord(wordLC) {
			b.buffer.AddToken(wordLC)
			if b.buffer.IsValid() {
				b.ngramList.Add(b.buffer.GetValue())
			}
		}

	} else { // parser encoutered a structure
		b.buffer.Reset()
	}
}

func createWord2IntDict(ngramList *NgramList, outPath string) error {
	// TODO
	return nil
}

func saveNgrams(ngramList *NgramList, savePath string) error {
	f, err := os.OpenFile(savePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	fw := bufio.NewWriter(f)
	defer fw.Flush()
	ngramList.DFSWalkthru(func(item *NgramNode) {
		fw.WriteString(fmt.Sprintf("%s %d\n", strings.Join(item.ngram, " "), item.count))
	})
	return nil
}

func CreateGloomyIndex(conf *vertical.ParserConf, ngramSize int) {
	baseIndexPath := filepath.Join(conf.OutDirectory, "baseindex.glm")
	outFile, err := os.OpenFile(baseIndexPath, os.O_CREATE, 0)
	if err != nil {
		panic(err)
	}
	builder := &IndexBuilder{
		outDir:        conf.OutDirectory,
		baseIndexFile: outFile,
		ngramList:     &NgramList{},
		ngramSize:     ngramSize,
		buffer:        vertical.NewNgramBuffer(ngramSize),
		stopWords:     conf.NgramStopStrings,
		ignoreWords:   conf.NgramIgnoreStrings,
	}
	vertical.ParseVerticalFile(conf, builder)

	wIndexPath := filepath.Join(conf.OutDirectory, "tmp_ngrams.glm")
	saveNgrams(builder.ngramList, wIndexPath)
}
