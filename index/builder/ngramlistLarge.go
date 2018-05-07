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
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/tomachalek/gloomy/index/column"
	"github.com/tomachalek/gloomy/util"
)

// ----------------------------------------------------

type chunkReader struct {
	path     string
	file     *os.File
	reader   *bufio.Reader
	decoder  *gob.Decoder
	buffer   bytes.Buffer
	currItem *NgramRecord
	finished bool
}

func (ch *chunkReader) readNext() {
	var ans NgramRecord
	err := ch.decoder.Decode(&ans)
	if err == io.EOF {
		ch.finished = true

	} else if err != nil {
		log.Print("Failed to decode ngramRecord", err)

	} else {
		ch.currItem = &ans
	}
}

func (ch *chunkReader) hasNext() bool {
	return !ch.finished
}

func (ch *chunkReader) getCurrent() *NgramRecord {
	return ch.currItem
}

func newChunkReader(path string) (*chunkReader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(f)
	ans := &chunkReader{
		path:     path,
		file:     f,
		reader:   reader,
		finished: false,
	}
	ans.decoder = gob.NewDecoder(reader)
	return ans, nil
}

// ----------------------------------------------------

type LargeNgramList struct {
	currNgramList  *RAMNgramList
	workingDirPath string
	chunks         []string
	chunkSize      int
}

func NewLargeNgramList(workingDirPath string, chunkSize int) *LargeNgramList {
	if !util.IsDir(workingDirPath) {
		err := os.MkdirAll(workingDirPath, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
	return &LargeNgramList{
		currNgramList:  &RAMNgramList{},
		workingDirPath: workingDirPath,
		chunks:         make([]string, 0, 10),
		chunkSize:      chunkSize,
	}
}

func (nn *LargeNgramList) Size() int {
	return nn.currNgramList.Size()
}

func (nn *LargeNgramList) Add(ngram []string, metadata column.Metadata) {
	nn.currNgramList.Add(ngram, metadata)
	if nn.currNgramList.Size() >= nn.chunkSize {
		nn.saveChunk()
		nn.currNgramList = &RAMNgramList{}
	}
}

func (nn *LargeNgramList) generateNewChunkFileName() string {
	return filepath.Join(nn.workingDirPath, fmt.Sprintf("chunk-%03d", len(nn.chunks)))
}

func (nn *LargeNgramList) saveChunk() error {
	chunkPath := nn.generateNewChunkFileName()
	log.Printf("Saving chunk %s", chunkPath)
	f, err := os.OpenFile(chunkPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664)
	defer f.Close()
	if err != nil {
		return err
	}
	fw := bufio.NewWriter(f)
	defer fw.Flush()

	enc := gob.NewEncoder(fw)
	nn.currNgramList.ForEach(func(n *NgramRecord) {
		err := enc.Encode(n)
		if err != nil {
			log.Printf("Failed to encode record: %s", err)
		}
	})
	nn.chunks = append(nn.chunks, chunkPath)
	return nil
}

func findFirstNonEmptyReader(readers []*chunkReader) *chunkReader {
	for _, v := range readers {
		if v.hasNext() {
			return v
		}
	}
	return nil
}

func (nn *LargeNgramList) findSmallestNgram(readers []*chunkReader) *NgramRecord {
	smallestIdx := 0
	firstReader := findFirstNonEmptyReader(readers)
	if firstReader == nil {
		return nil
	}
	smallestRec := firstReader.getCurrent()
	for i := 1; i < len(readers); i++ {
		if !readers[i].hasNext() {
			continue
		}
		switch ngramsCmp(readers[i].getCurrent().Ngram, smallestRec.Ngram) {
		case -1:
			smallestIdx = i
			smallestRec = readers[i].getCurrent()
		case 0:
			smallestRec.Count += readers[i].getCurrent().Count
			if readers[i].hasNext() {
				readers[i].readNext()
			}
		}
	}
	if readers[smallestIdx].hasNext() {
		readers[smallestIdx].readNext()
	}
	return smallestRec
}

func (nn *LargeNgramList) ForEach(fn func(n *NgramRecord)) {
	if nn.currNgramList.Size() > 0 {
		nn.saveChunk()
	}
	readers := make([]*chunkReader, len(nn.chunks))
	var err error
	for i, c := range nn.chunks {
		readers[i], err = newChunkReader(c)
		if err != nil {
			panic(err) // TODO
		}
		readers[i].readNext() // read 1st item
	}

	var curr *NgramRecord
	for curr = nn.findSmallestNgram(readers); curr != nil; curr = nn.findSmallestNgram(readers) {
		fn(curr)
	}
}
