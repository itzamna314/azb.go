package lib

import (
	. "gopkg.in/check.v1"
)

// func (s *S) TestSimpleChecks(c *C) {
// 	c.Assert(value, Equals, 42)
// 	c.Assert(s, Matches, "hel.*there")
// 	c.Assert(err, IsNil)
// 	c.Assert(foo, Equals, bar, Commentf("#CPUs == %d", runtime.NumCPU()))
// }

func (x *S) TestStack(c *C) {
	stack := &Stack{}

	c.Assert(stack.Len(), Equals, 0)

	s, ok := stack.Pop()
	c.Assert(ok, Equals, false)

	s, ok = stack.Peek()
	c.Assert(ok, Equals, false)
	c.Assert(stack.String(), Equals, "")

	stack.Push("foo")
	c.Assert(stack.Len(), Equals, 1)

	s, ok = stack.Peek()
	c.Assert(ok, Equals, true)
	c.Assert(s, Equals, "foo")
	c.Assert(stack.String(), Equals, "{foo}")

	s, ok = stack.Pop()
	c.Assert(stack.Len(), Equals, 0)
	c.Assert(ok, Equals, true)
	c.Assert(s, Equals, "foo")

	stack.Push("foo", "bar")
	c.Assert(stack.Len(), Equals, 2)
	c.Assert(stack.String(), Equals, "{foo} {bar}")

	s, ok = stack.Pop()
	c.Assert(stack.Len(), Equals, 1)
	c.Assert(ok, Equals, true)
	c.Assert(s, Equals, "bar")

	s, ok = stack.Pop()
	c.Assert(stack.Len(), Equals, 0)
	c.Assert(ok, Equals, true)
	c.Assert(s, Equals, "foo")

	stack.Push("foo", "bar")
	c.Assert(stack.Len(), Equals, 2)
	arr := stack.Array()
	c.Assert(arr, NotNil)
	c.Assert(len(arr), Equals, 2)
	c.Assert(arr[0], Equals, "bar")
	c.Assert(arr[1], Equals, "foo")
}
