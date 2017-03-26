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
	"fmt"
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

func (w *WordDictBuilder) Finalize(dstPath string) {
	tmp := make([]string, len(w.index))
	i := 0
	for k := range w.index {
		tmp[i] = k
		i++
	}
	sort.Strings(tmp)
	i = 0
	for _, v := range tmp {
		w.index[v] = i
		i++
	}
	w.save(tmp, dstPath)
}

func (w *WordDictBuilder) save(data []string, dstPath string) error {
	f, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY, 0664)
	defer f.Close()
	if err != nil {
		return err
	}
	fw := bufio.NewWriter(f)
	defer fw.Flush()
	log.Print("Words data len: ", len(data))
	fw.WriteString(fmt.Sprintf("%d\n", len(data)))
	for _, w := range data {
		fw.WriteString(w + "\n")
	}
	return nil
}

func NewWordDictBuilder() *WordDictBuilder {
	return &WordDictBuilder{
		index: make(map[string]int),
	}
}
