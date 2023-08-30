package tests

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/HyperCore-Team/go-tss/common"
	"github.com/HyperCore-Team/go-tss/keygen"
	"github.com/HyperCore-Team/go-tss/keysign"
	"github.com/HyperCore-Team/go-tss/tss"
	"os"
	"path"
	"strconv"
	"sync"
	"testing"
	"time"
)

func Test_Different_Messages_KeySign(t *testing.T) {
	var testPubKeys = []string{
		"D2Ou8kohzWyVESbCOE/yXHmCAaCbB2R1jDWRpECf1JY=", // 12D3KooWArSSkT7VYQPrbp6cLqWUTqQYb1rX77GhTJaUWMYjeVFs
		"8v5YUvEtN8vpNKejH1dmVi4BoEZX+c5EHoqQCXQM/WE=", // 3
		"Zlgbrnmk6xDkamTs004bZgUYbpiE5dV4rSg+MfSk4gU=", // 2
		"jzTMn5m27Cmt6EuCAuKnIzxNbVYY4EIywP0a9grmSok=", // 1
		"QN/e29H3ZWukslgoTOaSQjiIvTI8ZHDlZPUsJCj9lM4=", // 5
	}
	var testPriKeyArr = []string{
		"uqd3e5UDiYYHXsnV1ajK6Iggm/VxpXalzRJIQIR7fEYPY67ySiHNbJURJsI4T/JceYIBoJsHZHWMNZGkQJ/Ulg==", // 4
		"VtPQSjyBoE9sUM4Fta2DyMVGPVjuMBBOY9Ok9ZY38bLy/lhS8S03y+k0p6MfV2ZWLgGgRlf5zkQeipAJdAz9YQ==", // 3
		"czXp/ZG7mmGWcjXVKi8MWx6OrqTtU8HoFtWzYcl4c19mWBuueaTrEORqZOzTThtmBRhumITl1XitKD4x9KTiBQ==", // 2
		"t15Wtiil52NBuzqkHz8QkTjwHe98HXzKhev99MJ7kfKPNMyfmbbsKa3oS4IC4qcjPE1tVhjgQjLA/Rr2CuZKiQ==", // 1
		"ZEF7YkCR4LD7/nWUeUEstNaN96m2m8uPHs8ndNv0p9hA397b0fdla6SyWChM5pJCOIi9MjxkcOVk9SwkKP2Uzg==", // 5
	}
	var p2pKeys = []string{
		"12D3KooWArSSkT7VYQPrbp6cLqWUTqQYb1rX77GhTJaUWMYjeVFs",
		"12D3KooWSAumwg2rxzjsgv7LuWM4u3HqfLcnekg2NHc5TsNj5hgC",
		"12D3KooWGhsgEZA8Nc6xnYRLJxZns8KWtF6ztF73cDhAKXGiHjAc",
		"12D3KooWKTPAZDwc9VuBs7NrVsR9MrUo3L6HzrqP9grtiAFBTbCL",
		"12D3KooWEBcFfjLEDNMnTx3vnSwCSeokDCrtygtH8WmcWJTiGDFB",
	}
	common.InitLog("info", true, "four_nodes_test")
	partyNum := 5
	s := &FourNodeTestSuite{
		ports: []int{
			18666, 18667, 18668, 18669, 18670,
		},
		bootstrapPeer: "/ip4/127.0.0.1/tcp/18666/p2p/12D3KooWArSSkT7VYQPrbp6cLqWUTqQYb1rX77GhTJaUWMYjeVFs",
		servers:       make([]*tss.TssServer, partyNum),
		preParams:     getPreparams(),
	}

	conf := common.TssConfig{
		KeyGenTimeout:   90 * time.Second,
		KeySignTimeout:  90 * time.Second,
		PreParamTimeout: 120 * time.Second,
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

	time.Sleep(2 * time.Second)
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

	fmt.Println("Public Key: ", poolPubKey)
	time.Sleep(5 * time.Second)
	keysignResult := make(map[int]keysign.Response)
	for i := 0; i < partyNum; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			localPubKeys := append([]string{}, testPubKeys...)
			var keysignReq keysign.Request
			if idx == 3 {
				keysignReq = keysign.NewRequest(poolPubKey, []string{base64.StdEncoding.EncodeToString([]byte("helloworld")), base64.StdEncoding.EncodeToString([]byte(fmt.Sprint(3, "helloworld"))), base64.StdEncoding.EncodeToString([]byte(fmt.Sprint(2, "helloworld")))}, 10, localPubKeys, "0.13.0", "ecdsa")
			} else {
				keysignReq = keysign.NewRequest(poolPubKey, []string{base64.StdEncoding.EncodeToString([]byte("helloworld")), base64.StdEncoding.EncodeToString([]byte(fmt.Sprint(3, "helloworld"))), base64.StdEncoding.EncodeToString([]byte(fmt.Sprint(2, "helloworld")))}, 10, localPubKeys, "0.13.0", "ecdsa")
			}
			res, err := s.servers[idx].KeySign(keysignReq)
			if err != nil {
				panic(err)
			}
			lock.Lock()
			defer lock.Unlock()
			keysignResult[idx] = res
		}(i)
	}
	wg.Wait()
	fmt.Println("Keysignresult: ", keysignResult)
	//fmt.Println("Keysignresult: ", keysignResult[0].Signatures[0].Msg, keysignResult[0].Signatures[1].Msg)
	for i := 0; i < partyNum; i++ {
		sDec1, _ := base64.StdEncoding.DecodeString(keysignResult[i].Signatures[0].Msg)
		sDec2, _ := base64.StdEncoding.DecodeString(keysignResult[i].Signatures[1].Msg)
		sDec3, _ := base64.StdEncoding.DecodeString(keysignResult[i].Signatures[2].Msg)
		fmt.Println(i, "Keysignresult: ", string(sDec1), string(sDec2), string(sDec3))
		s.servers[i].Stop()
	}
	for i := 0; i < partyNum; i++ {
		tempFilePath := path.Join(os.TempDir(), "4nodes_test", strconv.Itoa(i))
		os.RemoveAll(tempFilePath)
	}
}
