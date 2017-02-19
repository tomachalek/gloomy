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
	"github.com/tomachalek/gloomy/index"
	"github.com/tomachalek/gloomy/vertical"
	"os"
	"path/filepath"
)

func help() {
	fmt.Println("HELP:")
}

func importVertical(conf *vertical.ParserConf) {
	if conf.VerticalFilePath == "" {
		fmt.Println("Vertical file not specified")
		os.Exit(1)
	}
	fmt.Println("Importing vertical file...", conf.VerticalFilePath)
	if conf.OutDirectory == "" {
		conf.OutDirectory = filepath.Dir(conf.VerticalFilePath)
	}
	fmt.Println("Output directory: ", conf.OutDirectory)
	index.ProcessVertical(conf)

}

func main() {
	flag.Parse()
	if len(flag.Args()) == 0 {
		fmt.Println("Missing action, try -h for help")
		os.Exit(1)

	} else {
		switch flag.Arg(0) {
		case "help":
			help()
		case "import-vertical":
			conf := vertical.LoadConfig(flag.Arg(1))
			importVertical(conf)
		case "import-words":
			// TODO
			break
		default:
			panic("Unknown action")
		}
	}
}
