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

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tomachalek/gloomy/index/builder"
	"github.com/tomachalek/gloomy/index/extras"
	"github.com/tomachalek/gloomy/index/gconf"
	"github.com/tomachalek/gloomy/service"
)

const (
	createIndexAction   = "create-index"
	extractNgramsAction = "extract-ngrams"
	searchServiceAction = "search-service"
	searchAction        = "search"
	appVersion          = "0.1.0"
)

func help(topic string) {
	if topic == "" {
		fmt.Print("Missing action to help with. Select one of the:\n\tcreate-index, extract-ngrams, search-service, search")
	}
	fmt.Printf("HELP on [%s]:\n", topic)
}

func createIndex(conf *gconf.IndexBuilderConf, ngramSize int) {
	if conf.InputFilePath == "" {
		fmt.Println("Vertical file not specified")
		os.Exit(1)
	}
	log.Println("Importing vertical file ", conf.InputFilePath)
	if conf.OutDirectory == "" {
		conf.OutDirectory = filepath.Dir(conf.InputFilePath)
	}
	fmt.Println("Output directory: ", conf.OutDirectory)
	t0 := time.Now()
	builder.CreateGloomyIndex(conf, ngramSize)
	fmt.Printf("DONE in %s\n", time.Since(t0))
}

func extractNgrams(conf *gconf.IndexBuilderConf, ngramSize int) {
	if conf.InputFilePath == "" {
		fmt.Println("Vertical file not specified")
		os.Exit(1)
	}
	log.Println("Processing vertical file ", conf.InputFilePath)
	if conf.OutDirectory == "" {
		conf.OutDirectory = filepath.Dir(conf.InputFilePath)
	}
	fmt.Println("Output directory: ", conf.OutDirectory)
	t0 := time.Now()
	extras.ExtractUniqueNgrams(conf, ngramSize)
	fmt.Printf("DONE in %s\n", time.Since(t0))
}

func loadSearchConf(confBasePath string) *gconf.SearchConf {
	if confBasePath == "" {
		var err error
		confBasePath, err = os.Getwd()
		if err != nil {
			panic(err)
		}
		confBasePath = filepath.Join(confBasePath, "gloomy.json")
	}
	return gconf.LoadSearchConf(confBasePath)
}

func searchCLI(confBasePath string, corpus string, query string, attrs []string, offset int, limit int, queryType int) {
	conf := loadSearchConf(confBasePath)
	t1 := time.Now()
	args := service.SearchArgs{
		CorpusID:  corpus,
		Phrase:    query,
		QueryType: queryType,
		Attrs:     attrs,
		Offset:    offset,
		Limit:     limit,
	}
	ans, err := service.Search(conf.DataPath, args)
	if err != nil {
		log.Printf("Srch error: %s", err)
	}
	t2 := time.Since(t1)
	for i := 0; ans.HasNext(); i++ {
		v := ans.Next()
		log.Printf("res[%d]: %s (count: %d, meta: %s)", i, v.Ngram, v.Count, v.Args)
	}
	log.Printf("Search time: %s", t2)
}

func startSearchService(confBasePath string) {
	conf := loadSearchConf(confBasePath)
	service.Serve(conf, appVersion)
}

func parseAttrs(attrStr string) []string {
	if len(attrStr) == 0 {
		return []string{}
	}
	return strings.Split(attrStr, ",")
}

func main() {
	ngramSize := flag.Int("ngram-size", 2, "N-gram size, 2: bigram (default), ...")
	srchConfPath := flag.String("conf-path", "", "Path to the gloomy.conf (by default, working dir is used")
	metadataAttrs := flag.String("attrs", "", "Metadata attributes separated by comma")
	resultLimit := flag.Int("limit", -1, "Result limit")
	resultOffset := flag.Int("offset", 0, "Result offset (starting from zero)")
	queryType := flag.String("qtype", "default", "Query type (0 = default, 1 = regexp)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Gloomy - an n-gram database >>>\n\nUsage:\n\t%s [options] [action] [config.json]\n\nAavailable actions:\n\tsearch, search-service, create-index, extract-ngrams\n\nOptions:\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(flag.Args()) == 0 {
		fmt.Println("Missing action, try -h for help")
		os.Exit(1)

	} else {
		switch flag.Arg(0) {
		case "help":
			help(flag.Arg(1))
		case createIndexAction:
			conf := gconf.LoadIndexBuilderConf(flag.Arg(1))
			createIndex(conf, *ngramSize)
		case extractNgramsAction:
			conf := gconf.LoadIndexBuilderConf(flag.Arg(1))
			extractNgrams(conf, *ngramSize)
		case searchServiceAction:
			startSearchService(*srchConfPath)
		case searchAction:
			if flag.Arg(1) == "" || flag.Arg(2) == "" {
				log.Fatal("Missing argument (both corpus and query must be specified)")
			}
			qtype := service.ImportQueryType(*queryType)
			if qtype < 0 {
				panic(fmt.Sprintf("Unknown query type: %s", *queryType))
			}
			searchCLI(*srchConfPath, flag.Arg(1), flag.Arg(2), parseAttrs(*metadataAttrs),
				*resultOffset, *resultLimit, qtype)
		default:
			fmt.Printf("Unknown action %s\n", flag.Arg(0))
			os.Exit(1)
		}
	}
}
