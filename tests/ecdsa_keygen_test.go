package tests

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	btsskeygen "github.com/HyperCore-Team/tss-lib/ecdsa/keygen"
	"github.com/ipfs/go-log"
	maddr "github.com/multiformats/go-multiaddr"

	"github.com/HyperCore-Team/go-tss/common"
	"github.com/HyperCore-Team/go-tss/conversion"
	"github.com/HyperCore-Team/go-tss/keygen"
	"github.com/HyperCore-Team/go-tss/messages"
	"github.com/HyperCore-Team/go-tss/tss"
)

func (s *FourNodeTestSuite) getEcdsaServer(index int, conf common.TssConfig, bootstrap string, testPriKeyArr []string, pubKeyWhitelist map[string]bool) *tss.TssServer {
	priKey, err := conversion.GetPriKey(testPriKeyArr[index])
	if err != nil {
		panic(err)
	}
	baseHome := path.Join(os.TempDir(), "4nodes_test", strconv.Itoa(index))
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
	instance, err := tss.NewTss(peerIDs, s.ports[index], priKey, "Asgard", baseHome, conf, nil, "", messages.ECDSAKEYGEN, pubKeyWhitelist)
	if err != nil {
		panic(err)
	}
	return instance
}

func getPreparams() []*btsskeygen.LocalPreParams {
	var preParamArray []*btsskeygen.LocalPreParams
	buf, err := os.ReadFile(path.Join(testFileLocation, preParamTestFile))
	if err != nil {
		return preParamArray
	}
	preParamsStr := strings.Split(string(buf), ",")
	for _, item := range preParamsStr {
		var preParam btsskeygen.LocalPreParams
		val, err := hex.DecodeString(item)
		if err != nil {
			return preParamArray
		}
		err = json.Unmarshal(val, &preParam)
		if err != nil {
			return preParamArray
		}
		preParamArray = append(preParamArray, &preParam)
	}
	return preParamArray
}

func Test_ECDSA_Keygen(t *testing.T) {
	var testPubKeys = []string{
		"D2Ou8kohzWyVESbCOE/yXHmCAaCbB2R1jDWRpECf1JY=", // 12D3KooWArSSkT7VYQPrbp6cLqWUTqQYb1rX77GhTJaUWMYjeVFs
		"8v5YUvEtN8vpNKejH1dmVi4BoEZX+c5EHoqQCXQM/WE=", // 3
		"Zlgbrnmk6xDkamTs004bZgUYbpiE5dV4rSg+MfSk4gU=", // 2
		"jzTMn5m27Cmt6EuCAuKnIzxNbVYY4EIywP0a9grmSok=", // 1
	}
	var testPriKeyArr = []string{
		"uqd3e5UDiYYHXsnV1ajK6Iggm/VxpXalzRJIQIR7fEYPY67ySiHNbJURJsI4T/JceYIBoJsHZHWMNZGkQJ/Ulg==", // 4
		"VtPQSjyBoE9sUM4Fta2DyMVGPVjuMBBOY9Ok9ZY38bLy/lhS8S03y+k0p6MfV2ZWLgGgRlf5zkQeipAJdAz9YQ==", // 3
		"czXp/ZG7mmGWcjXVKi8MWx6OrqTtU8HoFtWzYcl4c19mWBuueaTrEORqZOzTThtmBRhumITl1XitKD4x9KTiBQ==", // 2
		"t15Wtiil52NBuzqkHz8QkTjwHe98HXzKhev99MJ7kfKPNMyfmbbsKa3oS4IC4qcjPE1tVhjgQjLA/Rr2CuZKiQ==", // 1
	}
	var p2pKeys = []string{
		"12D3KooWArSSkT7VYQPrbp6cLqWUTqQYb1rX77GhTJaUWMYjeVFs",
		"12D3KooWSAumwg2rxzjsgv7LuWM4u3HqfLcnekg2NHc5TsNj5hgC",
		"12D3KooWGhsgEZA8Nc6xnYRLJxZns8KWtF6ztF73cDhAKXGiHjAc",
		"12D3KooWKTPAZDwc9VuBs7NrVsR9MrUo3L6HzrqP9grtiAFBTbCL",
	}

	partyNum := 4
	common.InitLog("info", true, "four_nodes_test")
	log.SetLogLevel("tss-lib", "debug")
	s := &FourNodeTestSuite{
		ports: []int{
			18666, 18667, 18668, 18669,
		},
		bootstrapPeer: "/ip4/127.0.0.1/tcp/18666/p2p/12D3KooWArSSkT7VYQPrbp6cLqWUTqQYb1rX77GhTJaUWMYjeVFs",
		servers:       make([]*tss.TssServer, partyNum),
		preParams:     getPreparams(),
	}

	conf := common.TssConfig{
		KeyGenTimeout:   90 * time.Second,
		KeySignTimeout:  90 * time.Second,
		PreParamTimeout: 200 * time.Second,
		EnableMonitor:   false,
	}
	var whiteList map[string]bool
	whiteList = make(map[string]bool)
	for _, pubKey := range p2pKeys {
		whiteList[pubKey] = true
	}

	var wg sync.WaitGroup
	for i := 0; i < partyNum; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if idx == 0 {
				s.servers[idx] = s.getEcdsaServer(idx, conf, "", testPriKeyArr, whiteList)
			} else {
				s.servers[idx] = s.getEcdsaServer(idx, conf, s.bootstrapPeer, testPriKeyArr, whiteList)
			}
		}(i)
		time.Sleep(time.Second)
	}
	wg.Wait()
	for i := 0; i < partyNum; i++ {
		s.servers[i].Start()
	}
	lock := &sync.Mutex{}
	keygenResult := make(map[int]keygen.Response)
	for i := 0; i < partyNum; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			var req keygen.Request
			localPubKeys := append([]string{}, testPubKeys...)
			req = keygen.NewRequest(localPubKeys, 10, "0.13.0", "ecdsa")
			res, err := s.servers[idx].Keygen(req)
			if err != nil {
				panic(err)
			}
			lock.Lock()
			defer lock.Unlock()
			keygenResult[idx] = res
		}(i)
	}
	wg.Wait()
	var poolPubKey string
	for _, item := range keygenResult {
		if len(poolPubKey) == 0 {
			poolPubKey = item.PubKey
		}
	}
	fmt.Println("Public Key: ", hex.EncodeToString([]byte(poolPubKey)))
	time.Sleep(5 * time.Second)
	for i := 0; i < partyNum; i++ {
		s.servers[i].Stop()
	}
	//for i := 0; i < partyNum; i++ {
	//	tempFilePath := path.Join(os.TempDir(), "4nodes_test", strconv.Itoa(i))
	//	os.RemoveAll(tempFilePath)
	//}
}
