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

package service

import (
	"fmt"
	"github.com/tomachalek/gloomy/index"
	"github.com/tomachalek/gloomy/service/query"
	"github.com/tomachalek/gloomy/wdict"
	"path/filepath"
	"regexp"
	"strings"
)

type SearchResultItem struct {
	Ngram []string `json:"ngram"`
	Count int      `json:"count"`
	Args  []string `json:"args"`
}

// --------------------------------------------------------------

type SearchArgs struct {
	CorpusID  string
	Phrase    string
	Attrs     []string
	Offset    int
	Limit     int
	QueryType int
}

func (s SearchArgs) clone() SearchArgs {
	var newAttrs []string
	copy(newAttrs, s.Attrs)
	return SearchArgs{
		CorpusID:  s.CorpusID,
		Phrase:    s.Phrase,
		Attrs:     newAttrs,
		Offset:    s.Offset,
		Limit:     s.Limit,
		QueryType: s.QueryType,
	}
}

// ---------------------------------------------------------------

type SearchResult struct {
	result *index.NgramSearchResult
	wdict  *wdict.WordDictReader
}

func (sr *SearchResult) Size() int {
	return sr.result.Size()
}

func (sr *SearchResult) HasNext() bool {
	return sr.result.HasNext()
}

func (sr *SearchResult) Next() *SearchResultItem {
	ans := sr.result.Next()
	if ans != nil {
		return &SearchResultItem{
			Ngram: sr.wdict.DecodeNgram(ans.Ngram),
			Count: ans.Count,
			Args:  ans.Metadata,
		}
	}
	return nil
}

// ---------------------------------------------------------------

func loadRange(index *index.SearchableIndex, indices []int) {
	min := indices[0]
	max := indices[0]
	for _, v := range indices {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	index.LoadRange(min, max)
}

func translateWidxToColIdx(index *index.SearchableIndex, indices []int) []int {
	wi := 0
	for i := 0; i < len(indices); i++ {
		indices[wi] = index.GetCol0Idx(indices[i])
		if indices[wi] > -1 {
			wi++
		}
	}
	return indices[:wi]
}

func searchByPrefix(wd *wdict.WordDictReader, sindex *index.SearchableIndex, args SearchArgs) *index.NgramSearchResult {
	var res *index.NgramSearchResult
	indices := wd.FindByPrefix(args.Phrase[:len(args.Phrase)-1])
	indices = translateWidxToColIdx(sindex, indices)
	loadRange(sindex, indices)
	ch := make(chan *index.NgramSearchResult, len(indices))
	for _, colIdx := range indices {
		go func(v int) {
			ch <- sindex.GetNgramsOfColIdx(v)
		}(colIdx)
	}
	var chunk *index.NgramSearchResult
	for range indices {
		chunk = <-ch
		if res == nil {
			res = chunk

		} else {
			res.Append(chunk)
		}
	}
	close(ch)
	return res
}

func searchByRegexp(wd *wdict.WordDictReader, sindex *index.SearchableIndex, args SearchArgs) *index.NgramSearchResult {
	parser := query.NewParser()
	parser.Parse(args.Phrase)
	prefixes := parser.GetAllPrefixes()

	ans := &index.NgramSearchResult{}
	for _, prefix := range prefixes {
		args2 := args.clone()
		args2.Phrase = prefix
		args2.QueryType = 0 // not needed here; just to keep things consistent
		if strings.HasSuffix(args2.Phrase, "*") {
			ans.Append(searchByPrefix(wd, sindex, args2))

		} else {
			ans.Append(sindex.GetNgramsOf(args2.Phrase))
		}
	}
	rg := regexp.MustCompile(fmt.Sprintf("^%s$", args.Phrase))
	rgList := []*regexp.Regexp{rg} // TODO currently only the fist word
	ans.Filter(func(v *index.NgramResultItem) bool {
		ngram := wd.DecodeNgram(v.Ngram)
		matches := false
		for i, ptr := range rgList {
			if ptr.MatchString(ngram[i]) {
				matches = true
				break
			}
		}
		return matches
	})
	return ans
}

func Search(basePath string, args SearchArgs) (*SearchResult, error) {
	fullPath := filepath.Join(basePath, args.CorpusID)
	gindex := index.LoadNgramIndex(fullPath, args.Attrs)
	wd, err := wdict.LoadWordDict(fullPath)
	if err != nil {
		return nil, err
	}
	sindex := index.OpenSearchableIndex(gindex, wd)
	var res *index.NgramSearchResult

	if args.QueryType == 1 {
		res = searchByRegexp(wd, sindex, args)

	} else {
		if strings.HasSuffix(args.Phrase, "*") {
			res = searchByPrefix(wd, sindex, args)

		} else {
			res = sindex.GetNgramsOf(args.Phrase)
		}
	}
	if res.Size() >= args.Offset+args.Limit {
		res.Slice(args.Offset, args.Offset+args.Limit)
	}
	ans := &SearchResult{result: res, wdict: wd}
	return ans, nil
}

// ---------------------------------------------------------

type resultRowsResp struct {
	Size       int                 `json:"size"`
	Rows       []*SearchResultItem `json:"rows"`
	SearchTime float64             `json:"searchTime"`
}
