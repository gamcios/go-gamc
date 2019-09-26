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
package account

import (
	"gamc.pro/gamcio/go-gamc/core"
	"gamc.pro/gamcio/go-gamc/core/address"
	"gamc.pro/gamcio/go-gamc/storage/cdb"
	"gamc.pro/gamcio/go-gamc/util/config"
	"gamc.pro/gamcio/go-gamc/util/logging"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

const (
	DefaultAddressUnlockedDuratuion = time.Second * 60
)

var (
	ErrSaveDBError = errors.New("Failed to save datebase")
	ErrInitDBError = errors.New("Failure of database initialization")
)

type AccountManager struct {
	addrManger *address.AddressManager
	db         cdb.Storage
	mutex      sync.Mutex
}

func NewAccountManager(config *config.Config, db cdb.Storage) (*AccountManager, error) {
	if config == nil || db == nil {
		logging.CLog().WithFields(logrus.Fields{
			"err": core.ErrInvalidArgument,
		}).Error("Failed to init AccountManager")
		return nil, core.ErrInvalidArgument
	}
	accMgr := new(AccountManager)
	var err error
	accMgr.addrManger, err = address.NewAddressManager(config)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to create address manager")
		return nil, err
	}

	accMgr.db = db

	if accMgr.db == nil {
		logging.CLog().WithFields(logrus.Fields{}).Error("Failed to init db")
		return nil, ErrInitDBError
	}

	return accMgr, nil
}

func (am *AccountManager) AddressManager() *address.AddressManager {
	return am.addrManger
}

//return address,mnemonicWord
func (am *AccountManager) NewAccount(passphrase []byte) (*core.Address, string, error) {
	add, err := am.addrManger.NewAddress(passphrase)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to new account")
		return nil, "", err
	}
	memo, err := am.addrManger.GetMnemonic(add, passphrase)
	if err != nil {
		return nil, "", err
	}
	return add, memo, nil
}

func (am *AccountManager) AddressIsValid(address string) (*core.Address, error) {
	addr, err := core.AddressParse(address)
	if err != nil {
		return nil, err
	}
	return addr, err
}

func (am *AccountManager) UpdateAccount(address string, oldPassphrase, newPassphrase []byte) error {
	addr, err := am.AddressIsValid(address)
	if err != nil {
		return err
	}
	return am.addrManger.UpdatePassphrase(addr, oldPassphrase, newPassphrase)
}

func (am *AccountManager) ImportAccount(priKey, passphrase []byte) (*core.Address, error) {
	addr, err := am.addrManger.ImportByPrivateKey(priKey, passphrase)
	if err != nil {
		return nil, err
	}
	return addr, err
}

func (am *AccountManager) GetAllAddress() []*core.Address {
	return am.addrManger.Accounts()
}

func (am *AccountManager) Sign(address *core.Address, hash []byte) ([]byte, error) {
	return am.addrManger.SignHash(address, hash)
}

func (am *AccountManager) SignBlock(address *core.Address, block *core.Block) error {
	return am.addrManger.SignBlock(address, block)
}

func (am *AccountManager) Verify(pubKey []byte, message, sig []byte) bool {
	return am.Verify(pubKey, message, sig)
}

func (am *AccountManager) UnLock(address *core.Address, passphrase []byte, duratuion time.Duration) error {
	if duratuion == 0 {
		duratuion = DefaultAddressUnlockedDuratuion
	}
	return am.addrManger.Unlock(address, passphrase, duratuion)
}

func (am *AccountManager) Lock(address *core.Address) error {
	return am.addrManger.Lock(address)
}
