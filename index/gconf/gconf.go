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

package gconf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/tomachalek/vertigo"
)

func stripSuffix(filePath string) string {
	i := strings.LastIndex(filePath, ".")
	if i > -1 {
		return filePath[:i]
	}
	return filePath
}

// ---------------------------------------------------------

type SearchConf struct {
	DataPath      string `json:"dataPath"`
	ServerPort    int    `json:"serverPort"`
	ServerAddress string `json:"serverAddress"`
}

func LoadSearchConf(confPath string) *SearchConf {
	rawData, err := ioutil.ReadFile(confPath)
	if err != nil {
		panic(err)
	}
	var conf SearchConf
	err = json.Unmarshal(rawData, &conf)
	if err != nil {
		panic(err)
	}
	return &conf
}

// ---------------------------------------------------------

type IndexBuilderConf struct {
	vertigo.ParserConf

	SourceType string `json:"sourceType"`

	OutDirectory string `json:"outDirectory"`

	MinNgramFreq int `json:"minNgramFreq"`

	Args map[string]string `json:"args"`

	NgramIgnoreStructs []string `json:"ngramIgnoreStructs"`

	NgramStopStrings []string `json:"ngramStopStrings"`

	NgramIgnoreStrings []string `json:"ngramIgnoreStrings"`

	NgramMatchPrefix []string `json:"ngramMatchPrefix"`

	UseStrictStructParser bool `json:"useStrictStructParser"`
}

func (i *IndexBuilderConf) GetParserConf() *vertigo.ParserConf {
	accumType := vertigo.AccumulatorTypeComb
	if i.UseStrictStructParser {
		accumType = vertigo.AccumulatorTypeStack
	}
	return &vertigo.ParserConf{
		InputFilePath:         i.InputFilePath,
		FilterArgs:            i.FilterArgs,
		StructAttrAccumulator: accumType,
	}
}

func LoadIndexBuilderConf(confPath string) *IndexBuilderConf {
	rawData, err := ioutil.ReadFile(confPath)
	if err != nil {
		panic(err)
	}
	var conf IndexBuilderConf
	err = json.Unmarshal(rawData, &conf)
	if err != nil {
		panic(err)
	}
	return &conf
}

// ---------------------------------------------------------

type OutputFiles struct {
	conf      *IndexBuilderConf
	indexDir  string
	filePerm  os.FileMode
	dirPerm   os.FileMode
	ngramSize int
}

func (o *OutputFiles) GetSortedIndexTmpPath(mode int) (*os.File, error) {
	inFilenamePrefix := stripSuffix(filepath.Base(o.conf.InputFilePath))
	fileName := fmt.Sprintf("%s_%d-grams.tmp", inFilenamePrefix, o.ngramSize)
	outPath := filepath.Join(o.indexDir, fileName)
	return os.OpenFile(outPath, mode, o.filePerm)
}

func (o *OutputFiles) GetIndexDir() string {
	return o.indexDir
}

func NewOutputFiles(conf *IndexBuilderConf, ngramSize int, filePerm os.FileMode, dirPerm os.FileMode) *OutputFiles {
	inFilenamePrefix := stripSuffix(filepath.Base(conf.InputFilePath))
	outDir := filepath.Join(conf.OutDirectory, inFilenamePrefix)
	err := os.MkdirAll(outDir, dirPerm)
	if err != nil {
		panic(err)
	}
	return &OutputFiles{
		conf:      conf,
		indexDir:  outDir,
		filePerm:  filePerm,
		dirPerm:   dirPerm,
		ngramSize: ngramSize,
	}
}
