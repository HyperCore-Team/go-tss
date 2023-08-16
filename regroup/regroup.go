package keyRegroup

import (
	"github.com/HyperCore-Team/go-tss/common"
	"github.com/HyperCore-Team/go-tss/p2p"
	"github.com/HyperCore-Team/go-tss/storage"
	bcrypto "github.com/HyperCore-Team/tss-lib/crypto"
)

type TssKeyRegroup interface {
	GetTssKeyGenChannels() chan *p2p.Message
	GetTssCommonStruct() *common.TssCommon
	//NewPartyInit(req Request) (*btss.ReSharingParameters, *btss.ReSharingParameters, []*btss.PartyID, []*btss.PartyID, error)
	GenerateNewKey(req Request, localStateItem storage.KeygenLocalState) (*bcrypto.ECPoint, error)
}
