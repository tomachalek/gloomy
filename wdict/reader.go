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

func loadWords(srcPath string) ([]string, error) {
	var ans []string
	f, err := os.Open(srcPath)
	if err != nil {
		return ans, err
	}
	fr := bufio.NewScanner(f)
	fr.Scan() // size
	size, err := strconv.ParseInt(fr.Text(), 10, 64)
	if err != nil {
		panic(err)
	}
	ans = make([]string, size)
	for i := 0; fr.Scan(); i++ {
		ans[i] = fr.Text()
	}
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

type WordDictReader struct {
	data []string
	wmap []*string
}

func LoadWordDict(dataPath string) (*WordDictReader, error) {
	words, err := loadWords(filepath.Join(dataPath, "words.dict"))
	if err != nil {
		return nil, err
	}
	wmap := make([]*string, len(words))
	for i := 0; i < len(words); i++ {
		wmap[i] = &words[i]
	}
	return &WordDictReader{data: words, wmap: wmap}, err
}

func (w *WordDictReader) Find(word string) int {
	left := 0
	right := len(w.data) - 1
	pivot := len(w.data) / 2
	for left < right && w.data[pivot] != word {
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

func (w *WordDictReader) DecodeNgram(ngram []int) []string {
	ans := make([]string, len(ngram))
	for i, val := range ngram {
		ans[i] = *w.wmap[val]
	}
	return ans
}
