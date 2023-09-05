package tests

import (
	"fmt"
	"os"
	"path"
	"strconv"

	btsskeygen "github.com/HyperCore-Team/tss-lib/ecdsa/keygen"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	maddr "github.com/multiformats/go-multiaddr"

	"github.com/HyperCore-Team/go-tss/common"
	"github.com/HyperCore-Team/go-tss/conversion"
	"github.com/HyperCore-Team/go-tss/messages"
	"github.com/HyperCore-Team/go-tss/tss"
)

const (
	testFileLocation = "../test_data"
	preParamTestFile = "preParam_test.data"
)

type FourNodeTestSuite struct {
	servers       []*tss.TssServer
	ports         []int
	bootstrapPeer string
	preParams     []*btsskeygen.LocalPreParams
}

type Config struct {
	servers             []*tss.TssServer
	ports               []int
	bootstrapPeer       string
	preParams           []*btsskeygen.LocalPreParams
	partyNum            int
	newPartyNum         int
	oldPartyStartOffset int
	pubKeys             []string
	priKeyArr           []string
	keyGenSignPubkeys   []string
	blockHeight         int64
	version             string
	filename            string
}

func (c *Config) getTssServer(index int, conf common.TssConfig, bootstrap string, algo messages.Algo, whiteList map[string]bool) *tss.TssServer {
	priKey, err := conversion.GetPriKey(c.priKeyArr[index])
	if err != nil {
		panic(err)
	}
	baseHome := path.Join(os.TempDir(), c.filename, strconv.Itoa(index))
	if _, err := os.Stat(baseHome); os.IsNotExist(err) {
		err := os.MkdirAll(baseHome, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
	var peerIDs []maddr.Multiaddr
	if len(bootstrap) > 0 {
		multiAddr, err := maddr.NewMultiaddr(bootstrap)
		if err != nil {
			panic(err)
		}
		peerIDs = []maddr.Multiaddr{multiAddr}
	} else {
		peerIDs = nil
	}
	var instance *tss.TssServer
	if algo == messages.ECDSAKEYREGROUP {
		instance, err = tss.NewTss(peerIDs, c.ports[index], priKey, "Asgard", baseHome, conf, c.preParams[index], "", algo, whiteList)
		if err != nil {
			panic(err)
		}
	} else {
		instance, err = tss.NewTss(peerIDs, c.ports[index], priKey, "Asgard", baseHome, conf, nil, "", algo, whiteList)
	}
	return instance
}

func GetPubKeyFromID(id string) (crypto.PubKey, error) {
	peerID, err := peer.Decode(id)
	if err != nil {
		return nil, fmt.Errorf("fail to decode peer id: %w", err)
	}
	pk, err := peerID.ExtractPublicKey()
	return pk, err
}
