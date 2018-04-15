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

package filter

import (
	"fmt"
	"log"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/tomachalek/gloomy/util"
)

const (
	defaultSystemPluginDir = "/usr/local/lib/gloomy"
)

// CustomFilter compiled as a plug-in can be used to
// filter out specific n-grams from a created index
// (out output file).
type CustomFilter func(words []string, tags []string) bool

func findPluginLib(pathSuff string) (string, error) {
	paths := []string{
		pathSuff,
		filepath.Join(util.GetWorkingDir(), pathSuff),
		filepath.Join(defaultSystemPluginDir, pathSuff),
	}
	for _, fullPath := range paths {
		if util.IsFile(fullPath) {
			return fullPath, nil
		}
	}
	return "", fmt.Errorf("Failed to find plug-in file in %s", strings.Join(paths, ", "))
}

// LoadCustomFilter loads a compiled .so plugin from a defined
// path and selects a function identified by fn.
// In case libPath does not point to an existing file, the function
// handles it as a path suffix and tries other locations (working
// directory, /usr/local/lib/gloomy).
func LoadCustomFilter(libPath string, fn string) CustomFilter {
	if libPath != "" && fn != "" {
		fullPath, err := findPluginLib(libPath)
		if err != nil {
			panic(err)
		}
		p, err := plugin.Open(fullPath)
		if err != nil {
			panic(err)
		}
		f, err := p.Lookup(fn)
		if err != nil {
			panic(err)
		}
		log.Printf("Using filter plug-in %s from %s", fn, fullPath)
		return *f.(*CustomFilter)
	}
	log.Print("No custom filter plug-in defined")
	return func(words []string, tags []string) bool {
		return true
	}
}
