// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package iterator

import (
	"bufio"
	"io"
	"strings"
)

// Iterator returns items
type Iterator interface {
	// Next returns the next item from the iterator. io.EOF is returned when no more items are left.
	Next() (item string, err error)
}

type listIterator struct {
	items []string
	index int
}

type readerIterator struct {
	rd  *bufio.Reader
	sep byte
}

// NewListIterator returns a new iterator from a list of items.
func NewListIterator(items []string) Iterator {
	return &listIterator{items: items}
}

func (l *listIterator) Next() (string, error) {
	if l.index < len(l.items) {
		l.index++
		return l.items[l.index-1], nil
	}
	return "", io.EOF
}

// NewReaderIterator returns a new iterator from a reader.
func NewReaderIterator(rd io.Reader, sep byte) Iterator {
	return &readerIterator{rd: bufio.NewReader(rd), sep: sep}
}

func (r *readerIterator) Next() (string, error) {
	s, err := r.rd.ReadString(r.sep)
	if err == io.EOF && s != "" {
		return s, nil
	}
	return strings.TrimSpace(s), err
}

// noopIterator is a no-op iterator.
type noopIterator struct {
}

// NewNoopIterator returns a new no-op iterator.
func NewNoopIterator() Iterator {
	return &noopIterator{}
}

// Next implements Iterator
func (n *noopIterator) Next() (string, error) {
	return "", io.EOF
}
