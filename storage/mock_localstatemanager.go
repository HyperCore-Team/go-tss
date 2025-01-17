package storage

import (
	"github.com/HyperCore-Team/go-tss/messages"
	"github.com/libp2p/go-libp2p/core/peer"
	maddr "github.com/multiformats/go-multiaddr"
)

// MockLocalStateManager is a mock use for test purpose
type MockLocalStateManager struct {
}

func (s *MockLocalStateManager) SaveLocalState(state KeygenLocalState, algo messages.Algo) error {
	return nil
}

func (s *MockLocalStateManager) GetLocalState(pubKey string, algo messages.Algo) (KeygenLocalState, error) {
	return KeygenLocalState{}, nil
}

func (s *MockLocalStateManager) SaveAddressBook(address map[peer.ID][]maddr.Multiaddr) error {
	return nil
}

func (s *MockLocalStateManager) RetrieveP2PAddresses() ([]maddr.Multiaddr, error) {
	return nil, nil
}
