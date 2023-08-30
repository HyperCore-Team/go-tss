package tests

import (
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/HyperCore-Team/go-tss/common"
	"github.com/HyperCore-Team/go-tss/keygen"
	"github.com/HyperCore-Team/go-tss/messages"
	keyRegroup "github.com/HyperCore-Team/go-tss/regroup"
	"github.com/HyperCore-Team/go-tss/tss"
)

func Test_Complex_ECDSA_Keygen_Multiple_Peers(t *testing.T) {
	var testPubKeys = []string{
		"D2Ou8kohzWyVESbCOE/yXHmCAaCbB2R1jDWRpECf1JY=", // 12D3KooWArSSkT7VYQPrbp6cLqWUTqQYb1rX77GhTJaUWMYjeVFs
		"8v5YUvEtN8vpNKejH1dmVi4BoEZX+c5EHoqQCXQM/WE=", // 3
		"Zlgbrnmk6xDkamTs004bZgUYbpiE5dV4rSg+MfSk4gU=", // 2
		"jzTMn5m27Cmt6EuCAuKnIzxNbVYY4EIywP0a9grmSok=", // 1
		"jzTMn5m27Cmt6EuCAuKnIzxNbVYY4EIywP0a9grmSok=", // 1

	}
	var testPriKeyArr = []string{
		"uqd3e5UDiYYHXsnV1ajK6Iggm/VxpXalzRJIQIR7fEYPY67ySiHNbJURJsI4T/JceYIBoJsHZHWMNZGkQJ/Ulg==", // 4
		"VtPQSjyBoE9sUM4Fta2DyMVGPVjuMBBOY9Ok9ZY38bLy/lhS8S03y+k0p6MfV2ZWLgGgRlf5zkQeipAJdAz9YQ==", // 3
		"czXp/ZG7mmGWcjXVKi8MWx6OrqTtU8HoFtWzYcl4c19mWBuueaTrEORqZOzTThtmBRhumITl1XitKD4x9KTiBQ==", // 2
		"t15Wtiil52NBuzqkHz8QkTjwHe98HXzKhev99MJ7kfKPNMyfmbbsKa3oS4IC4qcjPE1tVhjgQjLA/Rr2CuZKiQ==", // 1
		"t15Wtiil52NBuzqkHz8QkTjwHe98HXzKhev99MJ7kfKPNMyfmbbsKa3oS4IC4qcjPE1tVhjgQjLA/Rr2CuZKiQ==", // 1
	}
	var p2pKeys = []string{
		"12D3KooWArSSkT7VYQPrbp6cLqWUTqQYb1rX77GhTJaUWMYjeVFs",
		"12D3KooWSAumwg2rxzjsgv7LuWM4u3HqfLcnekg2NHc5TsNj5hgC",
		"12D3KooWGhsgEZA8Nc6xnYRLJxZns8KWtF6ztF73cDhAKXGiHjAc",
		"12D3KooWKTPAZDwc9VuBs7NrVsR9MrUo3L6HzrqP9grtiAFBTbCL",
		"12D3KooWKTPAZDwc9VuBs7NrVsR9MrUo3L6HzrqP9grtiAFBTbCL",
	}
	common.InitLog("info", true, "four_nodes_test")
	partyNum := 4
	s := &FourNodeTestSuite{
		ports: []int{
			18666, 18667, 18668, 18669, 18670, 18671, 18672, 18673,
		},
		bootstrapPeer: "/ip4/127.0.0.1/tcp/18666/p2p/12D3KooWArSSkT7VYQPrbp6cLqWUTqQYb1rX77GhTJaUWMYjeVFs",
		servers:       make([]*tss.TssServer, partyNum+1),
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
	for i := 0; i < partyNum+1; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if idx == 0 {
				s.servers[idx] = s.getEcdsaServer(idx, conf, "", testPriKeyArr, whiteList)
			} else if idx == partyNum {
				bootstrap := "/ip4/127.0.0.1/tcp/18668/p2p/12D3KooWGhsgEZA8Nc6xnYRLJxZns8KWtF6ztF73cDhAKXGiHjAc"
				s.servers[idx] = s.getEcdsaServer(idx, conf, bootstrap, testPriKeyArr, whiteList)
			} else {
				s.servers[idx] = s.getEcdsaServer(idx, conf, s.bootstrapPeer, testPriKeyArr, whiteList)
			}
		}(i)
		time.Sleep(time.Second)
	}
	wg.Wait()
	for i := 0; i < partyNum+1; i++ {
		s.servers[i].Start()
	}

	time.Sleep(2 * time.Second)
	lock := &sync.Mutex{}
	keygenResult := make(map[int]keygen.Response)
	for i := 0; i < partyNum+1; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			var req keygen.Request
			var localPubKeys []string
			if idx != partyNum {
				localPubKeys = append([]string{}, testPubKeys[:partyNum]...)
			} else {
				localPubKeys = append([]string{}, testPubKeys...)
			}
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
	for i := 0; i < partyNum+1; i++ {
		s.servers[i].Stop()
	}

	for i := 0; i < partyNum+1; i++ {
		tempFilePath := path.Join(os.TempDir(), "4nodes_test", strconv.Itoa(i))
		os.RemoveAll(tempFilePath)
	}
}

func Test_Complex_ECDSA_Keygen(t *testing.T) {
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
	lock := &sync.Mutex{}
	keygenResult := make(map[int]keygen.Response)
	for i := 0; i < partyNum; i++ {
		if i == 0 {
			continue
		}
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			var req keygen.Request
			localPubKeys := append([]string{}, testPubKeys[1:]...)
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
	for i := 0; i < partyNum; i++ {
		tempFilePath := path.Join(os.TempDir(), "4nodes_test", strconv.Itoa(i))
		os.RemoveAll(tempFilePath)
	}
}

func Test_Complex_ECDSA_Keygen_Remove_Resharing(t *testing.T) {
	common.InitLog("info", true, "4nodes_test")
	partyNum := 4
	config := &Config{
		ports: []int{
			18666, 18667, 18668, 18669, 18670,
		},
		bootstrapPeer: "/ip4/127.0.0.1/tcp/18667/p2p/12D3KooWArSSkT7VYQPrbp6cLqWUTqQYb1rX77GhTJaUWMYjeVFs",
		servers:       make([]*tss.TssServer, partyNum+1),
		pubKeys: []string{
			"8v5YUvEtN8vpNKejH1dmVi4BoEZX+c5EHoqQCXQM/WE=", // 2
			"D2Ou8kohzWyVESbCOE/yXHmCAaCbB2R1jDWRpECf1JY=", // 1
			"QN/e29H3ZWukslgoTOaSQjiIvTI8ZHDlZPUsJCj9lM4=", // 5
			"Zlgbrnmk6xDkamTs004bZgUYbpiE5dV4rSg+MfSk4gU=", // 3
			"jzTMn5m27Cmt6EuCAuKnIzxNbVYY4EIywP0a9grmSok=", // 4
		},
		priKeyArr: []string{
			"VtPQSjyBoE9sUM4Fta2DyMVGPVjuMBBOY9Ok9ZY38bLy/lhS8S03y+k0p6MfV2ZWLgGgRlf5zkQeipAJdAz9YQ==", // 2
			"uqd3e5UDiYYHXsnV1ajK6Iggm/VxpXalzRJIQIR7fEYPY67ySiHNbJURJsI4T/JceYIBoJsHZHWMNZGkQJ/Ulg==", // 1
			"ZEF7YkCR4LD7/nWUeUEstNaN96m2m8uPHs8ndNv0p9hA397b0fdla6SyWChM5pJCOIi9MjxkcOVk9SwkKP2Uzg==", // 5
			"czXp/ZG7mmGWcjXVKi8MWx6OrqTtU8HoFtWzYcl4c19mWBuueaTrEORqZOzTThtmBRhumITl1XitKD4x9KTiBQ==", // 3
			"t15Wtiil52NBuzqkHz8QkTjwHe98HXzKhev99MJ7kfKPNMyfmbbsKa3oS4IC4qcjPE1tVhjgQjLA/Rr2CuZKiQ==", // 4
		},
		keyGenSignPubkeys: []string{
			"D2Ou8kohzWyVESbCOE/yXHmCAaCbB2R1jDWRpECf1JY=", // 1
			"QN/e29H3ZWukslgoTOaSQjiIvTI8ZHDlZPUsJCj9lM4=", // 5
			"Zlgbrnmk6xDkamTs004bZgUYbpiE5dV4rSg+MfSk4gU=", // 3
			"jzTMn5m27Cmt6EuCAuKnIzxNbVYY4EIywP0a9grmSok=", // 4
		},
		blockHeight:         10,
		version:             "0.14.0",
		filename:            "4nodes_test",
		partyNum:            partyNum,
		newPartyNum:         1,
		oldPartyStartOffset: 1,
		preParams:           getPreparams(),
	}

	conf := common.TssConfig{
		KeyGenTimeout:     90 * time.Second,
		KeySignTimeout:    90 * time.Second,
		KeyRegroupTimeout: 90 * time.Second,
		PreParamTimeout:   120 * time.Second,
		EnableMonitor:     false,
	}

	var p2pKeys = []string{
		"12D3KooWArSSkT7VYQPrbp6cLqWUTqQYb1rX77GhTJaUWMYjeVFs", // 1
		"12D3KooWSAumwg2rxzjsgv7LuWM4u3HqfLcnekg2NHc5TsNj5hgC", // 2
		"12D3KooWGhsgEZA8Nc6xnYRLJxZns8KWtF6ztF73cDhAKXGiHjAc", // 3
		"12D3KooWKTPAZDwc9VuBs7NrVsR9MrUo3L6HzrqP9grtiAFBTbCL", // 4
		"12D3KooWEBcFfjLEDNMnTx3vnSwCSeokDCrtygtH8WmcWJTiGDFB", // 5
	}

	var whiteList map[string]bool
	whiteList = make(map[string]bool)
	for _, pubKey := range p2pKeys {
		whiteList[pubKey] = true
	}
	var wg sync.WaitGroup
	lock := &sync.Mutex{}
	fmt.Println(config.partyNum, " ", config.newPartyNum)
	for i := 0; i < config.partyNum+config.newPartyNum; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if idx == 1 {
				config.servers[idx] = config.getTssServer(idx, conf, "", messages.ECDSAKEYREGROUP, whiteList)
			} else {
				config.servers[idx] = config.getTssServer(idx, conf, config.bootstrapPeer, messages.ECDSAKEYREGROUP, whiteList)
			}
		}(i)

		time.Sleep(time.Second)
	}
	wg.Wait()
	for i := 0; i < config.partyNum; i++ {
		config.servers[i+config.oldPartyStartOffset].Start()
	}

	keygenResult := make(map[int]keygen.Response)
	fmt.Println("Key Gen Sign Pub Keys: ", config.keyGenSignPubkeys)
	fmt.Println("")
	for i := 0; i < config.partyNum; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			var req keygen.Request
			localPubKeys := append([]string{}, config.keyGenSignPubkeys...)
			req = keygen.NewRequest(localPubKeys, config.blockHeight, config.version, "ecdsa")
			res, err := config.servers[idx+config.oldPartyStartOffset].Keygen(req)
			if err != nil {
				panic(err)
			}
			lock.Lock()
			defer lock.Unlock()
			keygenResult[idx+config.oldPartyStartOffset] = res
		}(i)
	}
	wg.Wait()
	var pubKey string
	for _, item := range keygenResult {
		fmt.Println("PubKey: ", item.PubKey)
		if len(pubKey) == 0 {
			pubKey = item.PubKey
			break
		}
	}

	config.servers[0].Start()
	tempFilePath := path.Join(os.TempDir(), "4nodes_test", strconv.Itoa(3))
	os.RemoveAll(tempFilePath)
	time.Sleep(time.Second * 3)

	keyRegroupResult := make(map[int]keyRegroup.Response)
	for i := 0; i < config.partyNum+config.newPartyNum; i++ {
		req := keyRegroup.NewRequest(pubKey, config.pubKeys[1:], config.pubKeys[:4], config.blockHeight, config.version, "ecdsa")
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if idx == 0 {
				req.PoolPubKey = ""
			}
			res, _ := config.servers[idx].KeyRegroup(req)
			lock.Lock()
			defer lock.Unlock()
			keyRegroupResult[idx] = res
		}(i)
	}
	wg.Wait()
	newPubKey := keyRegroupResult[1].PubKey
	fmt.Println("PubKey", pubKey, "NewPubKey", newPubKey)
}
