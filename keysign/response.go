package keysign

import (
	"github.com/HyperCore-Team/go-tss/blame"
	"github.com/HyperCore-Team/go-tss/common"
)

// signature
type Signature struct {
	Msg        string `json:"signed_msg"`
	R          string `json:"r"`
	S          string `json:"s"`
	RecoveryID string `json:"recovery_id"`
	Signature  string `json:"signature"`
}

// Response key sign response
type Response struct {
	Signatures []Signature   `json:"signatures"`
	Status     common.Status `json:"status"`
	Blame      blame.Blame   `json:"blame"`
}

func NewSignature(msg, r, s, recoveryID string, signature string) Signature {
	return Signature{
		Msg:        msg,
		R:          r,
		S:          s,
		RecoveryID: recoveryID,
		Signature:  signature,
	}
}

func NewResponse(signatures []Signature, status common.Status, blame blame.Blame) Response {
	return Response{
		Signatures: signatures,
		Status:     status,
		Blame:      blame,
	}
}
