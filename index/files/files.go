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

package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tomachalek/gloomy/vertical"
)

func stripSuffix(filePath string) string {
	i := strings.LastIndex(filePath, ".")
	if i > -1 {
		return filePath[:i]
	}
	return filePath
}

type OutputFiles struct {
	conf      *vertical.ParserConf
	indexDir  string
	filePerm  os.FileMode
	dirPerm   os.FileMode
	ngramSize int
}

func (o *OutputFiles) GetSortedIndexTmpPath(mode int) (*os.File, error) {
	inFilenamePrefix := stripSuffix(filepath.Base(o.conf.VerticalFilePath))
	fileName := fmt.Sprintf("%s_%d-grams.tmp", inFilenamePrefix, o.ngramSize)
	outPath := filepath.Join(o.indexDir, fileName)
	return os.OpenFile(outPath, mode, o.filePerm)
}

func (o *OutputFiles) OpenIndexForPosition(posIdx int, mode int) (*os.File, error) {
	indexPath := filepath.Join(o.indexDir, fmt.Sprintf("index-p%d.glm", posIdx))
	return os.OpenFile(indexPath, mode, o.filePerm)
}

func (o *OutputFiles) GetIndexDir() string {
	return o.indexDir
}

func NewOutputFiles(conf *vertical.ParserConf, ngramSize int, filePerm os.FileMode, dirPerm os.FileMode) *OutputFiles {
	inFilenamePrefix := stripSuffix(filepath.Base(conf.VerticalFilePath))
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
