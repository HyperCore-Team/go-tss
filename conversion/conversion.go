package conversion

import (
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strings"

	"github.com/tendermint/btcd/btcec"

	"github.com/HyperCore-Team/tss-lib/crypto"
	btss "github.com/HyperCore-Team/tss-lib/tss"
	coskey "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/decred/dcrd/dcrec/edwards/v2"
	crypto2 "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"

	"github.com/HyperCore-Team/go-tss/messages"
)

// GetPeerIDFromSecp256PubKey convert the given pubkey into a peer.ID
func GetPeerIDFromSecp256PubKey(pk []byte) (peer.ID, error) {
	if len(pk) == 0 {
		return "", errors.New("empty public key raw bytes")
	}
	ppk, err := crypto2.UnmarshalSecp256k1PublicKey(pk)
	if err != nil {
		return "", fmt.Errorf("fail to convert pubkey to the crypto pubkey used in libp2p: %w", err)
	}
	return peer.IDFromPublicKey(ppk)
}

// GetPeerIDFromEDDSAPubKey convert the given public key into a peer.ID
func GetPeerIDFromEDDSAPubKey(pk []byte) (peer.ID, error) {
	if len(pk) == 0 {
		return "", errors.New("empty public key raw bytes")
	}
	ppk, err := crypto2.UnmarshalEd25519PublicKey(pk)
	if err != nil {
		return "", fmt.Errorf("fail to convert pubkey to the crypto pubkey used in libp2p: %w", err)
	}
	return peer.IDFromPublicKey(ppk)
}

func GetPeerIDFromPartyID(partyID *btss.PartyID) (peer.ID, error) {
	if partyID == nil || !partyID.ValidateBasic() {
		return "", errors.New("invalid partyID")
	}
	pkBytes := partyID.KeyInt().Bytes()
	return GetPeerIDFromEDDSAPubKey(pkBytes)
}

func PartyIDtoPubKey(party *btss.PartyID) (string, error) {
	if party == nil || !party.ValidateBasic() {
		return "", errors.New("invalid party")
	}
	partyKeyBytes := party.GetKey()
	pk := coskey.PubKey{
		Key: partyKeyBytes,
	}
	pubKey := base64.StdEncoding.EncodeToString(pk.Bytes())
	return pubKey, nil
}

func AccPubKeysFromPartyIDs(partyIDs []string, partyIDMap map[string]*btss.PartyID) ([]string, error) {
	pubKeys := make([]string, 0)
	for _, partyID := range partyIDs {
		blameParty, ok := partyIDMap[partyID]
		if !ok {
			return nil, errors.New("cannot find the blame party")
		}
		blamedPubKey, err := PartyIDtoPubKey(blameParty)
		if err != nil {
			return nil, err
		}
		pubKeys = append(pubKeys, blamedPubKey)
	}
	return pubKeys, nil
}

func SetupPartyIDMap(partiesID []*btss.PartyID) map[string]*btss.PartyID {
	partyIDMap := make(map[string]*btss.PartyID)
	for _, id := range partiesID {
		partyIDMap[id.Id] = id
	}
	return partyIDMap
}

func GetPeersID(partyIDtoP2PID map[string]peer.ID, localPeerID string) []peer.ID {
	if partyIDtoP2PID == nil {
		return nil
	}
	peerIDs := make([]peer.ID, 0, len(partyIDtoP2PID)-1)
	for _, value := range partyIDtoP2PID {
		if value.String() == localPeerID {
			continue
		}
		peerIDs = append(peerIDs, value)
	}
	return peerIDs
}

func SetupIDMaps(parties map[string]*btss.PartyID, partyIDtoP2PID map[string]peer.ID) error {
	for id, party := range parties {
		peerID, err := GetPeerIDFromPartyID(party)
		if err != nil {
			return err
		}
		partyIDtoP2PID[id] = peerID
	}
	return nil
}

func GetParties(keys []string, localPartyKey string, retLocalParty bool, moniker string) ([]*btss.PartyID, *btss.PartyID, error) {
	var localPartyID *btss.PartyID
	var unSortedPartiesID []*btss.PartyID
	sort.Strings(keys)
	for _, item := range keys {
		pkBytes, err := base64.StdEncoding.DecodeString(item)
		if err != nil {
			panic(err)
		}
		key := new(big.Int).SetBytes(pkBytes)

		// Set up the parameters
		// Note: The `id` and `moniker` fields are for convenience to allow you to easily track participants.
		// The `id` should be a unique string representing this party in the network and `moniker` can be anything (even left blank).
		// The `uniqueKey` is a unique identifying key for this peer (such as its p2p public key) as a big.Int.
		partyID := btss.NewPartyID(item, moniker, key)
		if item == localPartyKey {
			localPartyID = partyID
		}
		unSortedPartiesID = append(unSortedPartiesID, partyID)
	}

	if localPartyID == nil && retLocalParty {
		return nil, nil, errors.New("local party is not in the list")
	}

	partiesID := btss.SortPartyIDs(unSortedPartiesID)
	return partiesID, localPartyID, nil
}

func GetPreviousKeySignUicast(current string) string {
	if strings.HasSuffix(current, messages.EDDSAKEYSIGN2) {
		return messages.EDDSAKEYSIGN1
	}
	return messages.EDDSAKEYSIGN2
}

func isOnCurve(x, y *big.Int, curve elliptic.Curve) bool {
	return curve.IsOnCurve(x, y)
}

func GetTssPubKeyEDDSA(pubKeyPoint *crypto.ECPoint) (string, error) {
	// we check whether the point is on curve according to Kudelski report
	if pubKeyPoint == nil || !isOnCurve(pubKeyPoint.X(), pubKeyPoint.Y(), btss.Edwards()) {
		return "", errors.New("[EDDSA] invalid points")
	}
	tssPubKey := edwards.PublicKey{
		Curve: edwards.Edwards(),
		X:     pubKeyPoint.X(),
		Y:     pubKeyPoint.Y(),
	}
	pubKey := base64.StdEncoding.EncodeToString(tssPubKey.Serialize())
	return pubKey, nil
}

func BytesToHashString(msg []byte) (string, error) {
	h := sha256.New()
	_, err := h.Write(msg)
	if err != nil {
		return "", fmt.Errorf("fail to caculate sha256 hash: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func GetThreshold(value int) (int, error) {
	if value < 0 {
		return 0, errors.New("negative input")
	}
	threshold := int(math.Ceil(float64(value)*2.0/3.0)) - 1
	return threshold, nil
}

func GetTssPubKeyECDSA(pubKeyPoint *crypto.ECPoint) (string, error) {
	// we check whether the point is on curve according to Kudelski report
	if pubKeyPoint == nil || !isOnCurve(pubKeyPoint.X(), pubKeyPoint.Y(), btss.S256()) {
		return "", errors.New("[ECDSA] invalid points")
	}
	tssPubKey := btcec.PublicKey{
		Curve: btcec.S256(),
		X:     pubKeyPoint.X(),
		Y:     pubKeyPoint.Y(),
	}

	pubKey := base64.StdEncoding.EncodeToString(tssPubKey.SerializeCompressed())
	return pubKey, nil
}

//// EDDSA

func GetPrivateKeyFromB64String(b64Key string) (crypto2.PrivKey, error) {
	priHexBytes, err := base64.StdEncoding.DecodeString(b64Key)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}
	privateKey, err := crypto2.UnmarshalEd25519PrivateKey(priHexBytes)
	return privateKey, err
}

func GetEDDSAPrivateKeyRawBytes(privateKey crypto2.PrivKey) ([]byte, error) {
	var keyBytesArray [64]byte
	pk, err := privateKey.Raw()
	if err != nil {
		return nil, err
	}
	copy(keyBytesArray[:], pk[:])
	return keyBytesArray[:], nil
}
