package tss

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	btsskeygen "github.com/binance-chain/tss-lib/ecdsa/keygen"
	maddr "github.com/multiformats/go-multiaddr"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/tss/go-tss/common"
	"gitlab.com/thorchain/tss/go-tss/conversion"
	"gitlab.com/thorchain/tss/go-tss/monero_multi_sig/keygen"
)

const (
	partyNum         = 6
	testFileLocation = "../test_data"
	preParamTestFile = "preParam_test.data"
)

var (
	testPubKeys = []string{
		"thorpub1addwnpepq2m5ng0e6vm66feecrwxp37cdvmezsysghskz3t5w2du4c48qwupxn96nrr",
		"thorpub1addwnpepq2ryyje5zr09lq7gqptjwnxqsy2vcdngvwd6z7yt5yjcnyj8c8cn559xe69",
		"thorpub1addwnpepqfey5l8v7azq0r4jlkd9hqqu8md0ff3vmtw2s6453zuzy8uf29fz54r7sr0",
		"thorpub1addwnpepqfjcw5l4ay5t00c32mmlky7qrppepxzdlkcwfs2fd5u73qrwna0vzag3y4j",
		"thorpub1addwnpepqtdklw8tf3anjz7nn5fly3uvq2e67w2apn560s4smmrt9e3x52nt2svmmu3",
		"thorpub1addwnpepqtspqyy6gk22u37ztra4hq3hdakc0w0k60sfy849mlml2vrpfr0wvm6uz09",
	}

	testPriKeyArr = []string{
		"Zjc1NGI1YzBiOTIwZjVlYzNlMzY0YjFiZjMwOTQ3ZGQ4YTg0ZWNjYzVjZjE0YmQwN2E2ZmQyMjdkMGQzNmYyNQ==",
		"ZThiMDAxOTk2MDc4ODk3YWE0YThlMjdkMWY0NjA1MTAwZDgyNDkyYzdhNmMwZWQ3MDBhMWIyMjNmNGMzYjVhYg==",
		"YzI2ZDU3YzMwMmMwMjRkY2NjY2E0NTg0ZjdjODI4ZjE4MzhiZTJiMDRmMDU3Y2ZlMjRiMDQ1MGNmNGFhZTRlOQ==",
		"ZTc2ZjI5OTIwOGVlMDk2N2M3Yzc1MjYyODQ0OGUyMjE3NGJiOGRmNGQyZmVmODg0NzQwNmUzYTk1YmQyODlmNA==",
		"MjQ1MDc2MmM4MjU5YjRhZjhhNmFjMmI0ZDBkNzBkOGE1ZTBmNDQ5NGI4NzM4OTYyM2E3MmI0OWMzNmE1ODZhNw==",
		"YmNiMzA2ODU1NWNjMzk3NDE1OWMwMTM3MDU0NTNjN2YwMzYzZmVhZDE5NmU3NzRhOTMwOWIxN2QyZTQ0MzdkNg==",
	}
)

func TestPackage(t *testing.T) {
	TestingT(t)
}

type FourNodeTestSuite struct {
	servers       []*TssServer
	ports         []int
	preParams     []*btsskeygen.LocalPreParams
	bootstrapPeer string
	rpcAddress    []string
}

var _ = Suite(&FourNodeTestSuite{})

// setup four nodes for test
func (s *FourNodeTestSuite) SetUpTest(c *C) {
	common.InitLog("info", true, "four_nodes_test")
	conversion.SetupBech32Prefix()
	s.ports = []int{
		16666, 16667, 16668, 16669, 16670, 16671,
	}
	s.bootstrapPeer = "/ip4/127.0.0.1/tcp/16666/p2p/16Uiu2HAm7m9i8A7cPENuL97sa5b6Xq7TSDNF6gGrSBhN41jWCmop"
	s.preParams = getPreparams(c)
	s.servers = make([]*TssServer, partyNum)
	s.rpcAddress = make([]string, partyNum)
	conf := common.TssConfig{
		KeyGenTimeout:   60 * time.Second,
		KeySignTimeout:  60 * time.Second,
		PreParamTimeout: 5 * time.Second,
		EnableMonitor:   false,
	}

	var wg sync.WaitGroup
	for i := 0; i < partyNum; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if idx == 0 {
				s.servers[idx] = s.getTssServer(c, idx, conf, "")
			} else {
				s.servers[idx] = s.getTssServer(c, idx, conf, s.bootstrapPeer)
			}
		}(i)

		time.Sleep(time.Second)
	}
	wg.Wait()
	for i := 0; i < partyNum; i++ {
		c.Assert(s.servers[i].Start(), IsNil)
	}

	// update the remote wallet address once you want to test
	remoteAddress := []string{"188.166.183.111", "178.128.155.101", "188.166.158.53", "104.236.7.106", "104.248.200.163", "139.59.237.127"}
	for i := 0; i < partyNum; i++ {
		var rpcAddress string
		rpcAddress = fmt.Sprintf("http://%s:18083/json_rpc", remoteAddress[i])
		s.rpcAddress[i] = rpcAddress
	}
}

func hash(payload []byte) []byte {
	h := sha256.New()
	h.Write(payload)
	return h.Sum(nil)
}

// we do for both join party schemes
func (s *FourNodeTestSuite) Test4NodesTss(c *C) {
	s.doTestKeygen(c, true)

	//time.Sleep(time.Second * 2)
	//s.doTestFailJoinParty(c, false)
	//time.Sleep(time.Second * 2)
	//s.doTestFailJoinParty(c, true)
	//
	//time.Sleep(time.Second * 2)
	//s.doTestBlame(c, false)
	//time.Sleep(time.Second * 2)
	//s.doTestBlame(c, true)
}

// generate a new key
func (s *FourNodeTestSuite) doTestKeygen(c *C, newJoinParty bool) {
	wg := sync.WaitGroup{}
	lock := &sync.Mutex{}
	keygenResult := make(map[int]keygen.Response)
	for i := 0; i < partyNum; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			var req keygen.Request
			if newJoinParty {
				req = keygen.NewRequest(testPubKeys, 10, "0.14.0", s.rpcAddress[idx])
			} else {
				req = keygen.NewRequest(testPubKeys, 10, "0.13.0", s.rpcAddress[idx])
			}
			res, err := s.servers[idx].Keygen(req)
			c.Assert(err, IsNil)
			lock.Lock()
			defer lock.Unlock()
			keygenResult[idx] = res
		}(i)
	}
	wg.Wait()
	var poolPubKey string
	for _, item := range keygenResult {
		if len(poolPubKey) == 0 {
			poolPubKey = item.PoolAddress
		} else {
			c.Assert(poolPubKey, Equals, item.PoolAddress)
		}
	}
}

// you need to ensure that you have enough money to run the keysign
func (s *FourNodeTestSuite) doTestKeySign(c *C, newJoinParty bool) {
	//keysignReqWithErr := keysign.NewRequest(poolPubKey, "helloworld", 10, testPubKeys, "0.13.0")
	//if newJoinParty {
	//	keysignReqWithErr = keysign.NewRequest(poolPubKey, "helloworld", 10, testPubKeys, "0.14.0")
	//}
	//
	//resp, err := s.servers[0].KeySign(keysignReqWithErr)
	//c.Assert(err, NotNil)
	//c.Assert(resp.S, Equals, "")
	//if !newJoinParty {
	//	keysignReqWithErr1 := keysign.NewRequest(poolPubKey, base64.StdEncoding.EncodeToString(hash([]byte("helloworld"))), 10, testPubKeys[:1], "0.13.0")
	//	resp, err = s.servers[0].KeySign(keysignReqWithErr1)
	//	c.Assert(err, NotNil)
	//	c.Assert(resp.S, Equals, "")
	//}
	//if !newJoinParty {
	//	keysignReqWithErr2 := keysign.NewRequest(poolPubKey, base64.StdEncoding.EncodeToString(hash([]byte("helloworld"))), 10, nil, "0.13.0")
	//	resp, err = s.servers[0].KeySign(keysignReqWithErr2)
	//	c.Assert(err, NotNil)
	//	c.Assert(resp.S, Equals, "")
	//}
	//var keysignReq keysign.Request
	//if newJoinParty {
	//	keysignReq = keysign.NewRequest(poolPubKey, base64.StdEncoding.EncodeToString(hash([]byte("helloworld"))), 10, testPubKeys, "0.14.0")
	//} else {
	//	keysignReq = keysign.NewRequest(poolPubKey, base64.StdEncoding.EncodeToString(hash([]byte("helloworld"))), 10, testPubKeys, "0.13.0")
	//}
	//keysignResult := make(map[int]keysign.Response)
	//for i := 0; i < partyNum; i++ {
	//	wg.Add(1)
	//	go func(idx int) {
	//		defer wg.Done()
	//		res, err := s.servers[idx].KeySign(keysignReq)
	//		c.Assert(err, IsNil)
	//		lock.Lock()
	//		defer lock.Unlock()
	//		keysignResult[idx] = res
	//	}(i)
	//}
	//wg.Wait()
	//var signature string
	//for _, item := range keysignResult {
	//	if len(signature) == 0 {
	//		signature = item.S + item.R
	//		continue
	//	}
	//	c.Assert(signature, Equals, item.S+item.R)
	//}
	//if newJoinParty {
	//	keysignReq = keysign.NewRequest(poolPubKey, base64.StdEncoding.EncodeToString(hash([]byte("helloworld"))), 10, nil, "0.14.0")
	//} else {
	//	keysignReq = keysign.NewRequest(poolPubKey, base64.StdEncoding.EncodeToString(hash([]byte("helloworld"))), 10, testPubKeys[:3], "0.13.0")
	//}
	//keysignResult1 := make(map[int]keysign.Response)
	//for i := 0; i < partyNum; i++ {
	//	wg.Add(1)
	//	go func(idx int) {
	//		defer wg.Done()
	//		res, err := s.servers[idx].KeySign(keysignReq)
	//		c.Assert(err, IsNil)
	//		lock.Lock()
	//		defer lock.Unlock()
	//		keysignResult1[idx] = res
	//	}(i)
	//}
	//wg.Wait()
	//signature = ""
	//for _, item := range keysignResult1 {
	//	if len(signature) == 0 {
	//		signature = item.S + item.R
	//		continue
	//	}
	//	c.Assert(signature, Equals, item.S+item.R)
	//}
}

//
//func (s *FourNodeTestSuite) doTestFailJoinParty(c *C, newJoinParty bool) {
//	// JoinParty should fail if there is a node that suppose to be in the keygen , but we didn't send request in
//	var req keygen.Request
//	if newJoinParty {
//		req = keygen.NewRequest(testPubKeys, 10, "0.14.0")
//	} else {
//		req = keygen.NewRequest(testPubKeys, 10, "0.13.0")
//	}
//	wg := sync.WaitGroup{}
//	lock := &sync.Mutex{}
//	keygenResult := make(map[int]keygen.Response)
//	// here we skip the first node
//	for i := 1; i < partyNum; i++ {
//		wg.Add(1)
//		go func(idx int) {
//			defer wg.Done()
//			res, err := s.servers[idx].Keygen(req)
//			c.Assert(err, IsNil)
//			lock.Lock()
//			defer lock.Unlock()
//			keygenResult[idx] = res
//		}(i)
//	}
//
//	wg.Wait()
//	c.Logf("result:%+v", keygenResult)
//	for _, item := range keygenResult {
//		c.Assert(item.PubKey, Equals, "")
//		c.Assert(item.Status, Equals, common.Fail)
//		var expectedFailNode string
//		if newJoinParty {
//			c.Assert(item.Blame.BlameNodes, HasLen, 2)
//			expectedFailNode := []string{"thorpub1addwnpepqtdklw8tf3anjz7nn5fly3uvq2e67w2apn560s4smmrt9e3x52nt2svmmu3", "thorpub1addwnpepq2ryyje5zr09lq7gqptjwnxqsy2vcdngvwd6z7yt5yjcnyj8c8cn559xe69"}
//			c.Assert(item.Blame.BlameNodes[0].Pubkey, Equals, expectedFailNode[0])
//			c.Assert(item.Blame.BlameNodes[1].Pubkey, Equals, expectedFailNode[1])
//		} else {
//			expectedFailNode = "thorpub1addwnpepqtdklw8tf3anjz7nn5fly3uvq2e67w2apn560s4smmrt9e3x52nt2svmmu3"
//			c.Assert(item.Blame.BlameNodes[0].Pubkey, Equals, expectedFailNode)
//		}
//	}
//}

//func (s *FourNodeTestSuite) doTestBlame(c *C, newJoinParty bool) {
//	expectedFailNode := "thorpub1addwnpepqtdklw8tf3anjz7nn5fly3uvq2e67w2apn560s4smmrt9e3x52nt2svmmu3"
//	var req keygen.Request
//	if newJoinParty {
//		req = keygen.NewRequest(testPubKeys, 10, "0.14.0")
//	} else {
//		req = keygen.NewRequest(testPubKeys, 10, "0.13.0")
//	}
//	wg := sync.WaitGroup{}
//	lock := &sync.Mutex{}
//	keygenResult := make(map[int]keygen.Response)
//	for i := 0; i < partyNum; i++ {
//		wg.Add(1)
//		go func(idx int) {
//			defer wg.Done()
//			res, err := s.servers[idx].Keygen(req)
//			c.Assert(err, NotNil)
//			lock.Lock()
//			defer lock.Unlock()
//			keygenResult[idx] = res
//		}(i)
//	}
//	// if we shutdown one server during keygen , he should be blamed
//
//	time.Sleep(time.Millisecond * 100)
//	s.servers[0].Stop()
//	defer func() {
//		conf := common.TssConfig{
//			KeyGenTimeout:   60 * time.Second,
//			KeySignTimeout:  60 * time.Second,
//			PreParamTimeout: 5 * time.Second,
//		}
//		s.servers[0] = s.getTssServer(c, 0, conf, s.bootstrapPeer)
//		c.Assert(s.servers[0].Start(), IsNil)
//		c.Log("we start the first server again")
//	}()
//	wg.Wait()
//	c.Logf("result:%+v", keygenResult)
//	for idx, item := range keygenResult {
//		if idx == 0 {
//			continue
//		}
//		c.Assert(item.PubKey, Equals, "")
//		c.Assert(item.Status, Equals, common.Fail)
//		c.Assert(item.Blame.BlameNodes, HasLen, 1)
//		c.Assert(item.Blame.BlameNodes[0].Pubkey, Equals, expectedFailNode)
//	}
//}

func (s *FourNodeTestSuite) TearDownTest(c *C) {
	// give a second before we shutdown the network
	time.Sleep(time.Second)
	for i := 0; i < partyNum; i++ {
		s.servers[i].Stop()
	}
	for i := 0; i < partyNum; i++ {
		tempFilePath := path.Join(os.TempDir(), strconv.Itoa(i))
		os.RemoveAll(tempFilePath)

	}
}

func (s *FourNodeTestSuite) getTssServer(c *C, index int, conf common.TssConfig, bootstrap string) *TssServer {
	priKey, err := conversion.GetPriKey(testPriKeyArr[index])
	c.Assert(err, IsNil)
	baseHome := path.Join(os.TempDir(), strconv.Itoa(index))
	if _, err := os.Stat(baseHome); os.IsNotExist(err) {
		err := os.Mkdir(baseHome, os.ModePerm)
		c.Assert(err, IsNil)
	}
	var peerIDs []maddr.Multiaddr
	if len(bootstrap) > 0 {
		multiAddr, err := maddr.NewMultiaddr(bootstrap)
		c.Assert(err, IsNil)
		peerIDs = []maddr.Multiaddr{multiAddr}
	} else {
		peerIDs = nil
	}
	instance, err := NewTss(peerIDs, s.ports[index], priKey, "Asgard", baseHome, conf, s.preParams[0], "")
	c.Assert(err, IsNil)
	return instance
}

func getPreparams(c *C) []*btsskeygen.LocalPreParams {
	var preParamArray []*btsskeygen.LocalPreParams
	buf, err := ioutil.ReadFile(path.Join(testFileLocation, preParamTestFile))
	c.Assert(err, IsNil)
	preParamsStr := strings.Split(string(buf), "\n")
	for _, item := range preParamsStr {
		var preParam btsskeygen.LocalPreParams
		val, err := hex.DecodeString(item)
		c.Assert(err, IsNil)
		c.Assert(json.Unmarshal(val, &preParam), IsNil)
		preParamArray = append(preParamArray, &preParam)
	}
	return preParamArray
}