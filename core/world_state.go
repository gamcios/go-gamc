// Copyright (C) 2018 go-gamc authors
//
// This file is part of the go-gamc library.
//
// the go-gamc library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-gamc library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-gamc library.  If not, see <http://www.gnu.org/licenses/>.
//

package core

import (
	"gamc.pro/gamcio/go-gamc/storage/cdb"
	"gamc.pro/gamcio/go-gamc/storage/mvccdb"
	"gamc.pro/gamcio/go-gamc/trie"
	"gamc.pro/gamcio/go-gamc/util/byteutils"
)

// WorldState interface of world state
type WorldState interface {
	Begin() error
	Commit() error
	RollBack() error

	Prepare(interface{}) (TxWorldState, error)
	Reset(addr byteutils.Hash, isResetChangeLog bool) error
	Flush() error
	Abort() error

	NextConsensusState(int64) (ConsensusState, error)
	SetConsensusState(ConsensusState)

	LoadAccountsRoot(byteutils.Hash) error
	LoadTxsRoot(byteutils.Hash) error

	Clone() (WorldState, error)

	AccountsRoot() byteutils.Hash
	TxsRoot() byteutils.Hash

	Accounts() ([]Account, error)
	GetOrCreateAccount(addr byteutils.Hash) (Account, error)

	GetTx(txHash byteutils.Hash) ([]byte, error)
	PutTx(txHash byteutils.Hash, txBytes []byte) error

	GetBlockHashByHeight(height uint64) ([]byte, error)
	GetBlock(txHash byteutils.Hash) ([]byte, error)
}

// TxWorldState is the world state of a single transaction
type TxWorldState interface {
	AccountsRoot() byteutils.Hash
	TxsRoot() byteutils.Hash

	CheckAndUpdate() ([]interface{}, error)
	Reset(addr byteutils.Hash, isResetChangeLog bool) error
	Close() error

	Accounts() ([]Account, error)
	GetOrCreateAccount(addr byteutils.Hash) (Account, error)

	GetTx(txHash byteutils.Hash) ([]byte, error)
	PutTx(txHash byteutils.Hash, txBytes []byte) error

	GetBlockHashByHeight(height uint64) ([]byte, error)
	GetBlock(txHash byteutils.Hash) ([]byte, error)
}

func newStates(stor cdb.Storage) (*states, error) {
	changelog, err := newChangeLog()
	if err != nil {
		return nil, err
	}
	stateDB, err := newStateDB(stor)
	if err != nil {
		return nil, err
	}

	accState, err := NewAccountState(nil, stateDB)
	if err != nil {
		return nil, err
	}

	txsState, err := trie.NewTrie(nil, stateDB, false)
	if err != nil {
		return nil, err
	}

	return &states{
		accState:  accState,
		txsState:  txsState,
		changelog: changelog,
		stateDB:   stateDB,
		innerDB:   stor,
		txid:      nil,
	}, nil
}

func newChangeLog() (*mvccdb.MVCCDB, error) {
	mem, err := cdb.NewMemoryStorage()
	if err != nil {
		return nil, err
	}
	db, err := mvccdb.NewMVCCDB(mem, false)
	if err != nil {
		return nil, err
	}

	db.SetStrictGlobalVersionCheck(true)
	return db, nil
}

func newStateDB(storage cdb.Storage) (*mvccdb.MVCCDB, error) {
	return mvccdb.NewMVCCDB(storage, true)
}

type states struct {
	accState       AccountState
	txsState       *trie.Trie
	consensusState ConsensusState
	changelog      *mvccdb.MVCCDB
	stateDB        *mvccdb.MVCCDB
	innerDB        cdb.Storage
	txid           interface{}
}

func (s *states) Replay(done *states) error {
	err := s.accState.Replay(done.accState)
	if err != nil {
		return err
	}
	_, err = s.txsState.Replay(done.txsState)
	if err != nil {
		return err
	}
	return nil
}

func (s *states) Clone() (*states, error) {
	changelog, err := newChangeLog()
	if err != nil {
		return nil, err
	}
	stateDB, err := newStateDB(s.innerDB)
	if err != nil {
		return nil, err
	}

	accState, err := NewAccountState(s.accState.RootHash(), stateDB)
	if err != nil {
		return nil, err
	}

	txsState, err := trie.NewTrie(s.txsState.RootHash(), stateDB, false)
	if err != nil {
		return nil, err
	}

	return &states{
		accState: accState,
		txsState: txsState,

		changelog: changelog,
		stateDB:   stateDB,
		innerDB:   s.innerDB,
		txid:      s.txid,
	}, nil
}

func (s *states) Begin() error {
	if err := s.changelog.Begin(); err != nil {
		return err
	}
	if err := s.stateDB.Begin(); err != nil {
		return err
	}
	return nil
}

func (s *states) Commit() error {
	if err := s.Flush(); err != nil {
		return err
	}
	// changelog is used to check conflict temporarily
	// we should rollback it when the transaction is over
	if err := s.changelog.RollBack(); err != nil {
		return err
	}
	if err := s.stateDB.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *states) RollBack() error {
	if err := s.Abort(); err != nil {
		return err
	}
	if err := s.changelog.RollBack(); err != nil {
		return err
	}
	if err := s.stateDB.RollBack(); err != nil {
		return err
	}

	return nil
}

func (s *states) Prepare(txid interface{}) (*states, error) {
	changelog, err := s.changelog.Prepare(txid)
	if err != nil {
		return nil, err
	}
	stateDB, err := s.stateDB.Prepare(txid)
	if err != nil {
		return nil, err
	}

	// Flush all changes in world state into merkle trie
	// make a snapshot of world state
	if err := s.Flush(); err != nil {
		return nil, err
	}

	accState, err := NewAccountState(s.AccountsRoot(), stateDB)
	if err != nil {
		return nil, err
	}

	txsState, err := trie.NewTrie(s.TxsRoot(), stateDB, true)
	if err != nil {
		return nil, err
	}

	return &states{
		accState: accState,
		txsState: txsState,

		changelog: changelog,
		stateDB:   stateDB,
		innerDB:   s.innerDB,
		txid:      txid,
	}, nil
}

func (s *states) CheckAndUpdateTo(parent *states) ([]interface{}, error) {
	dependency, err := s.changelog.CheckAndUpdate()
	if err != nil {
		return nil, err
	}
	_, err = s.stateDB.CheckAndUpdate()
	if err != nil {
		return nil, err
	}
	if err := parent.Replay(s); err != nil {
		return nil, err
	}
	return dependency, nil
}

func (s *states) Reset(addr byteutils.Hash, isResetChangeLog bool) error {

	if err := s.stateDB.Reset(); err != nil {
		return err
	}

	if err := s.Abort(); err != nil {
		return err
	}

	if isResetChangeLog {
		if err := s.changelog.Reset(); err != nil {
			return err
		}
		if addr != nil {
			// record dependency
			if err := s.changelog.Put(addr, addr); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *states) Flush() error {
	return s.accState.Flush()
}

func (s *states) Abort() error {
	// TODO: Abort txsState, eventsState, consensusState
	// we don't need to abort the three states now
	// because we only use abort in reset, close and rollback
	// in close & rollback, we won't use states any more
	// in reset, we won't change the three states before we reset them
	return s.accState.Abort()
}

func (s *states) Close() error {
	if err := s.changelog.Close(); err != nil {
		return err
	}
	if err := s.stateDB.Close(); err != nil {
		return err
	}
	if err := s.Abort(); err != nil {
		return err
	}

	return nil
}

func (s *states) AccountsRoot() byteutils.Hash {
	return s.accState.RootHash()
}

func (s *states) TxsRoot() byteutils.Hash {
	return s.txsState.RootHash()
}

func (s *states) Accounts() ([]Account, error) { // TODO delete
	return s.accState.Accounts()
}

func (s *states) GetOrCreateAccount(addr byteutils.Hash) (Account, error) {
	acc, err := s.accState.GetOrCreateAccount(addr)
	if err != nil {
		return nil, err
	}
	return s.recordAccount(acc)
}

func (s *states) recordAccount(acc Account) (Account, error) {
	if err := s.changelog.Put(acc.Address(), acc.Address()); err != nil {
		return nil, err
	}
	return acc, nil
}

// WorldState manange all current states in Blockchain
type worldState struct {
	*states
	snapshot *states
}

func (s *states) GetTx(txHash byteutils.Hash) ([]byte, error) {
	bytes, err := s.txsState.Get(txHash)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (s *states) PutTx(txHash byteutils.Hash, txBytes []byte) error {
	_, err := s.txsState.Put(txHash, txBytes)
	if err != nil {
		return err
	}
	return nil
}

func (s *states) LoadAccountsRoot(root byteutils.Hash) error {
	accState, err := NewAccountState(root, s.stateDB)
	if err != nil {
		return err
	}
	s.accState = accState
	return nil
}

func (s *states) LoadTxsRoot(root byteutils.Hash) error {
	txsState, err := trie.NewTrie(root, s.stateDB, false)
	if err != nil {
		return err
	}
	s.txsState = txsState
	return nil
}

func (s *states) GetBlockHashByHeight(height uint64) ([]byte, error) {
	bytes, err := s.innerDB.Get(byteutils.FromUint64(height))
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (s *states) GetBlock(hash byteutils.Hash) ([]byte, error) {
	bytes, err := s.innerDB.Get(hash)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// NewWorldState create a new empty WorldState
func NewWorldState(storage cdb.Storage) (WorldState, error) {
	states, err := newStates(storage)
	if err != nil {
		return nil, err
	}
	return &worldState{
		states:   states,
		snapshot: nil,
	}, nil
}

// Clone a new WorldState
func (ws *worldState) Clone() (WorldState, error) {
	s, err := ws.states.Clone()
	if err != nil {
		return nil, err
	}
	return &worldState{
		states:   s,
		snapshot: nil,
	}, nil
}

func (ws *worldState) NextConsensusState(elapsedSecond int64) (ConsensusState, error) {
	return ws.states.consensusState.NextConsensusState(elapsedSecond, ws)
}

func (ws *worldState) SetConsensusState(consensusState ConsensusState) {
	ws.states.consensusState = consensusState
}

func (ws *worldState) Begin() error {
	snapshot, err := ws.states.Clone()
	if err != nil {
		return err
	}
	if err := ws.states.Begin(); err != nil {
		return err
	}
	ws.snapshot = snapshot
	return nil
}

func (ws *worldState) Commit() error {
	if err := ws.states.Commit(); err != nil {
		return err
	}
	ws.snapshot = nil
	return nil
}

func (ws *worldState) RollBack() error {
	if err := ws.states.RollBack(); err != nil {
		return err
	}
	ws.states = ws.snapshot
	ws.snapshot = nil
	return nil
}

func (ws *worldState) Prepare(txid interface{}) (TxWorldState, error) {
	s, err := ws.states.Prepare(txid)
	if err != nil {
		return nil, err
	}
	txState := &txWorldState{
		states: s,
		txid:   txid,
		parent: ws,
	}
	return txState, nil
}

type txWorldState struct {
	*states
	txid   interface{}
	parent *worldState
}

func (tws *txWorldState) CheckAndUpdate() ([]interface{}, error) {
	dependencies, err := tws.states.CheckAndUpdateTo(tws.parent.states)
	if err != nil {
		return nil, err
	}
	tws.parent = nil
	return dependencies, nil
}

func (tws *txWorldState) Reset(addr byteutils.Hash, isResetChangeLog bool) error {
	if err := tws.states.Reset(addr, isResetChangeLog); err != nil {
		return err
	}
	return nil
}

func (tws *txWorldState) Close() error {
	if err := tws.states.Close(); err != nil {
		return err
	}
	tws.parent = nil
	return nil
}

func (tws *txWorldState) TxID() interface{} {
	return tws.txid
}
