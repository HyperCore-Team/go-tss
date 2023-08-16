package common

import (
	"bytes"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"

	"github.com/binance-chain/tss-lib/ecdsa/resharing"
	eddsaResharing "github.com/binance-chain/tss-lib/eddsa/resharing"

	ecdsaKeygen "github.com/binance-chain/tss-lib/ecdsa/keygen"
	ecdsaKeySign "github.com/binance-chain/tss-lib/ecdsa/signing"
	eddsaKeygen "github.com/binance-chain/tss-lib/eddsa/keygen"
	eddsaSigning "github.com/binance-chain/tss-lib/eddsa/signing"
	"github.com/binance-chain/tss-lib/tss"
	btss "github.com/binance-chain/tss-lib/tss"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	tcrypto "github.com/tendermint/tendermint/crypto"

	"github.com/HyperCore-Team/go-tss/blame"
	"github.com/HyperCore-Team/go-tss/messages"
)

func Contains(s []*btss.PartyID, e *btss.PartyID) bool {
	if e == nil {
		return false
	}
	for _, a := range s {
		if *a == *e {
			return true
		}
	}
	return false
}

func MsgToHashInt(msg []byte, algo messages.Algo) (*big.Int, error) {
	if algo == messages.ECDSAKEYSIGN {
		return hashToInt(msg, tss.S256()), nil
	} else if algo == messages.EDDSAKEYSIGN {
		return hashToInt(msg, tss.Edwards()), nil
	} else {
		return nil, errors.New("invalid algo")
	}
}

func MsgToHashString(msg []byte) (string, error) {
	if len(msg) == 0 {
		return "", errors.New("empty message")
	}
	h := sha256.New()
	_, err := h.Write(msg)
	if err != nil {
		return "", fmt.Errorf("fail to caculate sha256 hash: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func hashToInt(hash []byte, c elliptic.Curve) *big.Int {
	orderBits := c.Params().N.BitLen()
	orderBytes := (orderBits + 7) / 8
	if len(hash) > orderBytes {
		hash = hash[:orderBytes]
	}

	ret := new(big.Int).SetBytes(hash)
	excess := len(hash)*8 - orderBits
	if excess > 0 {
		ret.Rsh(ret, uint(excess))
	}
	return ret
}

func InitLog(level string, pretty bool, serviceValue string) {
	l, err := zerolog.ParseLevel(level)
	if err != nil {
		log.Warn().Msgf("%s is not a valid log-level, falling back to 'info'", level)
	}
	var out io.Writer = os.Stdout
	if pretty {
		out = zerolog.ConsoleWriter{Out: os.Stdout}
	}
	zerolog.SetGlobalLevel(l)
	log.Logger = log.Output(out).With().Str("service", serviceValue).Logger()
}

func generateSignature(msg []byte, msgID string, privKey tcrypto.PrivKey) ([]byte, error) {
	var dataForSigning bytes.Buffer
	dataForSigning.Write(msg)
	dataForSigning.WriteString(msgID)
	return privKey.Sign(dataForSigning.Bytes())
}

func verifySignature(pubKey tcrypto.PubKey, message, sig []byte, msgID string) bool {
	var dataForSign bytes.Buffer
	dataForSign.Write(message)
	dataForSign.WriteString(msgID)
	return pubKey.VerifySignature(dataForSign.Bytes(), sig)
}

func getHighestFreq(confirmedList map[string]string) (string, int, error) {
	if len(confirmedList) == 0 {
		return "", 0, errors.New("empty input")
	}
	freq := make(map[string]int, len(confirmedList))
	for _, n := range confirmedList {
		freq[n]++
	}
	maxFreq := -1
	var data string
	for key, counter := range freq {
		if counter > maxFreq {
			maxFreq = counter
			data = key
		}
	}
	return data, maxFreq, nil
}

// due to the nature of tss, we may find the invalid share of the previous round only
// when we get the shares from the peers in the current round. So, when we identify
// an error in this round, we check whether the previous round is the unicast
func checkUnicast(round blame.RoundInfo) bool {
	index := round.Index
	isEddsa := strings.Contains(round.RoundMsg, "EDDSA")
	if isEddsa {
		isKeyGen := strings.Contains(round.RoundMsg, "KGR")
		// keygen unicast blame
		if isKeyGen {
			if index == 1 {
				return true
			}
		}
		isRegroup := strings.Contains(round.RoundMsg, "DGR")
		if isRegroup {
			if index == 3 {
				return true
			} else {
				return false
			}
		}
		isKeySign := strings.Contains(round.RoundMsg, "SignR")
		if isKeySign {
			// keysign unicast blame
			if index > 2 {
				return true
			}
		}
	} else {
		isKeyGen := strings.Contains(round.RoundMsg, "KGR")
		// keygen unicast blame
		if isKeyGen {
			if index == 1 || index == 2 {
				return true
			}
			return false
		}
		isRegroup := strings.Contains(round.RoundMsg, "DGR")
		if isRegroup {
			if index == 3 {
				return true
			} else {
				return false
			}
		}
		// keysign unicast blame
		if index < 5 {
			return true
		}
	}
	return false
}

func GetMsgRound(msg []byte, partyID *btss.PartyID, isBroadcast bool) (blame.RoundInfo, error) {
	parsedMsg, err := btss.ParseWireMessage(msg, partyID, isBroadcast)
	if err != nil {
		return blame.RoundInfo{}, err
	}
	switch parsedMsg.Content().(type) {
	// ECDSA -- KeyGen
	case *ecdsaKeygen.KGRound1Message:
		return blame.RoundInfo{
			Index:    0,
			RoundMsg: messages.KEYGEN1,
		}, nil
	case *ecdsaKeygen.KGRound2Message1:
		return blame.RoundInfo{
			Index:    1,
			RoundMsg: messages.KEYGEN2aUnicast,
		}, nil
	case *ecdsaKeygen.KGRound2Message2:
		return blame.RoundInfo{
			Index:    2,
			RoundMsg: messages.KEYGEN2b,
		}, nil
	case *ecdsaKeygen.KGRound3Message:
		return blame.RoundInfo{
			Index:    3,
			RoundMsg: messages.KEYGEN3,
		}, nil

	// ECDSA -- KeySign
	case *ecdsaKeySign.SignRound1Message1:
		return blame.RoundInfo{
			Index:    0,
			RoundMsg: messages.KEYSIGN1aUnicast,
		}, nil
	case *ecdsaKeySign.SignRound1Message2:
		return blame.RoundInfo{
			Index:    1,
			RoundMsg: messages.KEYSIGN1b,
		}, nil
	case *ecdsaKeySign.SignRound2Message:
		return blame.RoundInfo{
			Index:    2,
			RoundMsg: messages.KEYSIGN2Unicast,
		}, nil
	case *ecdsaKeySign.SignRound3Message:
		return blame.RoundInfo{
			Index:    3,
			RoundMsg: messages.KEYSIGN3,
		}, nil
	case *ecdsaKeySign.SignRound4Message:
		return blame.RoundInfo{
			Index:    4,
			RoundMsg: messages.KEYSIGN4,
		}, nil
	case *ecdsaKeySign.SignRound5Message:
		return blame.RoundInfo{
			Index:    5,
			RoundMsg: messages.KEYSIGN5,
		}, nil
	case *ecdsaKeySign.SignRound6Message:
		return blame.RoundInfo{
			Index:    6,
			RoundMsg: messages.KEYSIGN6,
		}, nil
	case *ecdsaKeySign.SignRound7Message:
		return blame.RoundInfo{
			Index:    7,
			RoundMsg: messages.KEYSIGN7,
		}, nil
	case *ecdsaKeySign.SignRound8Message:
		return blame.RoundInfo{
			Index:    8,
			RoundMsg: messages.KEYSIGN8,
		}, nil
	case *ecdsaKeySign.SignRound9Message:
		return blame.RoundInfo{
			Index:    9,
			RoundMsg: messages.KEYSIGN9,
		}, nil

	// EDDSA -- Keygen
	case *eddsaKeygen.KGRound1Message:
		return blame.RoundInfo{
			Index:    0,
			RoundMsg: messages.EDDSAKEYGEN1,
		}, nil
	case *eddsaKeygen.KGRound2Message1:
		return blame.RoundInfo{
			Index:    1,
			RoundMsg: messages.EDDSAKEYGEN2a,
		}, nil
	case *eddsaKeygen.KGRound2Message2:
		return blame.RoundInfo{
			Index:    2,
			RoundMsg: messages.EDDSAKEYGEN2b,
		}, nil

	// ECDSA -- Resharing
	case *resharing.DGRound1Message:
		return blame.RoundInfo{
			Index:    0,
			RoundMsg: messages.KEYREGROUP1,
		}, nil
	case *resharing.DGRound2Message1:
		return blame.RoundInfo{
			Index:    1,
			RoundMsg: messages.KEYREGROUP2a,
		}, nil
	case *resharing.DGRound2Message2:
		return blame.RoundInfo{
			Index:    2,
			RoundMsg: messages.KEYREGROUP2b,
		}, nil
	case *resharing.DGRound3Message1:
		return blame.RoundInfo{
			Index:    3,
			RoundMsg: messages.KEYREGROUP3a,
		}, nil
	case *resharing.DGRound3Message2:
		return blame.RoundInfo{
			Index:    4,
			RoundMsg: messages.KEYREGROUP3b,
		}, nil
	case *resharing.DGRound4Message:
		return blame.RoundInfo{
			Index:    5,
			RoundMsg: messages.KEYREGROUP4,
		}, nil

	// EDDSA -- Signing
	case *eddsaSigning.SignRound1Message:
		return blame.RoundInfo{
			Index:    0,
			RoundMsg: messages.EDDSAKEYSIGN1,
		}, nil
	case *eddsaSigning.SignRound2Message:
		return blame.RoundInfo{
			Index:    1,
			RoundMsg: messages.EDDSAKEYSIGN2,
		}, nil
	case *eddsaSigning.SignRound3Message:
		return blame.RoundInfo{
			Index:    2,
			RoundMsg: messages.EDDSAKEYSIGN3,
		}, nil

	// EDDSA -- Resharing
	case *eddsaResharing.DGRound1Message:
		return blame.RoundInfo{
			Index:    1,
			RoundMsg: messages.EDDSAKEYREGROUP1,
		}, nil
	case *eddsaResharing.DGRound2Message:
		return blame.RoundInfo{
			Index:    2,
			RoundMsg: messages.EDDSAKEYREGROUP2,
		}, nil
	case *eddsaResharing.DGRound3Message1:
		return blame.RoundInfo{
			Index:    3,
			RoundMsg: messages.EDDSAKEYREGROUP3a,
		}, nil
	case *eddsaResharing.DGRound3Message2:
		return blame.RoundInfo{
			Index:    4,
			RoundMsg: messages.EDDSAKEYREGROUP3b,
		}, nil
	case *eddsaResharing.DGRound4Message:
		return blame.RoundInfo{
			Index:    5,
			RoundMsg: messages.EDDSAKEYREGROUP4,
		}, nil

	default:
		{
			return blame.RoundInfo{}, errors.New("unknown round")
		}
	}
}

func (t *TssCommon) NotifyTaskDone() error {
	msg := messages.TssTaskNotifier{TaskDone: true}
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("fail to marshal the request body %w", err)
	}
	wrappedMsg := messages.WrappedMessage{
		MessageType: messages.TSSTaskDone,
		MsgID:       t.msgID,
		Payload:     data,
	}
	t.P2PPeersLock.RLock()
	peers := t.P2PPeers
	t.P2PPeersLock.RUnlock()
	t.renderToP2P(&messages.BroadcastMsgChan{
		WrappedMessage: wrappedMsg,
		PeersID:        peers,
	})
	return nil
}

func (t *TssCommon) processRequestMsgFromPeer(peersID []peer.ID, msg *messages.TssControl, requester bool) error {
	// we need to send msg to the peer
	if !requester {
		if msg == nil {
			return errors.New("empty message")
		}
		reqKey := msg.ReqKey
		storedMsg := t.blameMgr.GetRoundMgr().Get(reqKey)
		if storedMsg == nil {
			t.logger.Debug().Msg("we do not have this message either")
			return nil
		}
		msg.Msg = storedMsg
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("fail to marshal the request body %w", err)
	}
	wrappedMsg := messages.WrappedMessage{
		MessageType: messages.TSSControlMsg,
		MsgID:       t.msgID,
		Payload:     data,
	}

	t.renderToP2P(&messages.BroadcastMsgChan{
		WrappedMessage: wrappedMsg,
		PeersID:        peersID,
	})
	return nil
}
