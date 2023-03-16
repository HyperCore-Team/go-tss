package keygen

import (
	"gitlab.com/thorchain/tss/go-tss/blame"
	"gitlab.com/thorchain/tss/go-tss/common"
)

// Response keygen response
type Response struct {
	PubKey    string        `json:"pub_key"`
	Status    common.Status `json:"status"`
	Blame     blame.Blame   `json:"blame"`
	StrPubKey string        `json:"str_pub_key"`
	Threshold int           `json:"threshold"`
}

// NewResponse create a new instance of keygen.Response
func NewResponse(pk string, status common.Status, blame blame.Blame, strPubKey string,
	threshold int) Response {
	return Response{
		PubKey:    pk,
		Status:    status,
		Blame:     blame,
		StrPubKey: strPubKey,
		Threshold: threshold,
	}
}
