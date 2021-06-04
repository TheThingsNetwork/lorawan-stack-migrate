package cmd_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack-migrate/cmd"
)

func TestListIterator(t *testing.T) {
	it := cmd.NewListIterator([]string{"one", "two", "three"})
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
		it := cmd.NewReaderIterator(bytes.NewBuffer(buf), '\n')
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
