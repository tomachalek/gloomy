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
	"time"

	"github.com/tomachalek/gloomy/index"
	"github.com/tomachalek/gloomy/tools"
	"github.com/tomachalek/gloomy/vertical"
)

const (
	createIndexAction   = "create-index"
	extractNgramsAction = "extract-ngrams"
)

func help() {
	fmt.Println("HELP:")
}

func createIndex(conf *vertical.ParserConf, ngramSize int) {
	if conf.VerticalFilePath == "" {
		fmt.Println("Vertical file not specified")
		os.Exit(1)
	}
	log.Println("Importing vertical file ", conf.VerticalFilePath)
	if conf.OutDirectory == "" {
		conf.OutDirectory = filepath.Dir(conf.VerticalFilePath)
	}
	fmt.Println("Output directory: ", conf.OutDirectory)
	t0 := time.Now()
	index.CreateGloomyIndex(conf, ngramSize)
	fmt.Printf("DONE in %s\n", time.Since(t0))
}

func extractNgrams(conf *vertical.ParserConf, ngramSize int) {
	if conf.VerticalFilePath == "" {
		fmt.Println("Vertical file not specified")
		os.Exit(1)
	}
	log.Println("Processing vertical file ", conf.VerticalFilePath)
	if conf.OutDirectory == "" {
		conf.OutDirectory = filepath.Dir(conf.VerticalFilePath)
	}
	fmt.Println("Output directory: ", conf.OutDirectory)
	t0 := time.Now()
	tools.ExtractNgrams(conf, ngramSize)
	fmt.Printf("DONE in %s\n", time.Since(t0))
}

func main() {
	ngramSize := flag.Int("ngram-size", 2, "N-gram size, 2: bigram (default), ...")
	flag.Parse()
	if len(flag.Args()) == 0 {
		fmt.Println("Missing action, try -h for help")
		os.Exit(1)

	} else {
		switch flag.Arg(0) {
		case "help":
			help()
		case createIndexAction:
			conf := vertical.LoadConfig(flag.Arg(1))
			createIndex(conf, *ngramSize)
		case extractNgramsAction:
			conf := vertical.LoadConfig(flag.Arg(1))
			extractNgrams(conf, *ngramSize)
		default:
			panic("Unknown action")
		}
	}
}
