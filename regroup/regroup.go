package keyRegroup

import (
	bcrypto "github.com/binance-chain/tss-lib/crypto"
	"gitlab.com/thorchain/tss/go-tss/common"
	"gitlab.com/thorchain/tss/go-tss/p2p"
	"gitlab.com/thorchain/tss/go-tss/storage"
)

type TssKeyRegroup interface {
	GetTssKeyGenChannels() chan *p2p.Message
	GetTssCommonStruct() *common.TssCommon
	//NewPartyInit(req Request) (*btss.ReSharingParameters, *btss.ReSharingParameters, []*btss.PartyID, []*btss.PartyID, error)
	GenerateNewKey(req Request, localStateItem storage.KeygenLocalState) (*bcrypto.ECPoint, error)
}
