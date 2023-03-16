package tests

import (
	"fmt"
	"gitlab.com/thorchain/tss/go-tss/conversion"
	"testing"
)

func TestSimple(t *testing.T) {
	//sig := "406a67098d1b6bda83ddff7a78583cec0b393ba8086f0f3817a7ceb95cf48347072df68bf9fea5a476e764a75e8108b2af031c2aea3f346bfcd2e4524a4ac407"
	//pk := "c2558d89b470990234418e100238c00db88684ec0fc134848a0abb2b64ed4973"
	//test_msg := hex.EncodeToString([]byte("test"))
	//
	//key, _ := hex.DecodeString(pk)
	//pub, err := edwards.ParsePubKey(key)
	//fmt.Println("Error: ", err)
	//signature, _ := hex.DecodeString(sig)
	//newSig, err := edwards.ParseSignature(signature)
	//decMsg, _ := hex.DecodeString(test_msg)
	//fmt.Println("Result: ", edwards.Verify(pub, decMsg, newSig.R, newSig.S), ed25519.Verify(pub.SerializeCompressed(), decMsg, signature))
	//fmt.Println("Len msg: ", len(decMsg), "Len pk: ", len(pub.SerializeCompressed()), "Len sig: ", len(signature))
	res, _ := conversion.GetThreshold(5)
	fmt.Println("Result: ", res)
}
