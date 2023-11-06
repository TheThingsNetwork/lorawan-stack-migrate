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
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestListIterator(t *testing.T) {
	it := NewListIterator([]string{"one", "two", "three"})
	a := assertions.New(t)

	s, err := it.Next()
	a.So(s, should.Equal, "one")
	a.So(err, should.BeNil)

	s, err = it.Next()
	a.So(s, should.Equal, "two")
	a.So(err, should.BeNil)

	s, err = it.Next()
	a.So(s, should.Equal, "three")
	a.So(err, should.BeNil)

	_, err = it.Next()
	a.So(err, should.Equal, io.EOF)
	_, err = it.Next()
	a.So(err, should.Equal, io.EOF)
}

func TestReaderIterator(t *testing.T) {
	for _, sep := range []string{"\n", "\r\n"} {
		buf := []byte(strings.Join([]string{"one", "two", "three"}, sep))
		it := NewReaderIterator(bytes.NewBuffer(buf), '\n')
		a := assertions.New(t)

		s, err := it.Next()
		a.So(s, should.Equal, "one")
		a.So(err, should.BeNil)

		s, err = it.Next()
		a.So(s, should.Equal, "two")
		a.So(err, should.BeNil)

		s, err = it.Next()
		a.So(s, should.Equal, "three")
		a.So(err, should.BeNil)

		_, err = it.Next()
		a.So(err, should.Equal, io.EOF)
		_, err = it.Next()
		a.So(err, should.Equal, io.EOF)
	}
}
