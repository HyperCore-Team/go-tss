package conversion

import (
	"encoding/base64"
	"errors"
	"fmt"

	coskey "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/decred/dcrd/dcrec/edwards/v2"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/tendermint/btcd/btcec"
	tcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/HyperCore-Team/go-tss/messages"
)

// GetPeerIDFromPubKey get the peer.ID from bech32 format node pub key
func GetPeerIDFromPubKey(pubkey string) (peer.ID, error) {
	pk, err := base64.StdEncoding.DecodeString(pubkey)
	if err != nil {
		return "", fmt.Errorf("fail to parse account pub key(%s): %w", pubkey, err)
	}
	ppk, err := crypto.UnmarshalEd25519PublicKey(pk)
	if err != nil {
		return "", fmt.Errorf("fail to convert pubkey to the crypto pubkey used in libp2p: %w", err)
	}
	return peer.IDFromPublicKey(ppk)
}

// GetPeerIDsFromPubKeys convert a list of node pub key to their peer.ID
func GetPeerIDsFromPubKeys(pubkeys []string) ([]peer.ID, error) {
	var peerIDs []peer.ID
	for _, item := range pubkeys {
		peerID, err := GetPeerIDFromPubKey(item)
		if err != nil {
			return nil, err
		}
		peerIDs = append(peerIDs, peerID)
	}
	return peerIDs, nil
}

// GetPeerIDs return a slice of peer id
func GetPeerIDs(pubkeys []string) ([]peer.ID, error) {
	var peerIDs []peer.ID
	for _, item := range pubkeys {
		pID, err := GetPeerIDFromPubKey(item)
		if err != nil {
			return nil, fmt.Errorf("fail to get peer id from pubkey(%s):%w", item, err)
		}
		peerIDs = append(peerIDs, pID)
	}
	return peerIDs, nil
}

// GetPubKeysFromPeerIDs given a list of peer ids, and get a list og pub keys.
func GetPubKeysFromPeerIDs(peers []string) ([]string, error) {
	var result []string
	for _, item := range peers {
		pKey, err := GetPubKeyFromPeerID(item)
		if err != nil {
			return nil, fmt.Errorf("fail to get pubkey from peerID: %w", err)
		}
		result = append(result, pKey)
	}
	return result, nil
}

// GetPubKeyFromPeerID extract the pub key from PeerID
func GetPubKeyFromPeerID(pID string) (string, error) {
	peerID, err := peer.Decode(pID)
	if err != nil {
		return "", fmt.Errorf("fail to decode peer id: %w", err)
	}
	pk, err := peerID.ExtractPublicKey()
	if err != nil {
		return "", fmt.Errorf("fail to extract pub key from peer id: %w", err)
	}
	rawBytes, err := pk.Raw()
	if err != nil {
		return "", fmt.Errorf("faail to get pub key raw bytes: %w", err)
	}
	pubKey := coskey.PubKey{
		Key: rawBytes,
	}
	key := base64.StdEncoding.EncodeToString(pubKey.Bytes())
	return key, nil
}

func GetPriKey(priKeyString string) (tcrypto.PrivKey, error) {
	rawBytes, err := base64.StdEncoding.DecodeString(priKeyString)
	if err != nil {
		return nil, fmt.Errorf("fail to decode private key: %w", err)
	}
	var priKey ed25519.PrivKey
	if len(rawBytes) < 64 {
		return nil, fmt.Errorf("fail to decode private key: %w", err)
	}
	priKey = rawBytes[:64]
	return priKey, nil
}

func GetPriKeyRawBytes(priKey tcrypto.PrivKey) ([]byte, error) {
	var keyBytesArray [64]byte
	pk, ok := priKey.(ed25519.PrivKey)
	if !ok {
		return nil, errors.New("private key is not ed25519.PrivKey")
	}
	copy(keyBytesArray[:], pk[:])
	return keyBytesArray[:], nil
}

func CheckKeyOnCurve(pk string, algo messages.Algo) (bool, error) {
	pubKey, err := base64.StdEncoding.DecodeString(pk)
	if err != nil {
		return false, err
	}
	if algo == messages.EDDSAKEYSIGN || algo == messages.EDDSAKEYREGROUP || algo == messages.EDDSAKEYGEN {
		bPk, err := edwards.ParsePubKey(pubKey)
		if err == nil {
			return isOnCurve(bPk.X, bPk.Y, edwards.Edwards()), nil
		} else {
			return false, err
		}
	} else if algo == messages.ECDSAKEYSIGN || algo == messages.ECDSAKEYREGROUP || algo == messages.ECDSAKEYGEN {
		btPk, err := btcec.ParsePubKey(pubKey, btcec.S256())
		if err != nil {
			return false, err
		}
		return isOnCurve(btPk.X, btPk.Y, btcec.S256()), nil
	}
	return false, fmt.Errorf("invalid algo")
}
