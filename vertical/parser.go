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

package vertical

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

const (
	channelChunkSize = 30000 // please note that this affects the performance quite a lot
)

// --------------------------------------------------------

// ParserConf contains configuration parameters for
// vertical file parser
type ParserConf struct {

	// Source vertical file (either a plain text file or a gzip one)
	VerticalFilePath string `json:"verticalFilePath"`

	FilterArgs map[string]string `json:"filterArgs"`

	NgramIgnoreStructs []string `json:"ngramIgnoreStructs"`

	OutDirectory string `json:"outDirectory"`

	NgramStopStrings []string `json:"ngramStopStrings"`

	NgramIgnoreStrings []string `json:"ngramIgnoreStrings"`
}

// LoadConfig loads the configuration from a JSON file.
// In case of an error the program exits with panic.
func LoadConfig(path string) *ParserConf {
	rawData, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var conf ParserConf
	err = json.Unmarshal(rawData, &conf)
	if err != nil {
		panic(err)
	}
	return &conf
}

// --------------------------------------------------------

// Token is a representation of
// a parsed line. It connects both, positional attributes
// and currently accumulated structural attributes.
type Token struct {
	Word        string
	Attrs       []string
	StructAttrs map[string]string
}

func (v *Token) WordLC() string {
	return strings.ToLower(v.Word)
}

// --------------------------------------------------------

type VerticalMetaLine struct {
	Name  string
	Attrs map[string]string
}

// --------------------------------------------------------

type LineProcessor interface {
	ProcessLine(vline *Token)
}

// --------------------------------------------------------

func isOpenElement(tagSrc string) bool {
	return strings.HasPrefix(tagSrc, "<") && !strings.HasPrefix(tagSrc, "</") &&
		!strings.HasSuffix(tagSrc, "/>")
}

func isCloseElement(tagSrc string) bool {
	return strings.HasPrefix(tagSrc, "</")
}

func isSelfCloseElement(tagSrc string) bool {
	return strings.HasPrefix(tagSrc, "<") && strings.HasSuffix(tagSrc, "/>")
}

func parseAttrVal(src string) map[string]string {
	ans := make(map[string]string)
	rg := regexp.MustCompile("(\\w+)=\"([^\"]+)\"")
	srch := rg.FindAllStringSubmatch(src, -1)
	for i := 0; i < len(srch); i++ {
		ans[srch[i][1]] = srch[i][2]
	}
	return ans
}

func parseLine(line string, elmStack *Stack) *Token {
	var meta *VerticalMetaLine
	rg := regexp.MustCompile("<([\\w]+)(\\s*[^>]*)|>")
	srch := rg.FindStringSubmatch(line)

	switch {
	case isOpenElement(line):
		meta = &VerticalMetaLine{Name: srch[1], Attrs: parseAttrVal(srch[2])}
		elmStack.Push(meta)
	case isCloseElement(line):
		elmStack.Pop()
	case isSelfCloseElement(line):
		meta = &VerticalMetaLine{Name: srch[1], Attrs: parseAttrVal(srch[2])}
	default:
		items := strings.Split(line, "\t")
		return &Token{
			Word:        items[0],
			Attrs:       items[1:],
			StructAttrs: elmStack.GetAttrs(),
		}
	}
	return nil
}

// ParseVerticalFile processes a corpus vertical file
// line by line and applies a custom LineProcessor on
// them. The processing is parallelized in the sense
// that reading a file into lines and processing of
// the lines runs in different goroutines. To reduce
// overhead, the data are passed between goroutines
// in chunks.
func ParseVerticalFile(conf *ParserConf, lproc LineProcessor) {
	f, err := os.Open(conf.VerticalFilePath)
	if err != nil {
		panic(err)
	}

	var rd io.Reader
	if strings.HasSuffix(conf.VerticalFilePath, ".gz") {
		rd, err = gzip.NewReader(f)
		if err != nil {
			panic(err)
		}

	} else {
		rd = f
	}
	brd := bufio.NewScanner(rd)

	stack := NewStack()

	ch := make(chan []*Token)
	chunk := make([]*Token, channelChunkSize)
	go func() {
		i := 0
		for brd.Scan() {
			line := parseLine(brd.Text(), stack)
			chunk[i] = line
			i++
			if i == channelChunkSize {
				i = 0
				ch <- chunk
			}
		}
		if i > 0 {
			ch <- chunk[:i]
		}
		close(ch)
	}()

	for items := range ch {
		for _, item := range items {
			lproc.ProcessLine(item)
		}
	}

	fmt.Println("DONE: stack size: ", stack.Size())
}

//ParseVerticalFileNoGoRo is just for benchmarking purposes
func ParseVerticalFileNoGoRo(conf *ParserConf, lproc LineProcessor) {
	f, err := os.Open(conf.VerticalFilePath)
	if err != nil {
		panic(err)
	}
	rd := bufio.NewScanner(f)
	stack := NewStack()

	for rd.Scan() {
		line := parseLine(rd.Text(), stack)
		lproc.ProcessLine(line)
	}

	fmt.Println("DONE: stack size: ", stack.Size())
}
