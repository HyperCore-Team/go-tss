package eddsa

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/binance-chain/tss-lib/eddsa/keygen"

	"gitlab.com/thorchain/tss/go-tss/regroup"

	bcrypto "github.com/binance-chain/tss-lib/crypto"
	bkg "github.com/binance-chain/tss-lib/eddsa/keygen"
	bkr "github.com/binance-chain/tss-lib/eddsa/resharing"
	btss "github.com/binance-chain/tss-lib/tss"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	tcrypto "github.com/tendermint/tendermint/crypto"

	"gitlab.com/thorchain/tss/go-tss/blame"
	"gitlab.com/thorchain/tss/go-tss/common"
	"gitlab.com/thorchain/tss/go-tss/conversion"
	"gitlab.com/thorchain/tss/go-tss/messages"
	"gitlab.com/thorchain/tss/go-tss/p2p"
	"gitlab.com/thorchain/tss/go-tss/storage"
)

type TssKeyReGroup struct {
	logger          zerolog.Logger
	localNodePubKey string
	tssCommonStruct *common.TssCommon
	stopChan        chan struct{} // channel to indicate whether we should stop
	localParty      *btss.PartyID
	stateManager    storage.LocalStateManager
	commStopChan    chan struct{}
	p2pComm         *p2p.Communication
}

func NewTssKeyReGroup(localP2PID string,
	conf common.TssConfig,
	localNodePubKey string,
	broadcastChan chan *messages.BroadcastMsgChan,
	stopChan chan struct{},
	msgID string,
	stateManager storage.LocalStateManager,
	privateKey tcrypto.PrivKey,
	p2pComm *p2p.Communication) *TssKeyReGroup {
	return &TssKeyReGroup{
		logger: log.With().
			Str("module", "regroup").
			Str("msgID", msgID).Logger(),
		localNodePubKey: localNodePubKey,
		tssCommonStruct: common.NewTssCommon(localP2PID, broadcastChan, conf, msgID, privateKey, 1),
		stopChan:        stopChan,
		localParty:      nil,
		stateManager:    stateManager,
		commStopChan:    make(chan struct{}),
		p2pComm:         p2pComm,
	}
}

func (tKeyReGroup *TssKeyReGroup) GetTssKeyGenChannels() chan *p2p.Message {
	return tKeyReGroup.tssCommonStruct.TssMsg
}

func (tKeyReGroup *TssKeyReGroup) GetTssCommonStruct() *common.TssCommon {
	return tKeyReGroup.tssCommonStruct
}

func (tKeyReGroup *TssKeyReGroup) NewPartyInit(req keyRegroup.Request) (*btss.ReSharingParameters, *btss.ReSharingParameters, []*btss.PartyID, []*btss.PartyID, error) {
	var newPartiesID, oldPartiesID []*btss.PartyID
	var newLocalPartyID, oldLocalPartyID *btss.PartyID
	amNewParty := false
	for _, el := range req.NewPartyKeys {
		if tKeyReGroup.localNodePubKey == el {
			amNewParty = true
			break
		}
	}

	// we are the new party
	threshold, err := conversion.GetThreshold(len(req.NewPartyKeys))
	if err != nil {
		return nil, nil, nil, nil, errors.New("fail to get threshold")
	}

	if amNewParty {
		newPartiesID, newLocalPartyID, err = conversion.GetParties(req.NewPartyKeys, tKeyReGroup.localNodePubKey, true, "new_party")
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("fail to initlize the new parties: %w", err)
		}
		oldPartiesID, oldLocalPartyID, _ = conversion.GetParties(req.OldPartyKeys, tKeyReGroup.localNodePubKey, false, "old_party")
	} else {
		newPartiesID, _, _ = conversion.GetParties(req.NewPartyKeys, tKeyReGroup.localNodePubKey, false, "new_party")
		oldPartiesID, oldLocalPartyID, err = conversion.GetParties(req.OldPartyKeys, tKeyReGroup.localNodePubKey, true, "old_party")
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("old committee fails to initlize the new parties: %w", err)
		}
	}

	ctxNew := btss.NewPeerContext(newPartiesID)
	ctxOld := btss.NewPeerContext(oldPartiesID)
	var newParams, oldParams *btss.ReSharingParameters
	if newLocalPartyID != nil {
		newParams = btss.NewReSharingParameters(btss.Edwards(), ctxOld, ctxNew, newLocalPartyID, len(req.OldPartyKeys), threshold, len(req.NewPartyKeys), threshold)
	}
	if oldLocalPartyID != nil {
		oldParams = btss.NewReSharingParameters(btss.Edwards(), ctxOld, ctxNew, oldLocalPartyID, len(req.OldPartyKeys), threshold, len(req.NewPartyKeys), threshold)
	}

	return newParams, oldParams, oldPartiesID, newPartiesID, nil
}

func (tKeyReGroup *TssKeyReGroup) GenerateNewKey(req keyRegroup.Request, localStateItem storage.KeygenLocalState) (*bcrypto.ECPoint, error) {
	var newParams, oldParams *btss.ReSharingParameters
	var newPartiesID, oldPartiesID []*btss.PartyID
	var newKeyGenParty, oldKeyGenParty btss.Party
	var err error

	allParties := make(map[string]bool)

	for _, el := range req.OldPartyKeys {
		allParties[el] = true
	}

	for _, el := range req.NewPartyKeys {
		allParties[el] = true
	}
	partyNum := len(allParties)

	newParams, oldParams, oldPartiesID, newPartiesID, err = tKeyReGroup.NewPartyInit(req)
	if err != nil {
		tKeyReGroup.logger.Error().Err(err).Msgf("fail to init the party")
		return nil, err
	}
	keyGenPartyMap := new(sync.Map)
	outCh := make(chan btss.Message, partyNum*2)
	endCh := make(chan bkg.LocalPartySaveData, partyNum*2)
	errChan := make(chan struct{})
	blameMgr := tKeyReGroup.tssCommonStruct.GetBlameMgr()
	var localData keygen.LocalPartySaveData
	err = json.Unmarshal(localStateItem.LocalData, &localData)
	if newParams != nil {
		newKeyGenParty = bkr.NewLocalParty(newParams, localData, outCh, endCh)
		newKeyGenParty.PartyID().Moniker = common.NewParty
	}
	if oldParams != nil {
		oldKeyGenParty = bkr.NewLocalParty(oldParams, localData, outCh, endCh)
		oldKeyGenParty.PartyID().Moniker = common.OldParty
	}

	allPartiesID := append(newPartiesID, oldPartiesID...)
	partyIDMap := conversion.SetupPartyIDMap(allPartiesID)
	oldPartyIDMap := conversion.SetupPartyIDMap(oldPartiesID)
	newPartyIDMap := conversion.SetupPartyIDMap(newPartiesID)

	err1 := conversion.SetupIDMaps(partyIDMap, tKeyReGroup.tssCommonStruct.PartyIDtoP2PID)
	err2 := conversion.SetupIDMaps(partyIDMap, blameMgr.PartyIDtoP2PID)
	if err1 != nil || err2 != nil {
		tKeyReGroup.logger.Error().Msgf("error in creating mapping between partyID and P2P ID")
		return nil, err
	}
	// we never run multi keygen, so the moniker is set to default empty value
	if oldKeyGenParty != nil {
		keyGenPartyMap.Store("old_party", oldKeyGenParty)
	}
	if newKeyGenParty != nil {
		keyGenPartyMap.Store("new_party", newKeyGenParty)
	}

	partyInfo := &common.PartyInfo{
		PartyMap:      keyGenPartyMap,
		PartyIDMap:    partyIDMap,
		OldPartyIDMap: oldPartyIDMap,
		NewPartyIDMap: newPartyIDMap,
	}
	tKeyReGroup.tssCommonStruct.SetPartyInfo(partyInfo)
	blameMgr.SetPartyInfo(keyGenPartyMap, partyIDMap)
	tKeyReGroup.tssCommonStruct.P2PPeersLock.Lock()
	tKeyReGroup.tssCommonStruct.P2PPeers = conversion.GetPeersID(tKeyReGroup.tssCommonStruct.PartyIDtoP2PID, tKeyReGroup.tssCommonStruct.GetLocalPeerID())
	tKeyReGroup.tssCommonStruct.P2PPeersLock.Unlock()
	var keyGenWg sync.WaitGroup

	if oldKeyGenParty != nil {
		keyGenWg.Add(1)
		// start keygen
		go func() {
			defer keyGenWg.Done()
			defer tKeyReGroup.logger.Debug().Msgf(">>>>>>>>>>>>> party regroup started===%v\n", tKeyReGroup.localNodePubKey)
			if err := oldKeyGenParty.Start(); nil != err {
				tKeyReGroup.logger.Error().Err(err).Msg("fail to start party regroup party")
				close(errChan)
			}
		}()
	}
	if newKeyGenParty != nil {
		keyGenWg.Add(1)
		// start keygen
		go func() {
			defer keyGenWg.Done()
			defer tKeyReGroup.logger.Debug().Msgf(">>>>>>>>>>>>>  party regroup started  ===%v\n", tKeyReGroup.localNodePubKey)
			if err := newKeyGenParty.Start(); nil != err {
				tKeyReGroup.logger.Error().Err(err).Msg("fail to start party regroup party")
				close(errChan)
			}
		}()
	}

	keyGenLocalStateItem := storage.KeygenLocalState{
		ParticipantKeys: req.NewPartyKeys,
		LocalPartyKey:   tKeyReGroup.localNodePubKey,
	}

	keyGenWg.Add(1)
	go tKeyReGroup.tssCommonStruct.ProcessInboundMessages(tKeyReGroup.commStopChan, &keyGenWg)

	r, err, _ := tKeyReGroup.processKeyReGroup(errChan, outCh, endCh, oldKeyGenParty != nil && newKeyGenParty != nil, keyGenLocalStateItem, len(req.OldPartyKeys))
	if err != nil {
		close(tKeyReGroup.commStopChan)
		return nil, fmt.Errorf("fail to process key sign: %w", err)
	}
	select {
	case <-time.After(time.Second * 5):
		close(tKeyReGroup.commStopChan)

	case <-tKeyReGroup.tssCommonStruct.GetTaskDone():
		close(tKeyReGroup.commStopChan)
	}

	keyGenWg.Wait()
	return r, err
}

func (tKeyReGroup *TssKeyReGroup) processKeyReGroup(errChan chan struct{},
	outCh <-chan btss.Message,
	endCh <-chan bkg.LocalPartySaveData, bothOldNewParty bool, keyGenLocalStateItem storage.KeygenLocalState, oldPartyNum int,
) (*bcrypto.ECPoint, error, string) {
	// keyGenLocalStateItem storage.KeygenLocalState) (*bcrypto.ECPoint, error) {
	defer tKeyReGroup.logger.Debug().Msg("finished regroup process")
	tKeyReGroup.logger.Debug().Msg("start to read messages from local party")
	tssConf := tKeyReGroup.tssCommonStruct.GetConf()
	blameMgr := tKeyReGroup.tssCommonStruct.GetBlameMgr()
	var strPubKey string
	for {
		select {
		case <-errChan: // when keyGenParty return
			tKeyReGroup.logger.Error().Msg("regroup failed")
			return nil, errors.New("error channel closed fail to start local party"), ""

		case <-tKeyReGroup.stopChan: // when TSS processor receive signal to quit
			return nil, errors.New("received exit signal"), ""

		case <-time.After(tssConf.KeyRegroupTimeout):
			// we bail out after KeyRegroupTimeoutSeconds
			tKeyReGroup.logger.Error().Msgf("fail to generate message with %s", tssConf.KeyRegroupTimeout.String())
			lastMsg := blameMgr.GetLastMsg()
			failReason := blameMgr.GetBlame().FailReason
			if failReason == "" {
				failReason = blame.TssTimeout
			}
			if lastMsg == nil {
				tKeyReGroup.logger.Error().Msg("fail to start the party regroup, the last produced message of this node is none")
				return nil, errors.New("timeout before shared message is generated"), ""
			}
			blameNodesUnicast, err := blameMgr.GetUnicastBlame(messages.EDDSAKEYREGROUP3a)
			if err != nil {
				tKeyReGroup.logger.Error().Err(err).Msg("error in get unicast blame")
			}
			tKeyReGroup.tssCommonStruct.P2PPeersLock.RLock()
			threshold, err := conversion.GetThreshold(len(tKeyReGroup.tssCommonStruct.P2PPeers) + 1)
			tKeyReGroup.tssCommonStruct.P2PPeersLock.RUnlock()
			if err != nil {
				tKeyReGroup.logger.Error().Err(err).Msg("error in get the threshold to generate blame")
			}

			if len(blameNodesUnicast) > 0 && len(blameNodesUnicast) <= threshold {
				blameMgr.GetBlame().SetBlame(failReason, blameNodesUnicast, true)
			}
			blameNodesBroadcast, err := blameMgr.GetBroadcastBlame(lastMsg.Type())
			if err != nil {
				tKeyReGroup.logger.Error().Err(err).Msg("error in get broadcast blame")
			}
			blameMgr.GetBlame().AddBlameNodes(blameNodesBroadcast...)

			// if we cannot find the blame node, we check whether everyone send me the share
			if len(blameMgr.GetBlame().BlameNodes) == 0 {
				blameNodesMisingShare, isUnicast, err := blameMgr.TssMissingShareBlame(messages.EDDSAREGROUPROUNDS, messages.EDDSAKEYREGROUP)
				if err != nil {
					tKeyReGroup.logger.Error().Err(err).Msg("fail to get the node of missing share ")
				}
				if len(blameNodesMisingShare) > 0 && len(blameNodesMisingShare) <= threshold {
					blameMgr.GetBlame().AddBlameNodes(blameNodesMisingShare...)
					blameMgr.GetBlame().IsUnicast = isUnicast
				}
			}
			return nil, blame.ErrTssTimeOut, ""

		case msg := <-outCh:
			blameMgr.SetLastMsg(msg)
			// fixme since we change the logic of party.ID,the party order is incorrect for testing we need to regenerate the testing data.
			dest := msg.GetTo()
			if dest == nil {
				return nil, errors.New("did not expect a msg to have a nil destination during resharing"), ""
			}
			// to old members
			if msg.IsToOldCommittee() {
				err := tKeyReGroup.tssCommonStruct.ProcessRegroupOutCh(msg, messages.TSSPartyReGroupMsg, common.OldParty)
				if err != nil {
					tKeyReGroup.logger.Error().Err(err).Msg("fail to process the message")
					return nil, err, ""
				}
				continue
			}
			// to new members
			if !msg.IsToOldCommittee() && !msg.IsToOldAndNewCommittees() {
				err := tKeyReGroup.tssCommonStruct.ProcessRegroupOutCh(msg, messages.TSSPartyReGroupMsg, common.NewParty)
				if err != nil {
					tKeyReGroup.logger.Error().Err(err).Msg("fail to process the message")
					return nil, err, ""
				}
				continue
			}

			if msg.IsToOldAndNewCommittees() {
				messageRoutingOld := btss.MessageRouting{
					From:                    msg.GetFrom(),
					To:                      msg.GetTo()[:oldPartyNum],
					IsBroadcast:             msg.IsBroadcast(),
					IsToOldCommittee:        msg.IsToOldCommittee(),
					IsToOldAndNewCommittees: msg.IsToOldAndNewCommittees(),
				}

				messageRoutingNew := btss.MessageRouting{
					From:                    msg.GetFrom(),
					To:                      msg.GetTo()[oldPartyNum:],
					IsBroadcast:             msg.IsBroadcast(),
					IsToOldCommittee:        msg.IsToOldCommittee(),
					IsToOldAndNewCommittees: msg.IsToOldAndNewCommittees(),
				}

				msgByte, _, err := msg.WireBytes()
				if err != nil {
					tKeyReGroup.logger.Error().Err(err).Msgf("fail to parse the message")
				}
				data, err := btss.ParseWireMessage(msgByte, msg.GetFrom(), msg.IsBroadcast())
				if err != nil {
					tKeyReGroup.logger.Error().Err(err).Msgf("fail to parse the message")
				}

				oldTssMsg := btss.NewMessage(messageRoutingOld, data.Content(), msg.WireMsg())
				newTssMsg := btss.NewMessage(messageRoutingNew, data.Content(), msg.WireMsg())

				err = tKeyReGroup.tssCommonStruct.ProcessRegroupOutCh(oldTssMsg, messages.TSSPartyReGroupMsg, common.OldParty)
				if err != nil {
					tKeyReGroup.logger.Error().Err(err).Msg("fail to process the message")
					return nil, err, ""
				}
				err = tKeyReGroup.tssCommonStruct.ProcessRegroupOutCh(newTssMsg, messages.TSSPartyReGroupMsg, common.NewParty)
				if err != nil {
					tKeyReGroup.logger.Error().Err(err).Msg("fail to process the message")
					return nil, err, ""
				}
			}

		case msg := <-endCh:
			if msg.Xi != nil {
				_, err := msg.OriginalIndex()
				if err != nil {
					tKeyReGroup.logger.Error().Err(err).Msgf("fail to get the index of the message")
				}
				strPubKey, err = conversion.GetTssPubKeyEDDSA(msg.EDDSAPub)
				if err != nil {
					return nil, fmt.Errorf("fail to get thorchain pubkey: %w", err), ""
				}

				marshaledMsg, err := json.Marshal(msg)
				if err != nil {
					tKeyReGroup.logger.Error().Err(err).Msg("fail to marshal the result")
					return nil, errors.New("fail to marshal the result"), ""
				}
				keyGenLocalStateItem.LocalData = marshaledMsg
				keyGenLocalStateItem.PubKey = strPubKey
				if err := tKeyReGroup.stateManager.SaveLocalState(keyGenLocalStateItem, messages.EDDSAKEYREGROUP); err != nil {
					return nil, fmt.Errorf("fail to save keygen result to storage: %w", err), ""
				}
				address := tKeyReGroup.p2pComm.ExportPeerAddress()
				if err := tKeyReGroup.stateManager.SaveAddressBook(address); err != nil {
					tKeyReGroup.logger.Error().Err(err).Msg("fail to save the peer addresses")
				}

			} else {
				if bothOldNewParty {
					// we need to wait for the new party instance to return
					continue
				}
			}
			err := tKeyReGroup.tssCommonStruct.NotifyTaskDone()
			if err != nil {
				tKeyReGroup.logger.Error().Err(err).Msg("fail to broadcast the keysign done")
			}
			return msg.EDDSAPub, nil, strPubKey
		}
	}
}
