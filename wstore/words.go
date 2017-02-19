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

package wstore

import (
	"encoding/gob"
	"fmt"
	"os"
)

func SaveWordDict(data []string, outPath string) error {
	f, err := os.OpenFile(outPath, os.O_CREATE, 0)
	if err != nil {
		panic(fmt.Sprintf("Failed to save map: %s", err))
	}
	enc := gob.NewEncoder(f)
	return enc.Encode(data)
}

func LoadWordDict(dataPath string) ([]string, error) {
	f, err := os.Open(dataPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to load map: %s", err))
	}
	dec := gob.NewDecoder(f)
	var ans []string
	return ans, dec.Decode(&ans)
}

func FindIndex(word string, data []string) int {
	left := 0
	right := len(data) - 1
	pivot := len(data) / 2
	for left < right && data[pivot] != word {
		if data[left] <= word && word <= data[pivot] {
			tmp := pivot
			pivot = (left + pivot) / 2
			right = tmp

		} else if data[pivot] < word && word <= data[right] {
			tmp := pivot
			pivot = (pivot + right) / 2
			left = tmp

		} else {
			// TODO Not found
		}
	}
	if word == data[pivot] {
		return pivot
	}
	return -1
}
