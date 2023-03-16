package p2p

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/libp2p/go-libp2p-core/peer"
	maddr "github.com/multiformats/go-multiaddr"

	"github.com/libp2p/go-libp2p-core/crypto"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/tss/go-tss/messages"
)

type CommunicationTestSuite struct{}

var _ = Suite(&CommunicationTestSuite{})

func (CommunicationTestSuite) TestBasicCommunication(c *C) {
	whitelist := map[string]bool{}
	whitelist["12D3KooWE4qDcRrueTuRYWUdQZgcy7APZqBngVeXRt4Y6ytHizKV"] = true
	whitelist["12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu"] = true
	whitelist["12D3KooWKRyzVWW6ChFjQjK4miCty85Niy49tpPV95XdKu1BcvMA"] = true
	comm, err := NewCommunication("rendezvous", nil, 6668, "", whitelist)
	c.Assert(err, IsNil)
	c.Assert(comm, NotNil)
	comm.SetSubscribe(messages.TSSKeyGenMsg, "hello", make(chan *Message))
	c.Assert(comm.getSubscriber(messages.TSSKeySignMsg, "hello"), IsNil)
	c.Assert(comm.getSubscriber(messages.TSSKeyGenMsg, "hello"), NotNil)
	comm.CancelSubscribe(messages.TSSKeyGenMsg, "hello")
	comm.CancelSubscribe(messages.TSSKeyGenMsg, "whatever")
	comm.CancelSubscribe(messages.TSSKeySignMsg, "asdsdf")
}

func checkExist(a []maddr.Multiaddr, b string) bool {
	for _, el := range a {
		if el.String() == b {
			return true
		}
	}
	return false
}

func (CommunicationTestSuite) TestEstablishP2pCommunication(c *C) {
	bootstrapPeer := "/ip4/127.0.0.1/tcp/2220/p2p/12D3KooWSAumwg2rxzjsgv7LuWM4u3HqfLcnekg2NHc5TsNj5hgC"
	bootstrapPrivKey := "VtPQSjyBoE9sUM4Fta2DyMVGPVjuMBBOY9Ok9ZY38bLy/lhS8S03y+k0p6MfV2ZWLgGgRlf5zkQeipAJdAz9YQ=="
	fakeExternalIP := "11.22.33.44"
	fakeExternalMultiAddr := "/ip4/11.22.33.44/tcp/2220"
	validMultiAddr, err := maddr.NewMultiaddr(bootstrapPeer)
	c.Assert(err, IsNil)
	privKey, err := base64.StdEncoding.DecodeString(bootstrapPrivKey)
	c.Assert(err, IsNil)
	whitelist := map[string]bool{}
	whitelist["12D3KooWE4qDcRrueTuRYWUdQZgcy7APZqBngVeXRt4Y6ytHizKV"] = true
	whitelist["12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu"] = true
	whitelist["12D3KooWKRyzVWW6ChFjQjK4miCty85Niy49tpPV95XdKu1BcvMA"] = true
	comm, err := NewCommunication("commTest", nil, 2220, fakeExternalIP, whitelist)
	c.Assert(err, IsNil)
	c.Assert(comm.Start(privKey), IsNil)

	defer comm.Stop()
	sk1, _, err := crypto.GenerateEd25519Key(rand.Reader)
	sk1raw, _ := sk1.Raw()
	c.Assert(err, IsNil)
	comm2, err := NewCommunication("commTest", []maddr.Multiaddr{validMultiAddr}, 2221, "", whitelist)
	c.Assert(err, IsNil)
	err = comm2.Start(sk1raw)
	c.Assert(err, IsNil)
	defer comm2.Stop()

	// we connect to an invalid peer and see
	sk2, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
	c.Assert(err, IsNil)
	id, err := peer.IDFromPrivateKey(sk2)
	c.Assert(err, IsNil)
	invalidAddr := "/ip4/127.0.0.1/tcp/2220/p2p/" + id.String()
	invalidMultiAddr, err := maddr.NewMultiaddr(invalidAddr)
	c.Assert(err, IsNil)
	comm3, err := NewCommunication("commTest", []maddr.Multiaddr{invalidMultiAddr}, 2222, "", whitelist)
	c.Assert(err, IsNil)
	err = comm3.Start(sk1raw)
	c.Assert(err, ErrorMatches, "fail to connect to bootstrap peer: fail to connect to any peer")
	defer comm3.Stop()

	// we connect to one invalid and one valid address
	comm4, err := NewCommunication("commTest", []maddr.Multiaddr{invalidMultiAddr, validMultiAddr}, 2223, "", whitelist)
	c.Assert(err, IsNil)
	err = comm4.Start(sk1raw)
	c.Assert(err, IsNil)
	defer comm4.Stop()

	// we add test for external ip advertising
	c.Assert(checkExist(comm.host.Addrs(), fakeExternalMultiAddr), Equals, true)
	ps := comm2.host.Peerstore()
	c.Assert(checkExist(ps.Addrs(comm.host.ID()), fakeExternalMultiAddr), Equals, true)
	ps = comm4.host.Peerstore()
	c.Assert(checkExist(ps.Addrs(comm.host.ID()), fakeExternalMultiAddr), Equals, true)
}
