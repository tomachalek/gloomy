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

// ServerError represents a general HTTP server error
// which is expected to be JSON-exportable
type ServerError interface {
	ToJSON() string
	HTTPCode() int
}

// DefaultServerError is a basic implementation of
// ServerError used here in server.go
type DefaultServerError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// ToJSON converts error to JSON
func (e DefaultServerError) ToJSON() string {
	errAns, _ := json.Marshal(e)
	return string(errAns)
}

// HTTPCode returns a numeric HTTP code of the error
// (typically 40x, 50x)
func (e DefaultServerError) HTTPCode() int {
	return e.Code
}

func newServerError(desc interface{}, code int) ServerError {
	switch desc.(type) {
	case error:
		return DefaultServerError{Message: desc.(error).Error(), Code: code}
	case string:
		return DefaultServerError{Message: desc.(string), Code: code}
	default:
		return DefaultServerError{Message: "Unknow error", Code: code}
	}
}

// ImportQueryType imports end-user encoded query type (default, regexp)
// to internal numeric ones. Returns -1 in case query type is not found.
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

func requireStringArg(args map[string][]string, key string) (string, error) {
	v, ok := args[key]
	if ok && len(v) > 0 {
		return v[0], nil
	}
	return "", fmt.Errorf("Argument '%s' not found", key)
}

// ------------------------------------------------------

type serviceHandler struct {
	appVersion string
	conf       *gconf.SearchConf
}

func (s *serviceHandler) parsePath(p string) []string {
	return strings.Split(strings.Trim(p, "/"), "/")
}

func (s *serviceHandler) actionSearch(p []string, args map[string][]string) (interface{}, ServerError) {
	var err1, err2, err3, err4, err5 error
	t1 := time.Now()
	offset, err1 := fetchIntArg(args, "offset", 0)
	limit, err2 := fetchIntArg(args, "limit", -1)
	qtype, err3 := fetchStringArg(args, "qtype", "default")
	corpusID, err4 := requireStringArg(args, "corpus")
	query, err5 := requireStringArg(args, "q")
	if err := util.FirstError(err1, err2, err3, err4, err5); err != nil {
		return nil, newServerError(err, 500)
	}
	queryArgs := SearchArgs{
		CorpusID:  corpusID,
		Phrase:    query,
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
}

func (s *serviceHandler) actionInfo(p []string, args map[string][]string) (interface{}, ServerError) {
	ans := make(map[string]string)
	ans["name"] = "Gloomy - the n-gram database"
	ans["version"] = s.appVersion
	return ans, nil
}

func (s *serviceHandler) route(path []string, args map[string][]string) (interface{}, ServerError) {
	switch path[0] {
	case "":
		return s.actionInfo(path, args)
	case "search":
		return s.actionSearch(path, args)
	default:
		return nil, newServerError(fmt.Sprintf("Action '%s' not found", path[0]), http.StatusNotFound)
	}
}

func (s *serviceHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			http.Error(resp, fmt.Sprintf("%s", r), 500)
		}
	}()
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
		http.Error(resp, procErr.ToJSON(), procErr.HTTPCode())
	}
}

// ------------------------------------------------------

// Serve starts a simple HTTP server
func Serve(conf *gconf.SearchConf, appVersion string) {
	h := &serviceHandler{
		conf:       conf,
		appVersion: appVersion,
	}
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
