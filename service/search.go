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
	"github.com/tomachalek/gloomy/wstore"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type SearchResult struct {
	result *index.NgramSearchResult
	wdict  *wstore.WordIndex
}

func (sr *SearchResult) Size() int {
	return sr.result.GetSize()
}

func (sr *SearchResult) HasNext() bool {
	return sr.result.HasNext()
}

func (sr *SearchResult) Next() []string {
	ans := sr.result.Next()
	if ans != nil {
		return sr.wdict.DecodeNgram(ans)
	}
	return nil
}

func Search(basePath string, corpusId string, phrase string) (*SearchResult, error) {
	fullPath := filepath.Join(basePath, corpusId)
	gindex := index.LoadNgramIndex(fullPath)
	wdict, err := wstore.LoadWordDict(fullPath)
	if err != nil {
		return nil, err
	}
	sindex := index.OpenSearchableIndex(gindex, wdict)
	res := sindex.GetNgramsOf(phrase)
	ans := &SearchResult{result: res, wdict: wdict}
	return ans, nil
}

// ---------------------------------------------------------

type resultRowsResp struct {
	Size int        `json:"size"`
	Rows [][]string `json:"rows"`
}

type serviceHandler struct {
	conf *gconf.SearchConf
}

func (s serviceHandler) route(p []string, args map[string][]string) interface{} {
	switch p[0] {
	case "search":
		res, err := Search(s.conf.DataPath, args["corpus"][0], args["q"][0])
		if err != nil {
			log.Printf("ERROR: %s", err)
		}
		rows := make([][]string, res.Size())
		for i := 0; res.HasNext(); i++ {
			rows[i] = res.Next()
		}
		return &resultRowsResp{Size: res.Size(), Rows: rows}

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
	s := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", conf.ServerAddress, conf.ServerPort),
		Handler:        h,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}