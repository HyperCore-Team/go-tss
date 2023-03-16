package tests

import (
	"fmt"
	"github.com/ipfs/go-log"
	"gitlab.com/thorchain/tss/go-tss/common"
	"gitlab.com/thorchain/tss/go-tss/keygen"
	"gitlab.com/thorchain/tss/go-tss/messages"
	keyRegroup "gitlab.com/thorchain/tss/go-tss/regroup"
	"gitlab.com/thorchain/tss/go-tss/tss"
	"sync"
	"testing"
	"time"
)

func Test_ECDSA_Resharing(t *testing.T) {
	common.InitLog("info", true, "4nodes_test")
	log.SetLogLevel("tss-lib", "debug")
	partyNum := 4
	config := &Config{
		ports: []int{
			18666, 18667, 18668, 18669, 18670,
		},
		bootstrapPeer: "/ip4/127.0.0.1/tcp/18666/p2p/12D3KooWSAumwg2rxzjsgv7LuWM4u3HqfLcnekg2NHc5TsNj5hgC",
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
		KeyRegroupTimeout: 120 * time.Second,
		PreParamTimeout:   5 * time.Second,
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
	for i := 0; i < config.partyNum+config.newPartyNum; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if idx == 0 {
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
			keygenResult[idx] = res
		}(i)
	}
	wg.Wait()
	var pubKey string
	print("KeygenResult: ", keygenResult)
	for _, item := range keygenResult {
		fmt.Println("PubKey: ", item.PubKey)
		if len(pubKey) == 0 {
			pubKey = item.PubKey
			break
		}
	}

	config.servers[0].Start()
	time.Sleep(time.Second * 2)

	keyRegroupResult := make(map[int]keyRegroup.Response)
	for i := 0; i < config.partyNum+config.newPartyNum; i++ {
		req := keyRegroup.NewRequest(pubKey, config.pubKeys[config.oldPartyStartOffset:], config.pubKeys[:config.partyNum], config.blockHeight, config.version, "ecdsa")
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
			fmt.Println("Result: ", res.PubKey, "Idx", idx)
		}(i)
	}
	wg.Wait()
	newPubKey := keyRegroupResult[1].PubKey
	fmt.Println("PubKey", pubKey, "NewPubKey", newPubKey)
}
