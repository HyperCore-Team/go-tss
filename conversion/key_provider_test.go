package conversion

import (
	"encoding/base64"
	"testing"

	"github.com/HyperCore-Team/go-tss/messages"

	"github.com/stretchr/testify/assert"
	. "gopkg.in/check.v1"
)

const testPriKey = "OTI4OTdkYzFjMWFhMjU3MDNiMTE4MDM1OTQyY2Y3MDVkOWFhOGIzN2JlOGIwOWIwMTZjYTkxZjNjOTBhYjhlYQ=="

type KeyProviderTestSuite struct{}

var _ = Suite(&KeyProviderTestSuite{})

func TestGetPubKeysFromPeerIDs(t *testing.T) {
	input := []string{
		"16Uiu2HAmBdJRswX94UwYj6VLhh4GeUf9X3SjBRgTqFkeEMLmfk2M",
		"16Uiu2HAkyR9dsFqkj1BqKw8ZHAUU2yur6ZLRJxPTiiVYP5uBMeMG",
	}
	result, err := GetPubKeysFromPeerIDs(input)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	assert.Len(t, result, 2)
	assert.Equal(t, "AvC1l/VLW5u+w2NCbYIG7QRTQ6/uNJGBN9lSDWNcp/Fq", result[0])
	assert.Equal(t, "AjtMcCnMk3W6j/QvHSWO7mxJLpZ3XPxtEw24/q1dP4AL", result[1])
	input1 := append(input, "whatever")
	result, err = GetPubKeysFromPeerIDs(input1)
	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func (*KeyProviderTestSuite) TestGetPriKey(c *C) {
	pk, err := GetPriKey("whatever")
	c.Assert(err, NotNil)
	c.Assert(pk, IsNil)
	input := base64.StdEncoding.EncodeToString([]byte("whatever"))
	pk, err = GetPriKey(input)
	c.Assert(err, NotNil)
	c.Assert(pk, IsNil)
	pk, err = GetPriKey(testPriKey)
	c.Assert(err, IsNil)
	c.Assert(pk, NotNil)
	result, err := GetPriKeyRawBytes(pk)
	c.Assert(err, IsNil)
	c.Assert(result, NotNil)
	c.Assert(result, HasLen, 64)
}

func (KeyProviderTestSuite) TestGetPeerIDs(c *C) {
	pubKeys := []string{
		"D2Ou8kohzWyVESbCOE/yXHmCAaCbB2R1jDWRpECf1JY=", // 12D3KooWArSSkT7VYQPrbp6cLqWUTqQYb1rX77GhTJaUWMYjeVFs
		"8v5YUvEtN8vpNKejH1dmVi4BoEZX+c5EHoqQCXQM/WE=", // 12D3KooWSAumwg2rxzjsgv7LuWM4u3HqfLcnekg2NHc5TsNj5hgC
	}
	peers, err := GetPeerIDs(pubKeys)
	c.Assert(err, IsNil)
	c.Assert(peers, NotNil)
	c.Assert(peers, HasLen, 2)
	c.Assert(peers[0].String(), Equals, "12D3KooWArSSkT7VYQPrbp6cLqWUTqQYb1rX77GhTJaUWMYjeVFs")
	c.Assert(peers[1].String(), Equals, "12D3KooWSAumwg2rxzjsgv7LuWM4u3HqfLcnekg2NHc5TsNj5hgC")
	pubKeys1 := append(pubKeys, "helloworld")
	peers, err = GetPeerIDs(pubKeys1)
	c.Assert(err, NotNil)
	c.Assert(peers, IsNil)
}

func (KeyProviderTestSuite) TestGetPeerIDFromPubKey(c *C) {
	pID, err := GetPeerIDFromPubKey("D2Ou8kohzWyVESbCOE/yXHmCAaCbB2R1jDWRpECf1JY=")
	c.Assert(err, IsNil)
	c.Assert(pID.String(), Equals, "12D3KooWArSSkT7VYQPrbp6cLqWUTqQYb1rX77GhTJaUWMYjeVFs")
	pID1, err := GetPeerIDFromPubKey("whatever")
	c.Assert(err, NotNil)
	c.Assert(pID1.String(), Equals, "")
}

func (KeyProviderTestSuite) TestCheckKeyOnCurve(c *C) {
	_, err := CheckKeyOnCurve("aa", messages.ECDSAKEYGEN)
	c.Assert(err, NotNil)
	_, err = CheckKeyOnCurve("A5USsme4piKC377RzQr9U3k9yvcWe9oynB9XooXd6akm", messages.ECDSAKEYGEN)
	c.Assert(err, IsNil)
}
