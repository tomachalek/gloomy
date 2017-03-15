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
	"os"
	"sort"
)

type WordDict struct {
	index   map[string]int
	counter int
}

func (w *WordDict) AddToken(token string) {
	_, ok := w.index[token]
	if !ok {
		w.index[token] = w.counter
		w.counter++
	}
}

func (w *WordDict) GetTokenIndex(token string) int {
	idx, ok := w.index[token]
	if ok {
		return idx
	}
	return 0
}

func (w *WordDict) Save(dstPath string) error {
	tmp := make([]string, len(w.index))
	for k, v := range w.index {
		tmp[v] = k
	}
	f, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY, 0664)
	defer f.Close()
	if err != nil {
		return err
	}
	fw := bufio.NewWriter(f)
	sort.Strings(tmp)
	for _, w := range tmp {
		fw.WriteString(w + "\n")
	}
	return nil
}

func NewWordDict() *WordDict {
	return &WordDict{
		index: make(map[string]int),
	}
}
