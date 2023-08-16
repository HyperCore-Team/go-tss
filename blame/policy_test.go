package blame

import (
	"sort"
	"sync"
	"testing"

	bkg "github.com/binance-chain/tss-lib/eddsa/keygen"
	btss "github.com/binance-chain/tss-lib/tss"
	"github.com/libp2p/go-libp2p-core/peer"
	. "gopkg.in/check.v1"

	"github.com/HyperCore-Team/go-tss/conversion"
	"github.com/HyperCore-Team/go-tss/messages"
)

var (
	testPubKeys = [...]string{
		"8v5YUvEtN8vpNKejH1dmVi4BoEZX+c5EHoqQCXQM/WE=", // 3
		"D2Ou8kohzWyVESbCOE/yXHmCAaCbB2R1jDWRpECf1JY=", // 12D3KooWArSSkT7VYQPrbp6cLqWUTqQYb1rX77GhTJaUWMYjeVFs
		"Zlgbrnmk6xDkamTs004bZgUYbpiE5dV4rSg+MfSk4gU=", // 2
		"jzTMn5m27Cmt6EuCAuKnIzxNbVYY4EIywP0a9grmSok=", // 1
	}

	testPeers = []string{
		"12D3KooWGhsgEZA8Nc6xnYRLJxZns8KWtF6ztF73cDhAKXGiHjAc",
		"12D3KooWArSSkT7VYQPrbp6cLqWUTqQYb1rX77GhTJaUWMYjeVFs",
		"12D3KooWSAumwg2rxzjsgv7LuWM4u3HqfLcnekg2NHc5TsNj5hgC",
		"12D3KooWKTPAZDwc9VuBs7NrVsR9MrUo3L6HzrqP9grtiAFBTbCL",
	}
)

func TestPackage(t *testing.T) { TestingT(t) }

type policyTestSuite struct {
	blameMgr *Manager
}

var _ = Suite(&policyTestSuite{})

func (p *policyTestSuite) SetUpTest(c *C) {
	p.blameMgr = NewBlameManager()
	p1, err := peer.Decode(testPeers[0])
	c.Assert(err, IsNil)
	p2, err := peer.Decode(testPeers[1])
	c.Assert(err, IsNil)
	p3, err := peer.Decode(testPeers[2])
	c.Assert(err, IsNil)
	p.blameMgr.SetLastUnicastPeer(p1, "testType")
	p.blameMgr.SetLastUnicastPeer(p2, "testType")
	p.blameMgr.SetLastUnicastPeer(p3, "testType")
	localTestPubKeys := testPubKeys[:]
	sort.Strings(localTestPubKeys)
	partiesID, localPartyID, err := conversion.GetParties(localTestPubKeys, testPubKeys[0], false, "")
	c.Assert(err, IsNil)
	partyIDMap := conversion.SetupPartyIDMap(partiesID)
	err = conversion.SetupIDMaps(partyIDMap, p.blameMgr.PartyIDtoP2PID)
	c.Assert(err, IsNil)
	outCh := make(chan btss.Message, len(partiesID))
	endCh := make(chan bkg.LocalPartySaveData, len(partiesID))
	ctx := btss.NewPeerContext(partiesID)
	params := btss.NewParameters(btss.Edwards(), ctx, localPartyID, len(partiesID), 3)
	keyGenParty := bkg.NewLocalParty(params, outCh, endCh)

	testPartyMap := new(sync.Map)
	testPartyMap.Store("", keyGenParty)
	p.blameMgr.SetPartyInfo(testPartyMap, partyIDMap)
}

func (p *policyTestSuite) TestGetUnicastBlame(c *C) {
	_, err := p.blameMgr.GetUnicastBlame("testTypeWrong")
	c.Assert(err, NotNil)
	_, err = p.blameMgr.GetUnicastBlame("testType")
	c.Assert(err, IsNil)
}

func (p *policyTestSuite) TestGetBroadcastBlame(c *C) {
	pi := p.blameMgr.partyInfo

	r1 := btss.MessageRouting{
		From:                    pi.PartyIDMap["D2Ou8kohzWyVESbCOE/yXHmCAaCbB2R1jDWRpECf1JY="],
		To:                      nil,
		IsBroadcast:             false,
		IsToOldCommittee:        false,
		IsToOldAndNewCommittees: false,
	}
	msg := messages.WireMessage{
		Routing:   &r1,
		RoundInfo: "key1",
		Message:   nil,
	}

	p.blameMgr.roundMgr.Set("key1", &msg)
	blames, err := p.blameMgr.GetBroadcastBlame("key1")
	c.Assert(err, IsNil)
	var blamePubKeys []string
	for _, el := range blames {
		blamePubKeys = append(blamePubKeys, el.Pubkey)
	}
	sort.Strings(blamePubKeys)
	expected := testPubKeys[2:]
	sort.Strings(expected)
	c.Assert(blamePubKeys, DeepEquals, expected)
}

func (p *policyTestSuite) TestTssWrongShareBlame(c *C) {
	pi := p.blameMgr.partyInfo

	r1 := btss.MessageRouting{
		From:                    pi.PartyIDMap["D2Ou8kohzWyVESbCOE/yXHmCAaCbB2R1jDWRpECf1JY="],
		To:                      nil,
		IsBroadcast:             false,
		IsToOldCommittee:        false,
		IsToOldAndNewCommittees: false,
	}
	msg := messages.WireMessage{
		Routing:   &r1,
		RoundInfo: "key2",
		Message:   nil,
	}
	target, err := p.blameMgr.TssWrongShareBlame(&msg)
	c.Assert(err, IsNil)
	c.Assert(target, Equals, "D2Ou8kohzWyVESbCOE/yXHmCAaCbB2R1jDWRpECf1JY=")
}

func (p *policyTestSuite) TestTssMissingShareBlame(c *C) {
	localTestPubKeys := testPubKeys[:]
	sort.Strings(localTestPubKeys)
	blameMgr := p.blameMgr
	acceptedShares := blameMgr.acceptedShares
	// we only allow a message be updated only once.
	blameMgr.acceptShareLocker.Lock()
	acceptedShares[RoundInfo{0, "testRound", "123:0"}] = []string{"1", "2"}
	acceptedShares[RoundInfo{1, "testRound", "123:0"}] = []string{"1"}
	blameMgr.acceptShareLocker.Unlock()
	nodes, _, err := blameMgr.TssMissingShareBlame(2, messages.ECDSAKEYGEN)
	sort.Slice(nodes, func(i int, j int) bool {
		return nodes[i].Pubkey < nodes[j].Pubkey
	})
	c.Assert(err, IsNil)
	c.Assert(nodes[1].Pubkey, Equals, localTestPubKeys[2])
	// we test if the missing share happens in round2
	blameMgr.acceptShareLocker.Lock()
	acceptedShares[RoundInfo{0, "testRound", "123:0"}] = []string{"1", "2", "3"}
	blameMgr.acceptShareLocker.Unlock()
	nodes, _, err = blameMgr.TssMissingShareBlame(2, messages.ECDSAKEYGEN)
	sort.Slice(nodes, func(i int, j int) bool {
		return nodes[i].Pubkey < nodes[j].Pubkey
	})
	c.Assert(err, IsNil)
	results := []string{nodes[0].Pubkey, nodes[1].Pubkey}
	sort.Strings(results)
	c.Assert(results, DeepEquals, localTestPubKeys[1:3])
}
