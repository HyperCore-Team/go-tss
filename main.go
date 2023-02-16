package main

import (
	"encoding/base64"
	"fmt"
	"gitlab.com/thorchain/tss/go-tss/keysign"
	"os"
	"path"
	"strconv"
	"time"

	btsskeygen "github.com/binance-chain/tss-lib/ecdsa/keygen"
	maddr "github.com/multiformats/go-multiaddr"

	"gitlab.com/thorchain/tss/go-tss/common"
	"gitlab.com/thorchain/tss/go-tss/conversion"
	"gitlab.com/thorchain/tss/go-tss/keygen"
	"gitlab.com/thorchain/tss/go-tss/messages"
	"gitlab.com/thorchain/tss/go-tss/tss"
)

func getEcdsaServer(index int, port int, privKey string, param *btsskeygen.LocalPreParams, conf common.TssConfig, bootstrap string, pubKeyWhitelist map[string]bool) *tss.TssServer {
	priKey, err := conversion.GetPriKey(privKey)
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
	var preparam *btsskeygen.LocalPreParams
	if param != nil {
		preparam = param
	}
	instance, err := tss.NewTss(peerIDs, port, priKey, "Asgard", baseHome, conf, preparam, "", messages.ECDSAKEYGEN, pubKeyWhitelist)
	if err != nil {
		panic(err)
	}
	return instance
}

func getEddsaServer(index int, port int, privKey string, conf common.TssConfig, bootstrap string, whiteList map[string]bool) *tss.TssServer {
	priKey, err := conversion.GetPriKey(privKey)
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
	instance, err := tss.NewTss(peerIDs, port, priKey, "Asgard", baseHome, conf, nil, "", messages.EDDSAKEYGEN, whiteList)
	return instance
}

func main() {
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
	conf := common.TssConfig{
		KeyGenTimeout:   90 * time.Second,
		KeySignTimeout:  90 * time.Second,
		PreParamTimeout: 120 * time.Second,
		EnableMonitor:   false,
	}
	argsWithoutProg := os.Args[1:]
	index, err := strconv.Atoi(argsWithoutProg[0])
	if err != nil {
		panic(err)
	}
	ports := []int{
		18666, 18667, 18668, 18669,
	}
	port := ports[index]
	privKey := testPriKeyArr[index]
	bootStrap := "/ip4/192.168.88.247/tcp/18666/p2p/12D3KooWArSSkT7VYQPrbp6cLqWUTqQYb1rX77GhTJaUWMYjeVFs"
	if port == 18666 {
		bootStrap = ""
	}
	var p2pKeys = []string{
		"12D3KooWArSSkT7VYQPrbp6cLqWUTqQYb1rX77GhTJaUWMYjeVFs",
		"12D3KooWSAumwg2rxzjsgv7LuWM4u3HqfLcnekg2NHc5TsNj5hgC",
		"12D3KooWGhsgEZA8Nc6xnYRLJxZns8KWtF6ztF73cDhAKXGiHjAc",
		"12D3KooWKTPAZDwc9VuBs7NrVsR9MrUo3L6HzrqP9grtiAFBTbCL",
	}
	var whiteList map[string]bool
	whiteList = make(map[string]bool)
	for _, pubKey := range p2pKeys {
		whiteList[pubKey] = true
	}
	fmt.Println("Port: ", port, "PrivateKey: ", privKey, "Bootstrap: ", bootStrap)
	start := time.Now()
	server := getEddsaServer(index, port, privKey, conf, bootStrap, whiteList)
	server.Start()
	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Println("Start server: ", elapsed)
	time.Sleep(time.Second * 2)
	localPubKeys := append([]string{}, testPubKeys[:3]...)
	start = time.Now()
	var req keygen.Request
	req = keygen.NewRequest(localPubKeys, 10, "0.13.0", "eddsa")
	res, err := server.Keygen(req)
	if err != nil {
		panic(err)
	}
	fmt.Println("\n\nPubKey:", res.PubKey, "\n\n")
	t = time.Now()
	elapsed = t.Sub(start)
	fmt.Println("Generate PubKey: ", elapsed)

	start = time.Now()
	var keysignReq keysign.Request
	keysignReq = keysign.NewRequest(res.PubKey, []string{base64.StdEncoding.EncodeToString([]byte("helloworld")), base64.StdEncoding.EncodeToString([]byte("helloworld2"))}, 10, localPubKeys, "0.13.0", "eddsa")
	resSign, err := server.KeySign(keysignReq)
	if err != nil {
		panic(err)
	}
	fmt.Println("\n\nSignature: ", resSign.Signatures[0].Msg, "\n\n")
	end := time.Now()
	elapsed = end.Sub(start)
	fmt.Println("Sign single message: ", elapsed)

	server.Stop()
}
