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

package column

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

type ArgsList []*ArgsDict

func (a ArgsList) GetArg(ident string) *ArgsDict {
	for _, v := range a {
		if v.name == ident {
			return v
		}
	}
	return nil
}

func (a ArgsList) GetArgIdx(ident string) int {
	for i, v := range a {
		if v.name == ident {
			return i
		}
	}
	return -1
}

// -------------------------------------------------------------------

type ArgsDict struct {
	name    string
	colType string
	index   map[string]int
	counter int
}

func (ad *ArgsDict) Name() string { return ad.name }

func (ad *ArgsDict) AddValue(v string) int {
	_, ok := ad.index[v]
	if !ok {
		ad.index[v] = ad.counter
		ad.counter++
	}
	return ad.index[v]
}

func (ad *ArgsDict) Save(dirPath string) error {
	// TODO
	outPath := filepath.Join(dirPath, fmt.Sprintf("column_%s.dict", ad.name))
	f, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664)
	if err != nil {
		return err
	}
	defer f.Close()
	fw := bufio.NewWriter(f)
	defer fw.Flush()

	data := make([]string, len(ad.index))
	for k, v := range ad.index {
		data[v] = k
	}
	fw.WriteString(fmt.Sprintf("%d\n", len(data)))
	for _, v := range data {
		fw.WriteString(fmt.Sprintf("%s\n", v))
	}
	return nil
}

func NewArgsDict(name string, colType string) *ArgsDict {
	return &ArgsDict{
		name:    name,
		colType: colType,
		index:   make(map[string]int),
		counter: 0,
	}
}
