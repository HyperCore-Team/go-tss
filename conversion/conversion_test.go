package conversion

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"sort"
	"testing"

	"github.com/binance-chain/tss-lib/tss"

	"github.com/binance-chain/tss-lib/crypto"
	"github.com/btcsuite/btcd/btcec"
	coskey "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/libp2p/go-libp2p-core/peer"
	. "gopkg.in/check.v1"
)

var (
	testPubKeys = [...]string{
		"D2Ou8kohzWyVESbCOE/yXHmCAaCbB2R1jDWRpECf1JY=", // 12D3KooWArSSkT7VYQPrbp6cLqWUTqQYb1rX77GhTJaUWMYjeVFs
		"8v5YUvEtN8vpNKejH1dmVi4BoEZX+c5EHoqQCXQM/WE=", // 3
		"Zlgbrnmk6xDkamTs004bZgUYbpiE5dV4rSg+MfSk4gU=", // 2
		"jzTMn5m27Cmt6EuCAuKnIzxNbVYY4EIywP0a9grmSok=", // 1
	}
	testPeers = []string{
		"12D3KooWArSSkT7VYQPrbp6cLqWUTqQYb1rX77GhTJaUWMYjeVFs",
		"12D3KooWSAumwg2rxzjsgv7LuWM4u3HqfLcnekg2NHc5TsNj5hgC",
		"12D3KooWGhsgEZA8Nc6xnYRLJxZns8KWtF6ztF73cDhAKXGiHjAc",
		"12D3KooWKTPAZDwc9VuBs7NrVsR9MrUo3L6HzrqP9grtiAFBTbCL",
	}
)

type ConversionTestSuite struct {
	testPubKeys []string
	localPeerID peer.ID
}

var _ = Suite(&ConversionTestSuite{})

func (p *ConversionTestSuite) SetUpTest(c *C) {
	var err error
	p.testPubKeys = testPubKeys[:]
	sort.Strings(p.testPubKeys)
	p.localPeerID, err = peer.Decode("12D3KooWSAumwg2rxzjsgv7LuWM4u3HqfLcnekg2NHc5TsNj5hgC")
	c.Assert(err, IsNil)
}
func TestPackage(t *testing.T) { TestingT(t) }

func (p *ConversionTestSuite) TestAccPubKeysFromPartyIDs(c *C) {
	partiesID, _, err := GetParties(p.testPubKeys, p.testPubKeys[0], false, "")
	c.Assert(err, IsNil)
	partyIDMap := SetupPartyIDMap(partiesID)
	var keys []string
	for k := range partyIDMap {
		keys = append(keys, k)
	}

	got, err := AccPubKeysFromPartyIDs(keys, partyIDMap)
	c.Assert(err, IsNil)
	sort.Strings(got)
	c.Assert(got, DeepEquals, p.testPubKeys)
	got, err = AccPubKeysFromPartyIDs(nil, partyIDMap)
	c.Assert(err, Equals, nil)
	c.Assert(len(got), Equals, 0)
}

func (p *ConversionTestSuite) TestGetParties(c *C) {
	partiesID, localParty, err := GetParties(p.testPubKeys, p.testPubKeys[0], false, "")
	c.Assert(err, IsNil)
	pk := coskey.PubKey{
		Key: localParty.Key[:],
	}
	c.Assert(err, IsNil)
	got := base64.StdEncoding.EncodeToString(pk.Bytes())
	c.Assert(got, Equals, p.testPubKeys[0])
	var gotKeys []string
	for _, val := range partiesID {
		pk := coskey.PubKey{
			Key: val.Key,
		}
		got := base64.StdEncoding.EncodeToString(pk.Bytes())
		gotKeys = append(gotKeys, got)
	}
	sort.Strings(gotKeys)
	c.Assert(gotKeys, DeepEquals, p.testPubKeys)

	_, _, err = GetParties(p.testPubKeys, "", true, "")
	c.Assert(err, NotNil)
	_, _, err = GetParties(p.testPubKeys, "12", true, "")
	c.Assert(err, NotNil)
	_, _, err = GetParties(nil, "12", true, "")
	c.Assert(err, NotNil)
}

//
func (p *ConversionTestSuite) TestGetPeerIDFromPartyID(c *C) {
	_, localParty, err := GetParties(p.testPubKeys, p.testPubKeys[0], true, "")
	c.Assert(err, IsNil)
	peerID, err := GetPeerIDFromPartyID(localParty)
	c.Assert(err, IsNil)
	c.Assert(peerID, Equals, p.localPeerID)
	_, err = GetPeerIDFromPartyID(nil)
	c.Assert(err, NotNil)
	localParty.Index = -1
	_, err = GetPeerIDFromPartyID(localParty)
	c.Assert(err, NotNil)
}

func (p *ConversionTestSuite) TestGetPeerIDFromSecp256PubKey(c *C) {
	_, localParty, err := GetParties(p.testPubKeys, p.testPubKeys[0], true, "")
	c.Assert(err, IsNil)
	got, err := GetPeerIDFromEDDSAPubKey(localParty.Key[:])
	c.Assert(err, IsNil)
	c.Assert(got, Equals, p.localPeerID)
	_, err = GetPeerIDFromEDDSAPubKey(nil)
	c.Assert(err, NotNil)
}

func (p *ConversionTestSuite) TestGetPeersID(c *C) {
	localTestPubKeys := testPubKeys[:]
	sort.Strings(localTestPubKeys)
	partiesID, _, err := GetParties(p.testPubKeys, p.testPubKeys[0], true, "")
	c.Assert(err, IsNil)
	partyIDMap := SetupPartyIDMap(partiesID)
	partyIDtoP2PID := make(map[string]peer.ID)
	err = SetupIDMaps(partyIDMap, partyIDtoP2PID)
	c.Assert(err, IsNil)
	retPeers := GetPeersID(partyIDtoP2PID, p.localPeerID.String())
	var expectedPeers []string
	var gotPeers []string
	counter := 0
	for _, el := range testPeers {
		if el == p.localPeerID.String() {
			continue
		}
		expectedPeers = append(expectedPeers, el)
		gotPeers = append(gotPeers, retPeers[counter].String())
		counter++
	}
	sort.Strings(expectedPeers)
	sort.Strings(gotPeers)
	c.Assert(gotPeers, DeepEquals, expectedPeers)

	retPeers = GetPeersID(partyIDtoP2PID, "123")
	c.Assert(len(retPeers), Equals, 4)
	retPeers = GetPeersID(nil, "123")
	c.Assert(len(retPeers), Equals, 0)
}

func (p *ConversionTestSuite) TestPartyIDtoPubKey(c *C) {
	_, localParty, err := GetParties(p.testPubKeys, p.testPubKeys[0], true, "")
	c.Assert(err, IsNil)
	got, err := PartyIDtoPubKey(localParty)
	c.Assert(err, IsNil)
	c.Assert(got, Equals, p.testPubKeys[0])
	_, err = PartyIDtoPubKey(nil)
	c.Assert(err, NotNil)
	localParty.Index = -1
	_, err = PartyIDtoPubKey(nil)
	c.Assert(err, NotNil)
}

func (p *ConversionTestSuite) TestSetupIDMaps(c *C) {
	localTestPubKeys := testPubKeys[:]
	sort.Strings(localTestPubKeys)
	partiesID, _, err := GetParties(p.testPubKeys, p.testPubKeys[0], true, "")
	c.Assert(err, IsNil)
	partyIDMap := SetupPartyIDMap(partiesID)
	partyIDtoP2PID := make(map[string]peer.ID)
	err = SetupIDMaps(partyIDMap, partyIDtoP2PID)
	c.Assert(err, IsNil)
	var got []string

	for _, val := range partyIDtoP2PID {
		got = append(got, val.String())
	}
	sort.Strings(got)
	sort.Strings(testPeers)
	c.Assert(got, DeepEquals, testPeers)
	emptyPartyIDtoP2PID := make(map[string]peer.ID)
	SetupIDMaps(nil, emptyPartyIDtoP2PID)
	c.Assert(emptyPartyIDtoP2PID, HasLen, 0)
}

func (p *ConversionTestSuite) TestSetupPartyIDMap(c *C) {
	localTestPubKeys := testPubKeys[:]
	sort.Strings(localTestPubKeys)
	partiesID, _, err := GetParties(p.testPubKeys, p.testPubKeys[0], true, "")
	c.Assert(err, IsNil)
	partyIDMap := SetupPartyIDMap(partiesID)
	var pubKeys []string
	for _, el := range partyIDMap {
		pk := coskey.PubKey{
			Key: el.Key,
		}
		got := base64.StdEncoding.EncodeToString(pk.Bytes())
		pubKeys = append(pubKeys, got)
	}
	sort.Strings(pubKeys)
	c.Assert(p.testPubKeys, DeepEquals, pubKeys)

	ret := SetupPartyIDMap(nil)
	c.Assert(ret, HasLen, 0)
}

func (p *ConversionTestSuite) TestTssPubKey(c *C) {
	sk, err := btcec.NewPrivateKey(tss.Edwards())
	c.Assert(err, IsNil)
	point, err := crypto.NewECPoint(tss.Edwards(), sk.X, sk.Y)
	c.Assert(err, IsNil)
	_, err = GetTssPubKeyEDDSA(point)
	c.Assert(err, IsNil)

	// create an invalid point
	invalidPoint := crypto.NewECPointNoCurveCheck(tss.Edwards(), sk.X, new(big.Int).Add(sk.Y, big.NewInt(1)))
	_, err = GetTssPubKeyEDDSA(invalidPoint)
	c.Assert(err, NotNil)

	pk, err := GetTssPubKeyEDDSA(nil)
	c.Assert(err, NotNil)
	c.Assert(pk, Equals, "")
	// var point crypto.ECPoint
	c.Assert(json.Unmarshal([]byte(`{"Coords":[70074650318631491136896111706876206496089700125696166275258483716815143842813,72125378038650252881868972131323661098816214918201601489154946637636730727892]}`), &point), IsNil)
	pk, err = GetTssPubKeyECDSA(point)
	c.Assert(err, IsNil)
	c.Assert(pk, Equals, "Aprs2Lew/mkbrR/2QqR1UgBAaeooH3tltdfjzXSpMDP9")
}
