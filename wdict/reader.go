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
	"encoding/binary"
	"os"
	"path/filepath"
	"strconv"
)

func loadWords(srcPath string) (*WordDictReader, error) {
	f, err := os.Open(srcPath)
	if err != nil {
		return nil, err
	}
	fr := bufio.NewScanner(f)
	fr.Scan() // size
	size, err := strconv.ParseInt(fr.Text(), 10, 64)
	if err != nil {
		panic(err)
	}
	words := make([]string, size)
	tree := NewRadixTree()
	for i := 0; fr.Scan(); i++ {
		words[i] = fr.Text()
		tree.Add(fr.Text(), i)
	}
	ans := &WordDictReader{data: words, tree: tree}
	return ans, nil
}

func loadIndices(srcPath string) ([]int, error) {
	var ans []int
	f, err := os.Open(srcPath)
	if err != nil {
		return ans, err
	}
	defer f.Close()
	fr := bufio.NewReader(f)
	var readErr error
	var size int64

	readErr = binary.Read(fr, binary.LittleEndian, &size)
	if readErr != nil {
		return ans, readErr
	}
	var tmp int64
	ans = make([]int, size)
	for i := 0; i < int(size); i++ {
		readErr = binary.Read(fr, binary.LittleEndian, &tmp)
		ans[i] = int(tmp)
		if readErr != nil {
			return ans, readErr
		}
	}
	return ans, nil
}

// WordDictReader translates words (string) to index values (int)
// and vice versa
type WordDictReader struct {
	data []string
	tree *RadixTree
}

// LoadWordDict loads a word dictionary from a specified
// directory (file name is determined automatically).
func LoadWordDict(dataPath string) (*WordDictReader, error) {
	return loadWords(filepath.Join(dataPath, "words.dict"))
}

// Find searches (in O(log(n) time) for an index value
// of a specified word.
// In case the word is not found, -1 is returned.
func (w *WordDictReader) Find(word string) int {
	left := 0
	right := len(w.data) - 1
	pivot := len(w.data) / 2
	for right-left > 1 && w.data[pivot] != word {
		if w.data[left] <= word && word <= w.data[pivot] {
			tmp := pivot
			pivot = (left + pivot) / 2
			right = tmp

		} else if w.data[pivot] < word && word <= w.data[right] {
			tmp := pivot
			pivot = (pivot + right) / 2
			left = tmp

		} else {
			break
		}
	}
	if word == w.data[pivot] {
		return pivot
	}
	return -1
}

func (w *WordDictReader) FindByPrefix(prefix string) []int {
	return w.tree.FindIndicesByPrefix(prefix)
}

// DecodeNgram finds a string representation of a word array (= n-gram).
func (w *WordDictReader) DecodeNgram(ngram []int) []string {
	ans := make([]string, len(ngram))
	for i, val := range ngram {
		ans[i] = w.data[val]
	}
	return ans
}

func (w *WordDictReader) DecodeToken(widx int) string {
	return w.data[widx]
}
