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

package vertical

type StructAttrs struct {
	elms map[string]*VerticalMetaLine
}

func (sa *StructAttrs) Begin(v *VerticalMetaLine) {
	sa.elms[v.Name] = v
}

func (sa *StructAttrs) End(name string) *VerticalMetaLine {
	tmp := sa.elms[name]
	delete(sa.elms, name)
	return tmp
}

func (sa *StructAttrs) GetAttrs() map[string]string {
	ans := make(map[string]string)
	for k, v := range sa.elms {
		for k2, v2 := range v.Attrs {
			ans[k+"."+k2] = v2
		}
	}
	return ans
}

func (sa *StructAttrs) Size() int {
	return len(sa.elms)
}

func NewStructAttrs() *StructAttrs {
	return &StructAttrs{elms: make(map[string]*VerticalMetaLine)}
}
