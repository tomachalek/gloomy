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

package index

import (
	"testing"
)

func TestNgramsCmp(t *testing.T) {
	ans := ngramsCmp([]string{"foo", "bar", "baz"}, []string{"foo", "bar", "baz"})
	if ans != 0 {
		t.Errorf("Failed - expected 0, found %d", ans)
	}
}

func TestNgramsCmpFirstBigger(t *testing.T) {
	ans := ngramsCmp([]string{"zzz", "bar", "baz"}, []string{"foo", "bar", "baz"})
	if ans != 1 {
		t.Errorf("Failed - expected 1, found %d", ans)
	}
}

func TestNgramsCmpFirstBigger2(t *testing.T) {
	ans := ngramsCmp([]string{"foo", "bar", "baz"}, []string{"foo", "bar", "bay"})
	if ans != 1 {
		t.Errorf("Failed - expected 1, found %d", ans)
	}
}

func TestNgramsCmpFirstSmaller(t *testing.T) {
	ans := ngramsCmp([]string{"eon", "bar", "baz"}, []string{"foo", "bar", "baz"})
	if ans != -1 {
		t.Errorf("Failed - expected -1, found %d", ans)
	}
}

func TestEmptyNgramsCmp(t *testing.T) {
	ans := ngramsCmp([]string{}, []string{})
	if ans != 0 {
		t.Errorf("Failed - expected 0, found %d", ans)
	}
}

// --------

func TestNgramListAdd(t *testing.T) {
	nl := SortedNgramList{}
	v := []string{"foo", "bar"}
	nl.Add(v)
	if ngramsCmp(nl.firstNode.ngram, v) != 0 {
		t.Errorf("Failed - expected first ngram to be %s, found: %s", v, nl.firstNode.ngram)
	}
}

func TestNgramListAddMulti(t *testing.T) {
	n := SortedNgramList{}
	v1 := []string{"foo", "bar"}
	n.Add(v1)
	v2 := []string{"boo", "bar"}
	n.Add(v2)
	v3 := []string{"eon", "bar"}
	n.Add(v3)

	if ngramsCmp(n.firstNode.ngram, v2) != 0 {
		t.Errorf("Failed - expected first ngram to be %s, found: %s", v2, n.firstNode.ngram)
	}
	if ngramsCmp(n.firstNode.next.ngram, v3) != 0 {
		t.Errorf("Failed - expected second ngram to be %s, found: %s", v3, n.firstNode.next.ngram)
	}
	if ngramsCmp(n.firstNode.next.next.ngram, v1) != 0 {
		t.Errorf("Failed - expected second ngram to be %s, found: %s", v1, n.firstNode.next.next.ngram)
	}
}
