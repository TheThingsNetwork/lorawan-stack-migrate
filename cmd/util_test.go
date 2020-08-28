package cmd_test

import (
	"bytes"
	"io"
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

	s, err = it.Next()
	a.So(err, should.Equal, io.EOF)
	s, err = it.Next()
	a.So(err, should.Equal, io.EOF)
}

func TestReaderIterator(t *testing.T) {
	buf := []byte("one\ntwo\nthree")
	it := cmd.NewReaderIterator(bytes.NewBuffer(buf), byte('\n'))
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

	s, err = it.Next()
	a.So(err, should.Equal, io.EOF)
	s, err = it.Next()
	a.So(err, should.Equal, io.EOF)
}
