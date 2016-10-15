package lib

import (
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

func init() {
	Suite(&S{})
}

type S struct{}

// func (s *S) TestSimpleChecks(c *C) {
// 	c.Assert(value, Equals, 42)
// 	c.Assert(s, Matches, "hel.*there")
// 	c.Assert(err, IsNil)
// 	c.Assert(foo, Equals, bar, Commentf("#CPUs == %d", runtime.NumCPU()))
// }

func (s *S) TestBlobSpec(c *C) {
	// c.Assert(42, Equals, "42")
	// c.Assert(io.ErrClosedPipe, ErrorMatches, "io: .*on closed pipe")
	// c.Check(42, Equals, 42)

	var bs *BlobSpec
	var err error

	bs, err = ParseBlobSpec("")
	c.Assert(err, IsNil)
	c.Assert(bs.PathPresent, Equals, false)
	c.Assert(bs.Path, Equals, "")
	c.Assert(bs.Container, Equals, "")

	bs, err = ParseBlobSpec("foo")
	c.Assert(err, IsNil)
	c.Assert(bs.PathPresent, Equals, false)
	c.Assert(bs.Path, Equals, "")
	c.Assert(bs.Container, Equals, "foo")

	bs, err = ParseBlobSpec("foo/")
	c.Assert(err, IsNil)
	c.Assert(bs.PathPresent, Equals, true)
	c.Assert(bs.Path, Equals, "")
	c.Assert(bs.Container, Equals, "foo")

	bs, err = ParseBlobSpec("foo/bar")
	c.Assert(err, IsNil)
	c.Assert(bs.PathPresent, Equals, true)
	c.Assert(bs.Path, Equals, "bar")
	c.Assert(bs.Container, Equals, "foo")
}
