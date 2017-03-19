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
	"log"
	"os"
	"sort"
)

type wordDictExport struct {
	words   []string
	indices []int
}

func (wde *wordDictExport) Len() int {
	return len(wde.words)
}

func (wde *wordDictExport) Swap(i, j int) {
	wde.words[i], wde.words[j] = wde.words[j], wde.words[i]
	wde.indices[i], wde.indices[j] = wde.indices[j], wde.indices[i]
}

func (wde *wordDictExport) Less(i, j int) bool {
	return wde.words[i] < wde.words[j]
}

// ---------------------------------------

type WordDictBuilder struct {
	index   map[string]int
	counter int
}

func (w *WordDictBuilder) AddToken(token string) {
	_, ok := w.index[token]
	if !ok {
		w.index[token] = w.counter
		w.counter++
	}
}

func (w *WordDictBuilder) GetTokenIndex(token string) int {
	idx, ok := w.index[token]
	if ok {
		return idx
	}
	return 0
}

func saveWords(data []string, dstPath string) error {
	f, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY, 0664)
	defer f.Close()
	if err != nil {
		return err
	}
	fw := bufio.NewWriter(f)
	defer fw.Flush()
	log.Print("Words data len: ", len(data))
	for _, w := range data {
		//log.Print("w ", w)
		fw.WriteString(w + "\n")
	}
	return nil
}

func saveIndices(data []int, dstPath string) error {
	f, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY, 0664)
	defer f.Close()
	if err != nil {
		return err
	}
	fw := bufio.NewWriter(f)
	defer fw.Flush()
	werr := binary.Write(fw, binary.LittleEndian, int64(len(data)))
	log.Print("SAVING SIZE: ", int64(len(data)))
	if werr != nil {
		return werr
	}
	for _, v := range data {
		werr = binary.Write(fw, binary.LittleEndian, int64(v))
		if werr != nil {
			return werr
		}
	}
	return nil
}

func (w *WordDictBuilder) Save(dstPath string) error {
	we := wordDictExport{
		words:   make([]string, len(w.index)),
		indices: make([]int, len(w.index)),
	}
	for k, v := range w.index {
		we.words[v] = k
		we.indices[v] = v
	}
	sort.Sort(&we)
	err := saveWords(we.words, dstPath)
	if err != nil {
		return err
	}
	err = saveIndices(we.indices, dstPath+".idx")
	if err != nil {
		return err
	}
	return nil
}

func NewWordDictBuilder() *WordDictBuilder {
	return &WordDictBuilder{
		index: make(map[string]int),
	}
}
