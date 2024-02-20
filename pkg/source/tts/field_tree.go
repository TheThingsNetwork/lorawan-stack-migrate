// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tts

import (
	"slices"
	"strings"
)

// FieldTree is a tree of fields.
type FieldTree map[string]FieldTree

// Add adds the paths to the field tree.
func (n FieldTree) Add(paths ...string) FieldTree {
	for _, path := range paths {
		n := n
		components := strings.Split(path, ".")
		for _, component := range components {
			if n[component] == nil {
				n[component] = make(FieldTree)
			}
			n = n[component]
		}
	}
	return n
}

// Same returns true if the field tree is the same as the other field tree.
func (n FieldTree) Same(other FieldTree) bool {
	if len(n) != len(other) {
		return false
	}
	for k, v := range n {
		if !v.Same(other[k]) {
			return false
		}
	}
	return true
}

// Compress returns a new field tree that is the result of compressing the field tree with the reference field tree.
func (n FieldTree) Compress(reference FieldTree) FieldTree {
	result := make(FieldTree)
	for k, v := range n {
		sub, ok := reference[k]
		if !ok {
			panic("incomplete reference tree")
		}
		if sub.Same(v) {
			result[k] = make(FieldTree)
		} else {
			result[k] = v.Compress(sub)
		}
	}
	return result
}

func (n FieldTree) fields(prefix string) []string {
	var result []string
	for k, v := range n {
		if len(v) == 0 {
			result = append(result, prefix+k)
		} else {
			result = append(result, v.fields(prefix+k+".")...)
		}
	}
	return result
}

// Fields returns the fields of the field tree.
func (n FieldTree) Fields() []string {
	fields := n.fields("")
	slices.Sort(fields)
	return fields
}

// CompressFields compresses the fields with the reference field tree.
func CompressFields(fields []string, tree FieldTree) []string {
	return make(FieldTree).Add(fields...).Compress(tree).Fields()
}
