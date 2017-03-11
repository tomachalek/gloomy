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

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/tomachalek/gloomy/index/gconf"
	"github.com/tomachalek/gloomy/vertical"
	"log"
	"os"
)

type IndexBuilder struct {
	outputFiles *gconf.OutputFiles

	// index "word -> index"
	baseIndexFile *os.File

	// indices for n-gram positions 2, 3, 4
	// [number -> number]
	posIndices []*os.File

	ngramSize int

	minNgramFreq int

	ngramList *NgramList

	stopWords []string

	ignoreWords []string

	buffer *vertical.NgramBuffer

	wordDict *WordDict
}

func (b *IndexBuilder) GetOutputFiles() *gconf.OutputFiles {
	return b.outputFiles
}

func (b *IndexBuilder) GetNgramList() *NgramList {
	return b.ngramList
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
			b.wordDict.AddToken(wordLC)
			if b.buffer.IsValid() {
				b.ngramList.Add(b.buffer.GetValue())
			}
		}

	} else { // parser encoutered a structure
		b.buffer.Reset()
	}
}

func (b *IndexBuilder) CreateIndices() {
	counters := make([][]int, b.ngramSize-1)
	b.ngramList.DFSWalkthru(func(item *NgramNode) {
		if item.count >= b.minNgramFreq {
			for i := range counters {
				fmt.Println(i) // TODO
			}
		}
	})
}

func CreateIndexBuilder(conf *gconf.IndexBuilderConf, ngramSize int) *IndexBuilder {
	outputFiles := gconf.NewOutputFiles(conf, ngramSize, 0644, 0755)
	baseIndexFile, err := outputFiles.OpenIndexForPosition(0, os.O_CREATE|os.O_TRUNC|os.O_WRONLY)
	if err != nil {
		panic(err)
	}
	return &IndexBuilder{
		outputFiles:   outputFiles,
		baseIndexFile: baseIndexFile,
		ngramList:     &NgramList{},
		minNgramFreq:  conf.MinNgramFreq,
		ngramSize:     ngramSize,
		buffer:        vertical.NewNgramBuffer(ngramSize),
		stopWords:     conf.NgramStopStrings,
		ignoreWords:   conf.NgramIgnoreStrings,
		wordDict:      NewWordDict(),
	}
}

func saveEncodedNgrams(builder *IndexBuilder, minFreq int, saveFile *os.File) error {
	fw := bufio.NewWriter(saveFile)
	defer fw.Flush()
	builder.ngramList.DFSWalkthru(func(item *NgramNode) {
		if item.count >= minFreq {
			for _, w := range item.ngram {
				binary.Write(fw, binary.LittleEndian, builder.wordDict.GetTokenIndex(w))
			}
			//(fmt.Sprintf("%s\t%d\n", strings.Join(item.ngram, "\t"), item.count))
		}
	})
	return nil
}

func CreateGloomyIndex(conf *gconf.IndexBuilderConf, ngramSize int) {
	builder := CreateIndexBuilder(conf, ngramSize)
	vertical.ParseVerticalFile(conf.GetParserConf(), builder)
	sortedIndexTmp, err := builder.outputFiles.GetSortedIndexTmpPath(os.O_CREATE | os.O_TRUNC | os.O_WRONLY)
	if err != nil {
		panic(err)
	}
	//saveNgrams(builder.ngramList, conf.MinNgramFreq, sortedIndexTmp)
	saveEncodedNgrams(builder, conf.MinNgramFreq, sortedIndexTmp)
	log.Printf("Saved raw n-gram file %s", sortedIndexTmp.Name())
}
