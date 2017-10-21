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

package files

import (
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
)

// NewReader creates a buffered Scanner for
// plain or gzipped/bzipped files for ruther processing.
func NewReader(filePath string) (io.Reader, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	finfo, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !finfo.Mode().IsRegular() {
		return nil, fmt.Errorf("Path %s is not a regular file", filePath)
	}

	var rd io.Reader
	var err2 error
	if strings.HasSuffix(filePath, ".gz") {
		rd, err2 = gzip.NewReader(f)
		if err2 != nil {
			return nil, err2
		}

	} else if strings.HasSuffix(filePath, ".bz2") {
		rd = bzip2.NewReader(f)

	} else {
		rd = f
	}
	return rd, nil
}
