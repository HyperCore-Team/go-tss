package storage

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	maddr "github.com/multiformats/go-multiaddr"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/HyperCore-Team/go-tss/messages"

	"github.com/HyperCore-Team/go-tss/conversion"
	"github.com/libp2p/go-libp2p/core/peer"
)

// KeygenLocalState is a structure used to represent the data we saved locally for different keygen
type KeygenLocalState struct {
	PubKey          string   `json:"pub_key"`
	LocalData       []byte   `json:"local_data"`
	ParticipantKeys []string `json:"participant_keys"` // the paticipant of last key gen
	LocalPartyKey   string   `json:"local_party_key"`
}

// LocalStateManager provide necessary methods to manage the local state, save it , and read it back
// LocalStateManager doesn't have any opinion in regards to where it should be persistent to
type LocalStateManager interface {
	SaveLocalState(state KeygenLocalState, algo messages.Algo) error
	GetLocalState(pubKey string, algo messages.Algo) (KeygenLocalState, error)
	SaveAddressBook(addressBook map[peer.ID][]maddr.Multiaddr) error
	RetrieveP2PAddresses() ([]maddr.Multiaddr, error)
}

// FileStateMgr save the local state to file
type FileStateMgr struct {
	folder    string
	writeLock *sync.RWMutex
}

// NewFileStateMgr create a new instance of the FileStateMgr which implements LocalStateManager
func NewFileStateMgr(folder string) (*FileStateMgr, error) {
	if len(folder) > 0 {
		_, err := os.Stat(folder)
		if err != nil && os.IsNotExist(err) {
			if err := os.MkdirAll(folder, os.ModePerm); err != nil {
				return nil, err
			}
		}
	}
	return &FileStateMgr{
		folder:    folder,
		writeLock: &sync.RWMutex{},
	}, nil
}

func (fsm *FileStateMgr) getFilePathName(pubKey string, algo messages.Algo) (string, error) {
	ret, err := conversion.CheckKeyOnCurve(pubKey, algo)
	if err != nil {
		return "", err
	}
	if !ret {
		return "", errors.New("invalid pubkey for file name")
	}
	pubKeyBytes, err := base64.StdEncoding.DecodeString(pubKey)
	if err != nil {
		return "", err
	}
	hx := hex.EncodeToString(pubKeyBytes)
	localFileName := fmt.Sprintf("localstate-%s.json", hx)
	if len(fsm.folder) > 0 {
		return filepath.Join(fsm.folder, localFileName), nil
	}
	return localFileName, nil
}

// SaveLocalState save the local state to file
func (fsm *FileStateMgr) SaveLocalState(state KeygenLocalState, algo messages.Algo) error {
	buf, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("fail to marshal KeygenLocalState to json: %w", err)
	}
	filePathName, err := fsm.getFilePathName(state.PubKey, algo)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePathName, buf, 0o655)
}

// GetLocalState read the local state from file system
func (fsm *FileStateMgr) GetLocalState(pubKey string, algo messages.Algo) (KeygenLocalState, error) {
	if len(pubKey) == 0 {
		return KeygenLocalState{}, errors.New("pub key is empty")
	}
	filePathName, err := fsm.getFilePathName(pubKey, algo)
	if err != nil {
		return KeygenLocalState{}, err
	}
	if _, err := os.Stat(filePathName); os.IsNotExist(err) {
		return KeygenLocalState{}, err
	}

	buf, err := ioutil.ReadFile(filePathName)
	if err != nil {
		return KeygenLocalState{}, fmt.Errorf("file to read from file(%s): %w", filePathName, err)
	}
	var localState KeygenLocalState
	if err := json.Unmarshal(buf, &localState); nil != err {
		return KeygenLocalState{}, fmt.Errorf("fail to unmarshal KeygenLocalState: %w", err)
	}
	return localState, nil
}

func (fsm *FileStateMgr) SaveAddressBook(address map[peer.ID][]maddr.Multiaddr) error {
	if len(fsm.folder) < 1 {
		return errors.New("base file path is invalid")
	}
	filePathName := filepath.Join(fsm.folder, "address_book.seed")
	var buf bytes.Buffer

	for peer, addrs := range address {
		for _, addr := range addrs {
			// we do not save the loopback addr
			if strings.Contains(addr.String(), "127.0.0.1") {
				continue
			}
			record := addr.String() + "/p2p/" + peer.String() + "\n"
			_, err := buf.WriteString(record)
			if err != nil {
				return errors.New("fail to write the record to buffer")
			}
		}
	}
	fsm.writeLock.Lock()
	defer fsm.writeLock.Unlock()
	return ioutil.WriteFile(filePathName, buf.Bytes(), 0o655)
}

func (fsm *FileStateMgr) RetrieveP2PAddresses() ([]maddr.Multiaddr, error) {
	if len(fsm.folder) < 1 {
		return nil, errors.New("base file path is invalid")
	}
	filePathName := filepath.Join(fsm.folder, "address_book.seed")

	_, err := os.Stat(filePathName)
	if err != nil {
		return nil, err
	}
	fsm.writeLock.RLock()
	input, err := ioutil.ReadFile(filePathName)
	if err != nil {
		fsm.writeLock.RUnlock()
		return nil, err
	}
	fsm.writeLock.RUnlock()
	data := strings.Split(string(input), "\n")
	var peerAddresses []maddr.Multiaddr
	for _, el := range data {
		// we skip the empty entry
		if len(el) == 0 {
			continue
		}
		addr, err := maddr.NewMultiaddr(el)
		if err != nil {
			return nil, fmt.Errorf("invalid address in address book %w", err)
		}
		peerAddresses = append(peerAddresses, addr)
	}
	return peerAddresses, nil
}
