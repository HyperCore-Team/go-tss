package keygen

import (
	bcrypto "github.com/HyperCore-Team/tss-lib/crypto"

	"github.com/HyperCore-Team/go-tss/common"
	"github.com/HyperCore-Team/go-tss/p2p"
)

type TssKeyGen interface {
	GenerateNewKey(keygenReq Request) (*bcrypto.ECPoint, error)
	GetTssKeyGenChannels() chan *p2p.Message
	GetTssCommonStruct() *common.TssCommon
}
