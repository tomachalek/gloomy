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
	"github.com/tomachalek/gloomy/index/gconf"
	"github.com/tomachalek/gloomy/util"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ServerError interface {
	Error() string
	Code() int
}

type DefaultServerError struct {
	message string
	code    int
}

func (e DefaultServerError) Error() string {
	return e.message
}

func (e DefaultServerError) Code() int {
	return e.code
}

func newServerError(desc interface{}, code int) ServerError {
	switch desc.(type) {
	case error:
		return DefaultServerError{message: desc.(error).Error(), code: code}
	case string:
		return DefaultServerError{message: desc.(string), code: code}
	default:
		return DefaultServerError{message: "Unknow error", code: code}
	}
}

func ImportQueryType(qtype string) int {
	switch qtype {
	case "regexp":
		return 1
	case "default":
		return 0
	}
	return -1
}

// ------------------------------------------------------

func fetchIntArg(args map[string][]string, key string, dflt int) (int, error) {
	v, ok := args[key]
	if ok {
		return strconv.Atoi(v[0])
	}
	return dflt, nil
}

func fetchStringArg(args map[string][]string, key string, dflt string) (string, error) {
	v, ok := args[key]
	if ok {
		return v[0], nil
	}
	return dflt, nil
}

func (s serviceHandler) parsePath(p string) []string {
	return strings.Split(strings.Trim(p, "/"), "/")
}

// ------------------------------------------------------

type serviceHandler struct {
	conf *gconf.SearchConf
}

func (s serviceHandler) route(p []string, args map[string][]string) (interface{}, ServerError) {
	switch p[0] {
	case "search":
		var err1, err2, err3 error
		t1 := time.Now()
		offset, err1 := fetchIntArg(args, "offset", 0)
		limit, err2 := fetchIntArg(args, "limit", -1)
		qtype, err3 := fetchStringArg(args, "qtype", "default")
		if err := util.FirstError(err1, err2, err3); err != nil {
			return nil, newServerError(err, 500)
		}
		queryArgs := SearchArgs{
			CorpusID:  args["corpus"][0],
			Phrase:    args["q"][0],
			QueryType: ImportQueryType(qtype),
			Attrs:     args["attrs"],
			Offset:    offset,
			Limit:     limit,
		}
		res, err := Search(s.conf.DataPath, queryArgs)
		t2 := time.Since(t1)
		if err != nil {
			return nil, newServerError(err, 500)
		}
		rows := make([]*SearchResultItem, res.Size())
		for i := 0; res.HasNext(); i++ {
			rows[i] = res.Next()
		}
		return &resultRowsResp{Size: res.Size(), Rows: rows, SearchTime: t2.Seconds()}, nil

	default:
		return nil, newServerError("Function not found", http.StatusNotFound)
	}
}

func (s serviceHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "application/json")
	values := req.URL.Query()
	ans, procErr := s.route(s.parsePath(req.URL.Path), values)
	if procErr == nil {
		enc := json.NewEncoder(resp)
		err := enc.Encode(ans)
		if err != nil {
			fmt.Fprint(resp, err)
		}

	} else {
		http.Error(resp, http.StatusText(procErr.Code()), procErr.Code())
	}
}

// ------------------------------------------------------

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
