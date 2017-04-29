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
	"fmt"
	"github.com/tomachalek/gloomy/index"
	"github.com/tomachalek/gloomy/index/column"
	"github.com/tomachalek/gloomy/index/gconf"
	"github.com/tomachalek/gloomy/vertical"
	"github.com/tomachalek/gloomy/wdict"
	"log"
)

type IndexBuilder struct {
	outputFiles *gconf.OutputFiles

	ngramSize int

	minNgramFreq int

	ngramList *NgramList

	stopWords []string

	ignoreWords []string

	buffer *vertical.NgramBuffer

	wordDict *wdict.WordDictWriter

	nindex *index.DynamicNgramIndex
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
				meta := make([]column.AttrVal, b.nindex.MetadataWriter().NumCols())
				b.nindex.MetadataWriter().ForEachArg(func(i int, ad *column.ArgsDictWriter, col column.AttrValColumn) {
					if _, ok := vline.StructAttrs[ad.Name()]; ok {
						idx := ad.AddValue(vline.StructAttrs[ad.Name()])
						meta[i] = column.AttrVal(idx)
					}
				})
				b.ngramList.Add(b.buffer.GetValue(), meta)
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
	return &IndexBuilder{
		outputFiles:  outputFiles,
		ngramList:    &NgramList{},
		minNgramFreq: conf.MinNgramFreq,
		ngramSize:    ngramSize,
		buffer:       vertical.NewNgramBuffer(ngramSize),
		stopWords:    conf.NgramStopStrings,
		ignoreWords:  conf.NgramIgnoreStrings,
		wordDict:     wdict.NewWordDictWriter(),
		nindex:       index.NewDynamicNgramIndex(ngramSize, 10000, conf.Args), // TODO initial size
	}
}

func saveEncodedNgrams(builder *IndexBuilder, minFreq int) error {
	builder.wordDict.Finalize(builder.GetOutputFiles().GetIndexDir())
	builder.ngramList.DFSWalkthru(func(item *NgramNode) {
		if item.count >= minFreq {
			encodedNg := make([]int, len(item.ngram))
			for i, w := range item.ngram {
				encodedNg[i] = builder.wordDict.GetTokenIndex(w)
			}
			builder.nindex.AddNgram(encodedNg, item.count, item.args)
		}
	})
	builder.nindex.Finish()
	log.Printf("Done: %s", builder.nindex.GetInfo())
	builder.nindex.Save(builder.GetOutputFiles().GetIndexDir())
	return nil
}

func CreateGloomyIndex(conf *gconf.IndexBuilderConf, ngramSize int) {
	builder := CreateIndexBuilder(conf, ngramSize)
	vertical.ParseVerticalFile(conf.GetParserConf(), builder)
	saveEncodedNgrams(builder, conf.MinNgramFreq)
}
