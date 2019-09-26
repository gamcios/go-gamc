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
	corepb "gamc.pro/gamcio/go-gamc/core/pb"
	"gamc.pro/gamcio/go-gamc/storage/cdb"
	"gamc.pro/gamcio/go-gamc/trie"
	"gamc.pro/gamcio/go-gamc/util/byteutils"
	"errors"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"math/big"
)

// Errors
var (
	ErrBalanceInsufficient     = errors.New("cannot subtract a value which is bigger than current balance")
	ErrFrozenFundInsufficient  = errors.New("cannot subtract a value which is bigger than frozen fund")
	ErrPledgeFundInsufficient  = errors.New("cannot subtract a value which is bigger than pledge fund")
	ErrAccountNotFound         = errors.New("cannot found account in storage")
	ErrContractAccountNotFound = errors.New("cannot found contract account in storage please check contract address is valid or deploy is success")
)

// Iterator Variables in Account Storage
type Iterator interface {
	Next() (bool, error)
	Value() []byte
}

// Account Interface
type Account interface {
	Address() byteutils.Hash
	Balance() *big.Int
	FrozenFund() *big.Int
	PledgeFund() *big.Int
	Nonce() uint64
	CreditIndex() *big.Int
	VarsHash() byteutils.Hash
	Clone() (Account, error)

	ToBytes() ([]byte, error)
	FromBytes(bytes []byte, storage cdb.Storage) error

	IncrNonce()
	AddBalance(value *big.Int) error
	SubBalance(value *big.Int) error
	AddFrozenFund(value *big.Int) error
	SubFrozenFund(value *big.Int) error
	AddPledgeFund(value *big.Int) error
	SubPledgeFund(value *big.Int) error
	AddCreditIndex(value *big.Int) error
	SubCreditIndex(value *big.Int) error
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	Del(key []byte) error
	Iterator(prefix []byte) (Iterator, error)
}

// AccountState Interface
type AccountState interface {
	RootHash() byteutils.Hash

	Flush() error
	Abort() error

	DirtyAccounts() ([]Account, error)
	Accounts() ([]Account, error)

	Clone() (AccountState, error)
	Replay(AccountState) error

	GetOrCreateAccount(byteutils.Hash) (Account, error)
}

// account info in state Trie
type account struct {
	address     byteutils.Hash
	balance     *big.Int
	frozenFund  *big.Int
	pledgeFund  *big.Int
	nonce       uint64
	variables   *trie.Trie
	creditIndex *big.Int
	permissions []*corepb.Permission
}

// ToBytes converts domain Account to bytes
func (acc *account) ToBytes() ([]byte, error) {
	pbAcc := &corepb.Account{
		Address:    acc.address,
		Balance:    acc.balance.Bytes(),
		FrozenFund: acc.frozenFund.Bytes(),
		PledgeFund: acc.frozenFund.Bytes(),
		Nonce:      acc.nonce,

		VarsHash:    acc.variables.RootHash(),
		CreditIndex: acc.creditIndex.Bytes(),
		Permissions: acc.permissions,
	}
	bytes, err := proto.Marshal(pbAcc)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// FromBytes converts bytes to Account
func (acc *account) FromBytes(bytes []byte, storage cdb.Storage) error {
	pbAcc := &corepb.Account{}
	var err error
	if err = proto.Unmarshal(bytes, pbAcc); err != nil {
		return err
	}
	acc.address = pbAcc.Address
	acc.balance = new(big.Int).SetBytes(pbAcc.Balance)
	acc.frozenFund = new(big.Int).SetBytes(pbAcc.FrozenFund)
	acc.pledgeFund = new(big.Int).SetBytes(pbAcc.PledgeFund)
	acc.nonce = pbAcc.Nonce

	acc.variables, err = trie.NewTrie(pbAcc.VarsHash, storage, false)
	if err != nil {
		return err
	}
	acc.creditIndex = new(big.Int).SetBytes(pbAcc.CreditIndex)
	return nil
}

// Address return account's address
func (acc *account) Address() byteutils.Hash {
	return acc.address
}

// Balance return account's balance
func (acc *account) Balance() *big.Int {
	return acc.balance
}

// FrozenFund return account's frozen fund
func (acc *account) FrozenFund() *big.Int {
	return acc.frozenFund
}

// PledgeFund return account's pledge fund
func (acc *account) PledgeFund() *big.Int {
	return acc.frozenFund
}

// Nonce return account's nonce
func (acc *account) Nonce() uint64 {
	return acc.nonce
}

// CreditIndex return account's credit index
func (acc *account) CreditIndex() *big.Int {
	return acc.creditIndex
}

// VarsHash return account's variables hash
func (acc *account) VarsHash() byteutils.Hash {
	return acc.variables.RootHash()
}

// Clone account
func (acc *account) Clone() (Account, error) {
	variables, err := acc.variables.Clone()
	if err != nil {
		return nil, err
	}

	return &account{
		address:     acc.address,
		balance:     acc.balance,
		frozenFund:  acc.frozenFund,
		pledgeFund:  acc.pledgeFund,
		creditIndex: acc.creditIndex,
		nonce:       acc.nonce,
		variables:   variables,
		permissions: acc.permissions,
	}, nil
}

// IncrNonce by 1
func (acc *account) IncrNonce() {
	acc.nonce++
}

// AccountState manage account state in Block
type accountState struct {
	stateTrie    *trie.Trie
	dirtyAccount map[byteutils.HexHash]Account
	storage      cdb.Storage
}

// AddBalance to an account
func (acc *account) AddBalance(value *big.Int) error {
	balance := new(big.Int).Add(acc.balance, value)
	acc.balance = balance
	return nil
}

// SubBalance to an account
func (acc *account) SubBalance(value *big.Int) error {
	if acc.balance.Cmp(value) < 0 {
		return ErrBalanceInsufficient
	}
	balance := new(big.Int).Sub(acc.balance, value)
	acc.balance = balance
	return nil
}

// AddFrozenFund to an account
func (acc *account) AddFrozenFund(value *big.Int) error {
	frozenFund := new(big.Int).Add(acc.frozenFund, value)
	acc.frozenFund = frozenFund
	return nil
}

// SubFrozenFund to an account
func (acc *account) SubFrozenFund(value *big.Int) error {
	if acc.frozenFund.Cmp(value) < 0 {
		return ErrFrozenFundInsufficient
	}
	frozenFund := new(big.Int).Sub(acc.frozenFund, value)
	acc.frozenFund = frozenFund
	return nil
}

// AddPledgeFund to an account
func (acc *account) AddPledgeFund(value *big.Int) error {
	pledgeFund := new(big.Int).Add(acc.pledgeFund, value)
	acc.pledgeFund = pledgeFund
	return nil
}

// SubPledgeFund to an account
func (acc *account) SubPledgeFund(value *big.Int) error {
	if acc.pledgeFund.Cmp(value) < 0 {
		return ErrPledgeFundInsufficient
	}
	pledgeFund := new(big.Int).Sub(acc.pledgeFund, value)
	acc.pledgeFund = pledgeFund
	return nil
}

// AddCreditIndex to an account
func (acc *account) AddCreditIndex(value *big.Int) error {
	acc.creditIndex = new(big.Int).Add(acc.creditIndex, value)
	return nil
}

// SubCreditIndex to an account
func (acc *account) SubCreditIndex(value *big.Int) error {
	acc.creditIndex = new(big.Int).Sub(acc.creditIndex, value)
	return nil
}

// Put into account's storage
func (acc *account) Put(key []byte, value []byte) error {
	_, err := acc.variables.Put(key, value)
	return err
}

// Get from account's storage
func (acc *account) Get(key []byte) ([]byte, error) {
	return acc.variables.Get(key)
}

// Del from account's storage
func (acc *account) Del(key []byte) error {
	if _, err := acc.variables.Del(key); err != nil {
		return err
	}
	return nil
}

// Iterator map var from account's storage
func (acc *account) Iterator(prefix []byte) (Iterator, error) {
	return acc.variables.Iterator(prefix)
}

func (acc *account) String() string {
	return fmt.Sprintf("Account %p {Address: %v, Balance:%v, FrozenFund:%v, PledgeFund:%v, CreditIndex:%v; Nonce:%v; VarsHash:%v;}",
		acc,
		byteutils.Hex(acc.address),
		acc.balance.String(),
		acc.frozenFund.String(),
		acc.pledgeFund.String(),
		acc.creditIndex,
		acc.nonce,
		byteutils.Hex(acc.variables.RootHash()),
	)
}

// NewAccountState create a new account state
func NewAccountState(root byteutils.Hash, storage cdb.Storage) (AccountState, error) {
	stateTrie, err := trie.NewTrie(root, storage, false)
	if err != nil {
		return nil, err
	}

	return &accountState{
		stateTrie:    stateTrie,
		dirtyAccount: make(map[byteutils.HexHash]Account),
		storage:      storage,
	}, nil
}

func (as *accountState) Flush() error {
	for addr, acc := range as.dirtyAccount {
		bytes, err := acc.ToBytes()
		if err != nil {
			return err
		}
		key, err := addr.Hash()
		if err != nil {
			return err
		}
		as.stateTrie.Put(key, bytes)
	}
	as.dirtyAccount = make(map[byteutils.HexHash]Account)
	return nil
}

func (as *accountState) Abort() error {
	as.dirtyAccount = make(map[byteutils.HexHash]Account)
	return nil
}

// RootHash return root hash of account state
func (as *accountState) RootHash() byteutils.Hash {
	return as.stateTrie.RootHash()
}

func (as *accountState) Accounts() ([]Account, error) { // TODO delete
	accounts := []Account{}
	iter, err := as.stateTrie.Iterator(nil)
	if err != nil && err != cdb.ErrKeyNotFound {
		return nil, err
	}
	if err != nil {
		return accounts, nil
	}
	exist, err := iter.Next()
	if err != nil {
		return nil, err
	}
	for exist {
		acc := new(account)
		err = acc.FromBytes(iter.Value(), as.storage)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
		exist, err = iter.Next()
		if err != nil {
			return nil, err
		}
	}
	return accounts, nil
}

// DirtyAccounts return all changed accounts
func (as *accountState) DirtyAccounts() ([]Account, error) {
	accounts := []Account{}
	for _, account := range as.dirtyAccount {
		accounts = append(accounts, account)
	}
	return accounts, nil
}

// Relay merge the done account state
func (as *accountState) Replay(done AccountState) error {
	state := done.(*accountState)
	for addr, acc := range state.dirtyAccount {
		as.dirtyAccount[addr] = acc
	}
	return nil
}

// Clone an accountState
func (as *accountState) Clone() (AccountState, error) {
	stateTrie, err := as.stateTrie.Clone()
	if err != nil {
		return nil, err
	}

	dirtyAccount := make(map[byteutils.HexHash]Account)
	for addr, acc := range as.dirtyAccount {
		dirtyAccount[addr], err = acc.Clone()
		if err != nil {
			return nil, err
		}
	}

	return &accountState{
		stateTrie:    stateTrie,
		dirtyAccount: dirtyAccount,
		storage:      as.storage,
	}, nil
}

// GetOrCreateAccount according to the addr
func (as *accountState) GetOrCreateAccount(addr byteutils.Hash) (Account, error) {
	acc, err := as.getAccount(addr)
	if err != nil && err != ErrAccountNotFound {
		return nil, err
	}
	if err == ErrAccountNotFound {
		acc, err = as.newAccount(addr)
		if err != nil {
			return nil, err
		}
		return acc, nil
	}
	return acc, nil
}

func (as *accountState) newAccount(addr byteutils.Hash) (Account, error) {
	varTrie, err := trie.NewTrie(nil, as.storage, false)
	if err != nil {
		return nil, err
	}
	acc := &account{
		address:     addr,
		balance:     big.NewInt(0),
		frozenFund:  big.NewInt(0),
		pledgeFund:  big.NewInt(0),
		nonce:       0,
		variables:   varTrie,
		creditIndex: big.NewInt(0),
	}
	as.recordDirtyAccount(addr, acc)
	return acc, nil
}

func (as *accountState) recordDirtyAccount(addr byteutils.Hash, acc Account) {
	as.dirtyAccount[addr.Hex()] = acc
}

func (as *accountState) getAccount(addr byteutils.Hash) (Account, error) {
	// search in dirty account
	if acc, ok := as.dirtyAccount[addr.Hex()]; ok {
		return acc, nil
	}
	// search in storage
	bytes, err := as.stateTrie.Get(addr)
	if err != nil && err != cdb.ErrKeyNotFound {
		return nil, err
	}
	if err == nil {
		acc := new(account)
		err = acc.FromBytes(bytes, as.storage)
		if err != nil {
			return nil, err
		}
		as.recordDirtyAccount(addr, acc)
		return acc, nil
	}
	return nil, ErrAccountNotFound
}

func (as *accountState) String() string {
	return fmt.Sprintf("AccountState %p {RootHash:%s; dirtyAccount:%v; Storage:%p}",
		as,
		byteutils.Hex(as.stateTrie.RootHash()),
		as.dirtyAccount,
		as.storage,
	)
}
