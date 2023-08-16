package keysign

import (
	bc "github.com/binance-chain/tss-lib/common"

	"github.com/HyperCore-Team/go-tss/common"
	"github.com/HyperCore-Team/go-tss/p2p"
	"github.com/HyperCore-Team/go-tss/storage"
)

type TssKeySign interface {
	GetTssKeySignChannels() chan *p2p.Message
	GetTssCommonStruct() *common.TssCommon
	SignMessage(msgToSign [][]byte, localStateItem storage.KeygenLocalState, parties []string) ([]*bc.SignatureData, error)
}
