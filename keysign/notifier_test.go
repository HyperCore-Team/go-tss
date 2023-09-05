package keysign

import (
	"github.com/HyperCore-Team/go-tss/messages"
	. "gopkg.in/check.v1"

	"github.com/HyperCore-Team/go-tss/conversion"
)

type NotifierTestSuite struct{}

var _ = Suite(&NotifierTestSuite{})

func (*NotifierTestSuite) SetUpSuite(c *C) {
}

func (NotifierTestSuite) TestNewNotifier(c *C) {
	testMSg := [][]byte{[]byte("hello"), []byte("world")}
	poolPubKey := conversion.GetRandomPubKey()
	n, err := NewNotifier("", testMSg, poolPubKey, messages.ECDSAKEYSIGN)
	c.Assert(err, NotNil)
	c.Assert(n, IsNil)
	n, err = NewNotifier("aasfdasdf", nil, poolPubKey, messages.ECDSAKEYSIGN)
	c.Assert(err, NotNil)
	c.Assert(n, IsNil)

	n, err = NewNotifier("hello", testMSg, "", messages.ECDSAKEYSIGN)
	c.Assert(err, NotNil)
	c.Assert(n, IsNil)

	n, err = NewNotifier("hello", testMSg, poolPubKey, messages.ECDSAKEYSIGN)
	c.Assert(err, IsNil)
	c.Assert(n, NotNil)
	ch := n.GetResponseChannel()
	c.Assert(ch, NotNil)
}
