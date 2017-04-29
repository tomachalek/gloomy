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
	"encoding/json"
	"fmt"
	"github.com/tomachalek/gloomy/index"
	"github.com/tomachalek/gloomy/index/gconf"
	"github.com/tomachalek/gloomy/wdict"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"
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

func Search(basePath string, corpusId string, phrase string, attrs []string) (*SearchResult, error) {
	fullPath := filepath.Join(basePath, corpusId)
	gindex := index.LoadNgramIndex(fullPath, attrs)
	wd, err := wdict.LoadWordDict(fullPath)
	if err != nil {
		return nil, err
	}
	sindex := index.OpenSearchableIndex(gindex, wd)
	res := sindex.GetNgramsOf(phrase)
	ans := &SearchResult{result: res, wdict: wd}
	return ans, nil
}

// ---------------------------------------------------------

type resultRowsResp struct {
	Size       int                 `json:"size"`
	Rows       []*SearchResultItem `json:"rows"`
	SearchTime float64             `json:"searchTime"`
}

type serviceHandler struct {
	conf *gconf.SearchConf
}

func (s serviceHandler) route(p []string, args map[string][]string) interface{} {
	switch p[0] {
	case "search":
		t1 := time.Now()
		res, err := Search(s.conf.DataPath, args["corpus"][0], args["q"][0], args["attrs"])
		t2 := time.Since(t1)
		if err != nil {
			log.Printf("ERROR: %s", err)
		}
		rows := make([]*SearchResultItem, res.Size())
		for i := 0; res.HasNext(); i++ {
			rows[i] = res.Next()
		}
		return &resultRowsResp{Size: res.Size(), Rows: rows, SearchTime: t2.Seconds()}

	default:
		return nil
	}
}

func (s serviceHandler) parsePath(p string) []string {
	return strings.Split(strings.Trim(p, "/"), "/")
}

func (s serviceHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "application/json")
	values := req.URL.Query()
	ans := s.route(s.parsePath(req.URL.Path), values)
	if ans != nil {
		enc := json.NewEncoder(resp)
		err := enc.Encode(ans)
		if err != nil {
			fmt.Fprint(resp, err)
		}

	} else {
		http.Error(resp, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

// Serve starts a simple HTTP server
func Serve(conf *gconf.SearchConf) {
	h := serviceHandler{conf: conf}
	addr := fmt.Sprintf("%s:%d", conf.ServerAddress, conf.ServerPort)
	s := &http.Server{
		Addr:           addr,
		Handler:        h,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Printf("Listening on %s", addr)
	log.Fatal(s.ListenAndServe())
}
