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
	"strconv"
)

// ArgsWriterList represents a list of ArgsDictWriter
// instances. The order is significant and is kept
// the same during indexing process.
type ArgsWriterList []*ArgsDictWriter

// -------------------------------------------------------------------

// ArgsDictWriter handles writing of metadata word <-> index dictionary
// during indexing.
type ArgsDictWriter struct {
	name    string
	index   map[string]int
	counter int
}

// Name returns ArgsDictWriter which is a respective
// metadata attribute name.
func (adw *ArgsDictWriter) Name() string { return adw.name }

// AddValue adds a new token to the dictionary.
func (adw *ArgsDictWriter) AddValue(v string) int {
	_, ok := adw.index[v]
	if !ok {
		adw.index[v] = adw.counter
		adw.counter++
	}
	return adw.index[v]
}

// Save saves the dictionary to a file. The name
// is generated automatically.
func (adw *ArgsDictWriter) Save(dirPath string) error {
	// TODO
	outPath := filepath.Join(dirPath, fmt.Sprintf("column_%s.dict", adw.name))
	f, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664)
	if err != nil {
		return err
	}
	defer f.Close()
	fw := bufio.NewWriter(f)
	defer fw.Flush()

	data := make([]string, len(adw.index))
	for k, v := range adw.index {
		data[v] = k
	}
	fw.WriteString(fmt.Sprintf("%d\n", len(data)))
	for _, v := range data {
		fw.WriteString(fmt.Sprintf("%s\n", v))
	}
	return nil
}

// NewArgsDictWriter creates a new ArgsDictWriter
// instance.
func NewArgsDictWriter(name string) *ArgsDictWriter {
	return &ArgsDictWriter{
		name:    name,
		index:   make(map[string]int),
		counter: 0,
	}
}

// -------------------------------------------------------------------

// ArgsReaderList represents a list of ArgsDictReader instances.
// The order is significant as it is mapped to the original order
// user requested when called a search.
type ArgsReaderList []*ArgsDictReader

// GetArg returns an ArgDictReader based on the attribute
// name. In case nothing is found, nil is returned.
func (a ArgsReaderList) GetArg(ident string) *ArgsDictReader {
	for _, v := range a {
		if v.name == ident {
			return v
		}
	}
	return nil
}

// GetArgIdx returns an index of ArgsDictReader instance
// identified by a respective attribute name. If nothing
// is found, -1 is returned.
func (a ArgsReaderList) GetArgIdx(ident string) int {
	for i, v := range a {
		if v.name == ident {
			return i
		}
	}
	return -1
}

// -------------------------------------------------------------------

// ArgsDictReader handles reading of a dictionary
// containing word <-> index mapping.
type ArgsDictReader struct {
	name  string
	index map[AttrVal]string
}

// Name returns name of a respective metadata attribute.
func (ad *ArgsDictReader) Name() string { return ad.name }

// LoadArgsDict loads a specific attribute dictionary.
func LoadArgsDict(dirPath string, ident string) (*ArgsDictReader, error) {
	filePath := filepath.Join(dirPath, fmt.Sprintf("column_%s.dict", ident))
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fr := bufio.NewScanner(f)
	fr.Scan() // size
	_, err = strconv.ParseInt(fr.Text(), 10, 64)
	if err != nil {
		return nil, err
	}
	dict := &ArgsDictReader{
		name:  ident,
		index: make(map[AttrVal]string),
	}
	for i := 0; fr.Scan(); i++ {
		dict.index[AttrVal(i)] = fr.Text()
	}
	return dict, nil
}
