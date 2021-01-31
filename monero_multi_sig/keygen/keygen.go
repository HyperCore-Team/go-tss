package keygen

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	bkg "github.com/binance-chain/tss-lib/ecdsa/keygen"
	btss "github.com/binance-chain/tss-lib/tss"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	tcrypto "github.com/tendermint/tendermint/crypto"

	moneroWallet "gitlab.com/thorchain/tss/monero-wallet-rpc/wallet"

	"gitlab.com/thorchain/tss/go-tss/common"
	"gitlab.com/thorchain/tss/go-tss/conversion"
	"gitlab.com/thorchain/tss/go-tss/messages"
	"gitlab.com/thorchain/tss/go-tss/p2p"
	"gitlab.com/thorchain/tss/go-tss/storage"
)

type MoneroKeyGen struct {
	logger             zerolog.Logger
	localNodePubKey    string
	preParams          *bkg.LocalPreParams
	moneroCommonStruct *common.TssCommon
	stopChan           chan struct{} // channel to indicate whether we should stop
	localParty         *btss.PartyID
	stateManager       storage.LocalStateManager
	commStopChan       chan struct{}
	p2pComm            *p2p.Communication
}

type MoneroSharesStore struct {
	shares map[int][]string
	locker sync.Mutex
}

func GenMoneroShareStore() *MoneroSharesStore {
	shares := make(map[int][]string)
	return &MoneroSharesStore{
		shares,
		sync.Mutex{},
	}
}

func (ms *MoneroSharesStore) storeAndCheck(round int, share string, checkLength int) ([]string, bool) {
	ms.locker.Lock()
	defer ms.locker.Unlock()
	shares, ok := ms.shares[round]
	if ok {
		shares = append(shares, share)
		ms.shares[round] = shares
		if len(shares) == checkLength {
			return shares, true
		}
		return shares, false
	}
	ms.shares[round] = []string{share}
	return ms.shares[round], false
}

func NewMoneroKeyGen(localP2PID string,
	conf common.TssConfig,
	localNodePubKey string,
	broadcastChan chan *messages.BroadcastMsgChan,
	stopChan chan struct{},
	preParam *bkg.LocalPreParams,
	msgID string,
	stateManager storage.LocalStateManager,
	privateKey tcrypto.PrivKey,
	p2pComm *p2p.Communication) *MoneroKeyGen {
	return &MoneroKeyGen{
		logger: log.With().
			Str("module", "keygen").
			Str("msgID", msgID).Logger(),
		localNodePubKey:    localNodePubKey,
		preParams:          preParam,
		moneroCommonStruct: common.NewTssCommon(localP2PID, broadcastChan, conf, msgID, privateKey),
		stopChan:           stopChan,
		localParty:         nil,
		stateManager:       stateManager,
		commStopChan:       make(chan struct{}),
		p2pComm:            p2pComm,
	}
}

func (tKeyGen *MoneroKeyGen) GetTssKeyGenChannels() chan *p2p.Message {
	return tKeyGen.moneroCommonStruct.TssMsg
}

func (tKeyGen *MoneroKeyGen) GetTssCommonStruct() *common.TssCommon {
	return tKeyGen.moneroCommonStruct
}

func (tKeyGen *MoneroKeyGen) packAndSend(info string, exchangeRound int, localPartyID *btss.PartyID, msgType string) error {
	sendShare := common.MoneroShare{
		MultisigInfo:  info,
		MsgType:       msgType,
		ExchangeRound: exchangeRound,
	}
	msg, err := json.Marshal(sendShare)
	if err != nil {
		tKeyGen.logger.Error().Err(err).Msg("fail to encode the wallet share")
		return err
	}

	r := btss.MessageRouting{
		From:        localPartyID,
		IsBroadcast: true,
	}
	return tKeyGen.moneroCommonStruct.ProcessOutCh(msg, &r, "moneroMsg", messages.TSSKeyGenMsg)
}

func (tKeyGen *MoneroKeyGen) GenerateNewKey(keygenReq Request) (string, error) {
	partiesID, localPartyID, err := conversion.GetParties(keygenReq.Keys, tKeyGen.localNodePubKey)
	if err != nil {
		return "", fmt.Errorf("fail to get keygen parties: %w", err)
	}

	threshold, err := conversion.GetThreshold(len(partiesID))
	if err != nil {
		return "", err
	}

	// now we try to connect to the monero wallet rpc client
	client := moneroWallet.New(moneroWallet.Config{
		Address: keygenReq.rpcAddress,
	})

	walletName := tKeyGen.localNodePubKey + tKeyGen.GetTssCommonStruct().GetMsgID() + ".mo"
	passcode := tKeyGen.GetTssCommonStruct().GetNodePrivKey()
	walletDat := moneroWallet.RequestCreateWallet{
		Filename: walletName,
		Password: passcode,
		Language: "English",
	}
	err = client.CreateWallet(&walletDat)
	if err != nil {
		return "", err
	}

	var keyGenWg sync.WaitGroup

	blameMgr := tKeyGen.moneroCommonStruct.GetBlameMgr()

	outCh := make(chan btss.Message, len(partiesID))
	endCh := make(chan bkg.LocalPartySaveData, len(partiesID))

	ctx := btss.NewPeerContext(partiesID)
	params := btss.NewParameters(ctx, localPartyID, len(partiesID), threshold)
	keyGenParty := bkg.NewLocalParty(params, outCh, endCh, *tKeyGen.preParams)
	partyIDMap := conversion.SetupPartyIDMap(partiesID)
	err1 := conversion.SetupIDMaps(partyIDMap, tKeyGen.moneroCommonStruct.PartyIDtoP2PID)
	err2 := conversion.SetupIDMaps(partyIDMap, blameMgr.PartyIDtoP2PID)
	if err1 != nil || err2 != nil {
		tKeyGen.logger.Error().Msgf("error in creating mapping between partyID and P2P ID")
		return "", err
	}

	partyInfo := &common.PartyInfo{
		Party:      keyGenParty,
		PartyIDMap: partyIDMap,
	}

	tKeyGen.moneroCommonStruct.SetPartyInfo(partyInfo)
	blameMgr.SetPartyInfo(keyGenParty, partyIDMap)
	tKeyGen.moneroCommonStruct.P2PPeers = conversion.GetPeersID(tKeyGen.moneroCommonStruct.PartyIDtoP2PID, tKeyGen.moneroCommonStruct.GetLocalPeerID())
	keyGenWg.Add(1)
	// start keygen
	defer tKeyGen.logger.Debug().Msg("generate monero share")

	moneroShareChan := make(chan *common.MoneroShare, len(partiesID))

	var address string
	go func() {
		tKeyGen.moneroCommonStruct.ProcessInboundMessages(tKeyGen.commStopChan, &keyGenWg, moneroShareChan)
	}()

	share, err := client.PrepareMultisig()
	if err != nil {
		return "", err
	}

	var exchangeRound int32
	exchangeRound = 0
	err = tKeyGen.packAndSend(share.MultisigInfo, int(exchangeRound), localPartyID, common.MoneroSharepre)
	if err != nil {
		return "", err
	}
	exchangeRound += 1

	var globalErr error
	peerNum := len(partiesID) - 1
	shareStore := GenMoneroShareStore()
	keyGenWg.Add(1)
	go func() {
		defer keyGenWg.Done()
		for {
			select {
			case <-time.After(time.Minute * 10):
				close(tKeyGen.commStopChan)

			case share := <-moneroShareChan:
				switch share.MsgType {
				case common.MoneroSharepre:
					shares, ready := shareStore.storeAndCheck(int(exchangeRound)-1, share.MultisigInfo, peerNum)
					if !ready {
						continue
					}
					request := moneroWallet.RequestMakeMultisig{
						MultisigInfo: shares,
						Threshold:    uint64(threshold),
						Password:     passcode,
					}
					resp, err := client.MakeMultisig(&request)
					if err != nil {
						globalErr = err
						return
					}

					currentRound := atomic.LoadInt32(&exchangeRound)
					err = tKeyGen.packAndSend(resp.MultisigInfo, int(currentRound), localPartyID, common.MoneroKeyGenShareExchange)
					if err != nil {
						globalErr = err
						return
					}
					atomic.AddInt32(&exchangeRound, 1)

				case common.MoneroKeyGenShareExchange:
					shares, ready := shareStore.storeAndCheck(int(exchangeRound)-1, share.MultisigInfo, peerNum)
					if !ready {
						continue
					}

					finRequest := moneroWallet.RequestExchangeMultisigKeys{
						MultisigInfo: shares,
						Password:     passcode,
					}
					resp, err := client.ExchangeMultiSigKeys(&finRequest)
					if err != nil {
						globalErr = err
						return
					}
					// this indicate the wallet address is generated
					if len(resp.Address) != 0 {
						address = resp.Address
						err = tKeyGen.moneroCommonStruct.NotifyTaskDone()
						if err != nil {
							tKeyGen.logger.Error().Err(err).Msg("fail to broadcast the keysign done")
						}
						close(tKeyGen.commStopChan)
						return
					}

					currentRound := atomic.LoadInt32(&exchangeRound)
					err = tKeyGen.packAndSend(resp.MultisigInfo, int(currentRound), localPartyID, common.MoneroKeyGenShareExchange)
					if err != nil {
						globalErr = err
						return
					}
					atomic.AddInt32(&exchangeRound, 1)
				}
			case <-tKeyGen.moneroCommonStruct.GetTaskDone():
				close(tKeyGen.commStopChan)
			}
		}
	}()

	keyGenWg.Wait()
	if globalErr != nil {
		tKeyGen.logger.Error().Err(err).Msg("fail to create the monero multisig wallet")
	}
	tKeyGen.logger.Info().Msgf("wallet address is  %v\n", address)
	return address, err
}