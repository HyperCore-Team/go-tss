package p2p

import (
	"math/rand"
	"sort"
	"sync"
	"testing"
	"time"

	tnet "github.com/libp2p/go-libp2p-testing/net"
	"github.com/libp2p/go-libp2p/core/host"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/stretchr/testify/assert"

	"github.com/HyperCore-Team/go-tss/conversion"
)

func setupHostsLocally(t *testing.T, n int) []host.Host {
	mn := mocknet.New()
	var hosts []host.Host
	for i := 0; i < n; i++ {

		id := tnet.RandIdentityOrFatal(t)
		a := tnet.RandLocalTCPAddress()
		h, err := mn.AddPeer(id.PrivateKey(), a)
		if err != nil {
			t.Fatal(err)
		}
		hosts = append(hosts, h)
	}

	if err := mn.LinkAll(); err != nil {
		t.Error(err)
	}
	if err := mn.ConnectAllButSelf(); err != nil {
		t.Error(err)
	}
	return hosts
}

func TestPartyCoordinator(t *testing.T) {
	ApplyDeadline = false
	hosts := setupHostsLocally(t, 4)
	var pcs []PartyCoordinator
	var peers []string

	whitelist := make(map[string]bool)
	for _, el := range hosts {
		whitelist[el.ID().String()] = true
	}

	timeout := time.Second * 10
	for _, el := range hosts {
		pcs = append(pcs, *NewPartyCoordinator(el, nil, timeout, whitelist))
		peers = append(peers, el.ID().String())
	}

	defer func() {
		for _, el := range pcs {
			el.Stop()
		}
	}()

	msgID := conversion.RandStringBytesMask(64)
	wg := sync.WaitGroup{}

	for _, el := range pcs {
		wg.Add(1)

		go func(coordinator PartyCoordinator) {
			defer wg.Done()
			// we simulate different nodes join at different time
			time.Sleep(time.Second * time.Duration(rand.Int()%10))
			onlinePeers, err := coordinator.JoinPartyWithRetry(msgID, peers)
			if err != nil {
				t.Error(err)
			}
			assert.Nil(t, err)
			assert.Len(t, onlinePeers, 4)
		}(el)
	}

	wg.Wait()
}

func TestPartyCoordinatorTimeOut(t *testing.T) {
	ApplyDeadline = false
	timeout := time.Second
	hosts := setupHosts(t, 4)
	var pcs []*PartyCoordinator
	var peers []string
	whitelist := map[string]bool{}
	whitelist["12D3KooWE4qDcRrueTuRYWUdQZgcy7APZqBngVeXRt4Y6ytHizKV"] = true
	whitelist["12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu"] = true
	whitelist["12D3KooWKRyzVWW6ChFjQjK4miCty85Niy49tpPV95XdKu1BcvMA"] = true
	for _, el := range hosts {
		pcs = append(pcs, NewPartyCoordinator(el, nil, timeout, whitelist))
	}
	sort.Slice(pcs, func(i, j int) bool {
		return pcs[i].host.ID().String() > pcs[j].host.ID().String()
	})
	for _, el := range pcs {
		peers = append(peers, el.host.ID().String())
	}

	defer func() {
		for _, el := range pcs {
			el.Stop()
		}
	}()

	msgID := conversion.RandStringBytesMask(64)
	wg := sync.WaitGroup{}
	expected := peers[:2]
	sort.Strings(expected)

	for _, el := range pcs[:2] {
		wg.Add(1)
		go func(coordinator *PartyCoordinator) {
			defer wg.Done()
			onlinePeers, err := coordinator.JoinPartyWithRetry(msgID, peers)
			assert.Errorf(t, err, ErrJoinPartyTimeout.Error())
			var onlinePeersStr []string
			for _, el := range onlinePeers {
				onlinePeersStr = append(onlinePeersStr, el.String())
			}
			sort.Strings(onlinePeersStr)
			assert.EqualValues(t, onlinePeersStr, expected)
		}(el)
	}

	wg.Wait()
}
