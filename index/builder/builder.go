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
	"log"
	"strings"

	"github.com/tomachalek/gloomy/index"
	"github.com/tomachalek/gloomy/index/builder/tokenizer"
	"github.com/tomachalek/gloomy/index/column"
	"github.com/tomachalek/gloomy/index/gconf"
	"github.com/tomachalek/gloomy/wdict"
	"github.com/tomachalek/vertigo"
)

type NgramRecord struct {
	Ngram []string
	Count int
	Args  []column.AttrVal
}

// NgramList specifies a required ngram list implementation
// Gloomy provides a simple in-memory implementation and
// a more advanced one operating on multiple file chunks
// for large data
type NgramList interface {
	ForEach(fn func(n *NgramRecord))

	Size() int

	Add(ngram []string, metadata []column.AttrVal)
}

// IndexBuilder is an object for creating n-gram indices
type IndexBuilder struct {
	outputFiles *gconf.OutputFiles

	ngramSize int

	minNgramFreq int

	ngramList NgramList

	stopWords []string

	ignoreWords []string

	matchPrefixWords []string

	buffer *NgramBuffer

	wordDict *wdict.WordDictWriter

	nindex *index.DynamicNgramIndex
}

func (b *IndexBuilder) GetOutputFiles() *gconf.OutputFiles {
	return b.outputFiles
}

func (b *IndexBuilder) GetNgramList() NgramList {
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

func (b *IndexBuilder) usesWhitelist() bool {
	return len(b.matchPrefixWords) > 0
}

func (b *IndexBuilder) matchesPrefixWords(buff *NgramBuffer) bool {
	buffVal := buff.GetValue()
	for i, w := range buffVal {
		if i >= len(b.matchPrefixWords) {
			return true
		}
		if !strings.HasPrefix(w, b.matchPrefixWords[i]) {
			return false
		}
	}
	return true
}

func (b *IndexBuilder) ProcStruct(vline *vertigo.Structure) {}

func (b *IndexBuilder) ProcStructClose(vline *vertigo.StructureClose) {}

func (b *IndexBuilder) ProcToken(vline *vertigo.Token) {
	if vline != nil {
		wordLC := vline.WordLC()
		if b.isStopWord(wordLC) {
			b.buffer.Reset()

		} else if !b.isIgnoreWord(wordLC) {
			b.buffer.AddToken(wordLC)
			b.wordDict.AddToken(wordLC)
			if b.buffer.IsValid() && !b.usesWhitelist() || b.matchesPrefixWords(b.buffer) {
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
	b.ngramList.ForEach(func(item *NgramRecord) {
		if item.Count >= b.minNgramFreq {
			for i := range counters {
				fmt.Println(i) // TODO
			}
		}
	})
}

func CreateIndexBuilder(conf *gconf.IndexBuilderConf, ngramSize int) *IndexBuilder {
	outputFiles := gconf.NewOutputFiles(conf, ngramSize, 0644, 0755)

	var ngramList NgramList
	if conf.ProcChunkSize == 0 {
		ngramList = &RAMNgramList{}

	} else {
		if conf.TmpDir == "" {
			log.Panic("A 'tmpDir' must be configured in case procChunkSize > 0")
		}
		ngramList = NewLargeNgramList(conf.TmpDir, conf.ProcChunkSize)
	}

	return &IndexBuilder{
		outputFiles:      outputFiles,
		ngramList:        ngramList,
		minNgramFreq:     conf.MinNgramFreq,
		ngramSize:        ngramSize,
		buffer:           NewNgramBuffer(ngramSize),
		stopWords:        conf.NgramStopStrings,
		ignoreWords:      conf.NgramIgnoreStrings,
		matchPrefixWords: conf.NgramMatchPrefix,
		wordDict:         wdict.NewWordDictWriter(),
		nindex:           index.NewDynamicNgramIndex(ngramSize, 10000, conf.Args), // TODO initial size
	}
}

func saveEncodedNgrams(builder *IndexBuilder, minFreq int) error {
	builder.wordDict.Finalize(builder.GetOutputFiles().GetIndexDir())
	builder.ngramList.ForEach(func(item *NgramRecord) {
		if item.Count >= minFreq {
			encodedNg := make([]int, len(item.Ngram))
			for i, w := range item.Ngram {
				encodedNg[i] = builder.wordDict.GetTokenIndex(w)
			}
			builder.nindex.AddNgram(encodedNg, item.Count, item.Args)
		}
	})
	builder.nindex.Finish()
	log.Printf("Done: %s", builder.nindex.GetInfo())
	builder.nindex.Save(builder.GetOutputFiles().GetIndexDir())
	return nil
}

// CreateGloomyIndex is a high level function which based on
// provided configuration creates an n-gram index.
func CreateGloomyIndex(conf *gconf.IndexBuilderConf, ngramSize int) {
	builder := CreateIndexBuilder(conf, ngramSize)
	var procErr error

	switch conf.SourceType {
	case "vertical":
		procErr = vertigo.ParseVerticalFile(conf.GetParserConf(), builder)
	case "plain":
		procErr = tokenizer.ParseFile(conf.GetParserConf(), builder)
	default:
		panic(fmt.Errorf("Unknow source type: %s", conf.GetParserConf()))

	}

	if procErr == nil {
		saveEncodedNgrams(builder, conf.MinNgramFreq)

	} else {
		log.Panicf("Failed to process source with error: %s", procErr)
	}
}
