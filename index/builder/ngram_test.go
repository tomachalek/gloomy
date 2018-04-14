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

package builder

import (
	"testing"
)

func TestInitialization(t *testing.T) {
	ng := NewStdNgramBuffer(3)
	if ng.begin != 0 {
		t.Errorf("ng.begin != 0, value = %d", ng.begin)
	}
	if ng.write != -1 {
		t.Errorf("ng.write != -1, value = %d", ng.write)
	}
	if len(ng.data) != 3 {
		t.Error("ng.data length != 3")
	}
	if cap(ng.data) != 3 {
		t.Error("ng.data capacity != 3")
	}
}

func TestInsertFirst(t *testing.T) {
	ng := NewStdNgramBuffer(3)
	ng.AddToken("foo")
	if ng.begin != 1 {
		t.Error("ng.begin != 1")
	}
	if ng.write != 0 {
		t.Error("ng.write != 0")
	}
}

func TestInsertedValues(t *testing.T) {
	ng := NewStdNgramBuffer(3)
	ng.AddToken("foo")
	ng.AddToken("bar")
	ng.AddToken("baz")
	v := ng.GetValue()
	if v[0] != "foo" || v[1] != "bar" || v[2] != "baz" {
		t.Errorf("ng.GetValue() != [foo, bar, baz], value: %s", v)
	}
}

func TestInsertedValues2(t *testing.T) {
	ng := NewStdNgramBuffer(3)
	ng.AddToken("foo")
	ng.AddToken("bar")
	ng.AddToken("baz")
	ng.AddToken("dex")
	v := ng.GetValue()
	if v[0] != "bar" || v[1] != "baz" || v[2] != "dex" {
		t.Errorf("ng.GetValue() != [bar, baz, dex], value: %s", v)
	}
}

func TestReset(t *testing.T) {
	ng := NewStdNgramBuffer(3)
	ng.AddToken("foo")
	ng.AddToken("bar")
	ng.AddToken("baz")
	ng.Reset()
	if ng.begin != 0 {
		t.Errorf("ng.begin != 0, value = %d", ng.begin)
	}
	if ng.write != -1 {
		t.Errorf("ng.write != -1, value = %d", ng.write)
	}
	if ng.data[0] != "" || ng.data[1] != "" || ng.data[2] != "" {
		t.Errorf("ng.data != ['', '', ''], value: %s", ng.data)
	}
}
