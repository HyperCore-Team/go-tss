package conversion

import (
	"encoding/base64"
	"errors"
	"math/rand"

	"github.com/blang/semver"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

// GetRandomPubKey for test
func GetRandomPubKey() string {
	publicKey := ed25519.GenPrivKey().PubKey()
	return base64.StdEncoding.EncodeToString(publicKey.Bytes())
}

// GetRandomPeerID for test
func GetRandomPeerID() peer.ID {
	publicKey := ed25519.GenPrivKey().PubKey()
	var pk ed25519.PubKey
	copy(pk[:], publicKey.Bytes())
	peerID, _ := GetPeerIDFromEDDSAPubKey(pk)
	return peerID
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
)

func RandStringBytesMask(n int) string {
	b := make([]byte, n)
	for i := 0; i < n; {
		if idx := int(rand.Int63() & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i++
		}
	}
	return string(b)
}

func VersionLTCheck(currentVer, expectedVer string) (bool, error) {
	c, err := semver.Make(expectedVer)
	if err != nil {
		return false, errors.New("fail to parse the expected version")
	}
	v, err := semver.Make(currentVer)
	if err != nil {
		return false, errors.New("fail to parse the current version")
	}
	return v.LT(c), nil
}
