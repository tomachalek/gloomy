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

package wdict

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
)

// WordDictWriter writes a structure mapping words (string) to
// indices (int). In fact, it is a simple array of strings with
// similarly simple file representation where first item is a
// string encoded integer containing number of words and then
// there is a list of strings (all the values separated by LF).
type WordDictWriter struct {
	index   map[string]int
	counter int
}

// AddToken adds a single word to the dictionary.
// It is ok to add an already present value
// (in such case, nothing is done).
func (w *WordDictWriter) AddToken(token string) {
	_, ok := w.index[token]
	if !ok {
		w.index[token] = w.counter
		w.counter++
	}
}

// GetTokenIndex returns an array index within word
// dictionary of a specified token.
// If no such token is found then -1 is returned
// (normally, during data indexing, this should not happen)
func (w *WordDictWriter) GetTokenIndex(token string) int {
	idx, ok := w.index[token]
	if ok {
		return idx
	}
	return -1
}

// Finalize sorts the dictionary, attaches final
// indices (from 0 to N) to the tokens and saves the data.
// Please note that this means that before Finalize is called
// the indices are only temporary and cannot be used.
func (w *WordDictWriter) Finalize(dstPath string) {
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
	w.save(tmp, filepath.Join(dstPath, "words.dict"))
}

func (w *WordDictWriter) save(data []string, dstPath string) error {
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

// NewWordDictWriter creates a new instance
// of the WordDictWriter
func NewWordDictWriter() *WordDictWriter {
	return &WordDictWriter{
		index: make(map[string]int),
	}
}
