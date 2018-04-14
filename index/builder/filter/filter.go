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
	"log"
	"plugin"
)

type CustomFilter func(words []string, tags []string) bool

func LoadCustomFilter(libPath string, fn string) CustomFilter {
	if libPath != "" && fn != "" {
		p, err := plugin.Open(libPath)
		if err != nil {
			panic(err)
		}
		f, err := p.Lookup(fn)
		if err != nil {
			panic(err)
		}
		return *f.(*CustomFilter)
	}
	log.Print("No custom filter plug-in defined")
	return func(words []string, tags []string) bool {
		return true
	}
}
