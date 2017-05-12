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
	"github.com/tomachalek/gloomy/index"
	"github.com/tomachalek/gloomy/wdict"
	"path/filepath"
	"strings"
)

type SearchResultItem struct {
	Ngram []string `json:"ngram"`
	Count int      `json:"count"`
	Args  []string `json:"args"`
}

// ---------------------------------------------------------------

type SearchResult struct {
	result *index.NgramSearchResult
	wdict  *wdict.WordDictReader
}

func (sr *SearchResult) Size() int {
	return sr.result.GetSize()
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

func Search(basePath string, corpusId string, phrase string, attrs []string, offset int, limit int) (*SearchResult, error) {
	fullPath := filepath.Join(basePath, corpusId)
	gindex := index.LoadNgramIndex(fullPath, attrs)
	wd, err := wdict.LoadWordDict(fullPath)
	if err != nil {
		return nil, err
	}
	sindex := index.OpenSearchableIndex(gindex, wd)

	var res *index.NgramSearchResult
	if strings.HasSuffix(phrase, "*") {
		indices := wd.FindByPrefix(phrase[:len(phrase)-1])
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
			if res.GetSize() >= offset+limit {
				res.Slice(offset, limit)
			}
		}
		close(ch)

	} else {
		res = sindex.GetNgramsOf(phrase)
		if res.GetSize() >= offset+limit {
			res.Slice(offset, limit)
		}

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
