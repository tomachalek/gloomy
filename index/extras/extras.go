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

package extras

import (
	"bufio"
	"fmt"
	"github.com/tomachalek/gloomy/index/builder"
	"github.com/tomachalek/gloomy/index/gconf"
	"github.com/tomachalek/vertigo"
	"log"
	"os"
	"strings"
)

func saveNgrams(ngramList *builder.NgramList, minFreq int, saveFile *os.File) error {
	fw := bufio.NewWriter(saveFile)
	defer fw.Flush()
	ngramList.DFSWalkthru(func(item *builder.NgramNode) {
		if item.GetCount() >= minFreq {
			fw.WriteString(fmt.Sprintf("%s\t%d\n", strings.Join(item.GetNgram(), "\t"), item.GetCount()))
		}
	})
	return nil
}

func ExtractUniqueNgrams(conf *gconf.IndexBuilderConf, ngramSize int) {
	builder := builder.CreateIndexBuilder(conf, ngramSize)
	vertigo.ParseVerticalFile(conf.GetParserConf(), builder)
	sortedIndexTmp, err := builder.GetOutputFiles().GetSortedIndexTmpPath(os.O_CREATE | os.O_TRUNC | os.O_WRONLY)
	if err != nil {
		panic(err)
	}
	saveNgrams(builder.GetNgramList(), conf.MinNgramFreq, sortedIndexTmp)
	log.Printf("Saved raw n-gram file %s", sortedIndexTmp.Name())
}
