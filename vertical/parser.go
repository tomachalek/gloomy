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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

// --------------------------------------------------------

type ParserConf struct {
	VerticalFilePath   string            `json:"verticalFilePath"`
	FilterArgs         map[string]string `json:"filterArgs"`
	NgramIgnoreStructs []string          `json:"ngramIgnoreStructs"`
	OutDirectory       string            `json:"outDirectory"`
}

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

type VerticalLine struct {
	Word  string
	Attrs []string
}

func (v *VerticalLine) WordLC() string {
	return strings.ToLower(v.Word)
}

// --------------------------------------------------------

type VerticalMetaLine struct {
	Name  string
	Attrs map[string]string
}

// --------------------------------------------------------

type LineProcessor interface {
	ProcessLine(vline *VerticalLine)
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

func parseLine(line string, elmStack *Stack) *VerticalLine {
	var meta *VerticalMetaLine
	rg := regexp.MustCompile("<([\\w]+)(\\s*[^>]*)|>")
	srch := rg.FindStringSubmatch(line)

	switch {
	case isOpenElement(line):
		meta = &VerticalMetaLine{Name: srch[1], Attrs: parseAttrVal(srch[2])}
		elmStack.Push(meta)
		fmt.Println(meta)
		break
	case isCloseElement(line):
		elmStack.Pop()
		break
	case isSelfCloseElement(line):
		meta = &VerticalMetaLine{Name: srch[1], Attrs: parseAttrVal(srch[2])}
	default:
		items := strings.Split(line, "\t")
		return &VerticalLine{
			Word:  items[0],
			Attrs: items[1:],
		}
	}
	return nil
}

func ParseVerticalFile(conf *ParserConf, lproc LineProcessor) {
	f, err := os.Open(conf.VerticalFilePath)
	if err != nil {
		panic(err)
	}
	rd := bufio.NewScanner(f)
	stack := NewStack()
	for rd.Scan() {
		line := parseLine(rd.Text(), stack)
		fmt.Println("LINE: ", line)
		fmt.Println("ROLL: ", stack.GetAttrs())
		fmt.Println("----------------------")
		lproc.ProcessLine(line)
	}
	fmt.Println("DONE: stack size: ", stack.Size())
}
