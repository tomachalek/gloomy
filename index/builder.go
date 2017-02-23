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
	"os"
	"path/filepath"
	"sort"

	"github.com/tomachalek/gloomy/vertical"
	"github.com/tomachalek/gloomy/wstore"
)

type IndexBuilder struct {
	outDir        string
	baseIndexFile *os.File
	prevItem      *vertical.Token
	uniqSuccNum   map[string]map[string]bool
}

func (b *IndexBuilder) ProcessLine(vline *vertical.Token) {
	if vline != nil {
		//fmt.Println("LINE: ", vline)
		if b.prevItem != nil {
			if b.uniqSuccNum[b.prevItem.WordLC()] == nil {
				b.uniqSuccNum[b.prevItem.WordLC()] = make(map[string]bool)
			}
			b.uniqSuccNum[b.prevItem.WordLC()][vline.WordLC()] = true
		}
		b.prevItem = vline
	}
}

func createWord2IntDict(data map[string]map[string]bool, outPath string) error {
	index := make([]string, len(data))
	i := 0
	for k := range data {
		index[i] = k
		i++
	}
	sort.Strings(index)
	return wstore.SaveWordDict(index, outPath)
}

func CreateGloomyIndex(conf *vertical.ParserConf) {
	baseIndexPath := filepath.Join(conf.OutDirectory, "baseindex.glm")
	outFile, err := os.OpenFile(baseIndexPath, os.O_CREATE, 0)
	if err != nil {
		panic(err)
	}
	builder := &IndexBuilder{
		outDir:        conf.OutDirectory,
		baseIndexFile: outFile,
		uniqSuccNum:   make(map[string]map[string]bool),
	}
	vertical.ParseVerticalFile(conf, builder)

	wIndexPath := filepath.Join(conf.OutDirectory, "wordindex.glm")
	createWord2IntDict(builder.uniqSuccNum, wIndexPath)
	//fmt.Println("MAP: ", builder.uniqSuccNum)
}
