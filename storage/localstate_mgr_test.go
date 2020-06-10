package storage

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/tss/go-tss/conversion"
)

type FileStateMgrTestSuite struct{}

var _ = Suite(&FileStateMgrTestSuite{})

func TestPackage(t *testing.T) { TestingT(t) }

func (s *FileStateMgrTestSuite) SetUpTest(c *C) {
	conversion.SetupBech32Prefix()
}

func (s *FileStateMgrTestSuite) TestNewFileStateMgr(c *C) {
	folder := os.TempDir()
	f := filepath.Join(folder, "test", "test1", "test2")
	defer func() {
		err := os.RemoveAll(f)
		c.Assert(err, IsNil)
	}()
	fsm, err := NewFileStateMgr(f)
	c.Assert(err, IsNil)
	c.Assert(fsm, NotNil)
	_, err = os.Stat(f)
	c.Assert(err, IsNil)
	fileName, err := fsm.getFilePathName("whatever")
	c.Assert(err, NotNil)
	fileName, err = fsm.getFilePathName("thorpub1addwnpepqf90u7n3nr2jwsw4t2gzhzqfdlply8dlzv3mdj4dr22uvhe04azq5gac3gq")
	c.Assert(err, IsNil)
	c.Assert(fileName, Equals, filepath.Join(f, "localstate-thorpub1addwnpepqf90u7n3nr2jwsw4t2gzhzqfdlply8dlzv3mdj4dr22uvhe04azq5gac3gq.json"))
}

func (s *FileStateMgrTestSuite) TestSaveLocalState(c *C) {
	stateItem := KeygenLocalState{
		PubKey:    "wasdfasdfasdfasdfasdfasdf",
		LocalData: keygen.NewLocalPartySaveData(5),
		ParticipantKeys: []string{
			"A", "B", "C",
		},
		LocalPartyKey: "A",
	}
	folder := os.TempDir()
	f := filepath.Join(folder, "test", "test1", "test2")
	defer func() {
		err := os.RemoveAll(f)
		c.Assert(err, IsNil)
	}()
	fsm, err := NewFileStateMgr(f)
	c.Assert(err, IsNil)
	c.Assert(fsm, NotNil)
	c.Assert(fsm.SaveLocalState(stateItem), NotNil)
	stateItem.PubKey = "thorpub1addwnpepqf90u7n3nr2jwsw4t2gzhzqfdlply8dlzv3mdj4dr22uvhe04azq5gac3gq"
	c.Assert(fsm.SaveLocalState(stateItem), IsNil)
	filePathName := filepath.Join(f, "localstate-"+stateItem.PubKey+".json")
	_, err = os.Stat(filePathName)
	c.Assert(err, IsNil)
	item, err := fsm.GetLocalState(stateItem.PubKey)
	c.Assert(err, IsNil)
	c.Assert(reflect.DeepEqual(stateItem, item), Equals, true)
}
