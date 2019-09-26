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
package address

import (
	"gamc.pro/gamcio/go-gamc/conf"
	"gamc.pro/gamcio/go-gamc/core"
	"gamc.pro/gamcio/go-gamc/crypto"
	"gamc.pro/gamcio/go-gamc/crypto/cipher"
	"gamc.pro/gamcio/go-gamc/crypto/keystore"
	"gamc.pro/gamcio/go-gamc/util"
	"gamc.pro/gamcio/go-gamc/util/byteutils"
	"gamc.pro/gamcio/go-gamc/util/config"
	"gamc.pro/gamcio/go-gamc/util/logging"
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/tyler-smith/go-bip39"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	// ErrAddressNotFound address is not found.
	ErrAddressNotFound = errors.New("address is not found")
	// ErrAddressIsLocked address locked.
	ErrAddressIsLocked = errors.New("address is locked")
	ErrInvalidPrivateKey = errors.New("private key is invalid")
	ErrInvalidMnemonic = errors.New("mnemonic is invalid")
)

type addressInfo struct {
	// key address
	addr *core.Address
	// keystore save path
	path string
}

type AddressManager struct {
	// keystore
	ks *keystore.Keystore
	// key save path
	keydir string
	// address slice
	addresses []*addressInfo

	mu sync.Mutex
}

func NewAddressManager(config *config.Config) (*AddressManager, error) {
	am := new(AddressManager)
	am.ks = keystore.DefaultKS
	chaincfg := conf.GetChainConfig(config)

	tmpKeyDir, err := filepath.Abs(chaincfg.Keydir)
	if err != nil {
		return nil, err
	}
	am.keydir = tmpKeyDir

	if err := am.refreshAddresses(); err != nil {
		return nil, err
	}
	return am, err
}

// NewAccount returns a new address and keep it in keystore
func (am *AddressManager) NewAddress(passphrase []byte) (*core.Address, error) {
	priv, err := crypto.NewPrivateKey(nil)
	if err != nil {
		return nil, err
	}

	addr, err := am.setKeyStore(priv, passphrase)
	if err != nil {
		return nil, err
	}

	path, err := am.exportFile(addr, passphrase, false)
	if err != nil {
		return nil, err
	}

	am.updateAddressInfo(addr, path)

	return addr, nil
}

func (am *AddressManager) setKeyStore(priv keystore.PrivateKey, passphrase []byte) (*core.Address, error) {
	pub, err := priv.PublicKey().Encoded()
	if err != nil {
		return nil, err
	}
	addr, err := core.NewAddressFromPublicKey(pub)
	if err != nil {
		return nil, err
	}

	// set key to keystore
	err = am.ks.SetKey(addr.String(), priv, passphrase)
	if err != nil {
		return nil, err
	}

	return addr, nil
}

// Contains returns if contains address
func (am *AddressManager) Contains(addr *core.Address) bool {
	am.mu.Lock()
	defer am.mu.Unlock()

	for _, address := range am.addresses {
		if address.addr.Equals(addr) {
			return true
		}
	}
	return false
}

// Unlock unlock address with passphrase
func (am *AddressManager) Unlock(addr *core.Address, passphrase []byte, duration time.Duration) error {
	res, err := am.ks.ContainsAlias(addr.String())
	if err != nil || res == false {
		err = am.loadFile(addr, passphrase)
		if err != nil {
			return err
		}
	}
	return am.ks.Unlock(addr.String(), passphrase, duration)
}

// Lock lock address
func (am *AddressManager) Lock(addr *core.Address) error {
	return am.ks.Lock(addr.String())
}

// Accounts returns slice of address
func (am *AddressManager) Accounts() []*core.Address {
	am.refreshAddresses()

	am.mu.Lock()
	defer am.mu.Unlock()

	addrs := make([]*core.Address, len(am.addresses))
	for index, a := range am.addresses {
		addrs[index] = a.addr
	}
	return addrs
}

// loadFile import key to keystore in keydir
func (am *AddressManager) loadFile(addr *core.Address, passphrase []byte) error {
	address, err := am.getAddressInfo(addr)
	if err != nil {
		return err
	}

	raw, err := ioutil.ReadFile(address.path)
	if err != nil {
		return err
	}
	_, err = am.Load(raw, passphrase)
	return err
}

func (am *AddressManager) exportFile(addr *core.Address, passphrase []byte, overwrite bool) (path string, err error) {
	raw, err := am.Export(addr, passphrase)
	if err != nil {
		return "", err
	}

	acc, err := am.getAddressInfo(addr)
	// acc not found
	if err != nil {
		path = filepath.Join(am.keydir, addr.String())
	} else {
		path = acc.path
	}
	if err := util.FileWrite(path, raw, overwrite); err != nil {
		return "", err
	}
	return path, nil
}

func (am *AddressManager) ImportByPrivateKey(prikey, passphrase []byte) (*core.Address, error) {
	addr, err := am.LoadPrivate(prikey, passphrase)
	if err != nil {
		return nil, err
	}
	path, err := am.exportFile(addr, passphrase, false)
	if err != nil {
		return nil, err
	}

	am.updateAddressInfo(addr, path)

	return addr, nil
}

// Import import a key file to keystore, compatible ethereum keystore file, write to file
func (am *AddressManager) Import(keyjson, passphrase []byte) (*core.Address, error) {
	addr, err := am.Load(keyjson, passphrase)
	if err != nil {
		return nil, err
	}
	path, err := am.exportFile(addr, passphrase, false)
	if err != nil {
		return nil, err
	}

	am.updateAddressInfo(addr, path)

	return addr, nil
}

// Export export address to key file
func (am *AddressManager) Export(addr *core.Address, passphrase []byte) ([]byte, error) {
	key, err := am.ks.GetKey(addr.String(), passphrase)
	if err != nil {
		return nil, err
	}
	defer key.Clear()

	data, err := key.Encoded()
	if err != nil {
		return nil, err
	}
	defer ZeroBytes(data)

	cipher := cipher.NewCipher(uint8(keystore.SCRYPT))
	if err != nil {
		return nil, err
	}
	out, err := cipher.EncryptKey(addr.String(), data, passphrase)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Remove remove address and encrypted private key from keystore
func (am *AddressManager) RemoveAddress(addr *core.Address, passphrase []byte) error {
	err := am.ks.Delete(addr.String(), passphrase)
	if err != nil {
		return err
	}

	return nil
}

func (am *AddressManager) getAddressInfo(addr *core.Address) (*addressInfo, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	for _, address := range am.addresses {
		if address.addr.Equals(addr) {
			return address, nil
		}
	}
	return nil, ErrAddressNotFound
}

func (am *AddressManager) updateAddressInfo(addr *core.Address, path string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	var target *addressInfo
	for _, address := range am.addresses {
		if address.addr.Equals(addr) {
			target = address
			break
		}
	}
	if target != nil {
		target.path = path
	} else {
		target = &addressInfo{addr: addr, path: path}
		am.addresses = append(am.addresses, target)
	}
}

// Load load a key file to keystore, unable to write file
func (am *AddressManager) Load(keyjson, passphrase []byte) (*core.Address, error) {
	cipher := cipher.NewCipher(uint8(keystore.SCRYPT))
	data, err := cipher.DecryptKey(keyjson, passphrase)
	if err != nil {
		return nil, err
	}
	return am.LoadPrivate(data, passphrase)
}

// LoadPrivate load a private key to keystore, unable to write file
func (am *AddressManager) LoadPrivate(privatekey, passphrase []byte) (*core.Address, error) {
	defer ZeroBytes(privatekey)
	priv, err := crypto.NewPrivateKey(privatekey)
	if err != nil {
		return nil, err
	}
	defer priv.Clear()

	addr, err := am.setKeyStore(priv, passphrase)
	if err != nil {
		return nil, err
	}

	if _, err := am.getAddressInfo(addr); err != nil {
		am.mu.Lock()
		address := &addressInfo{addr: addr}
		am.addresses = append(am.addresses, address)
		am.mu.Unlock()
	}
	return addr, nil
}

// Update update addr locked passphrase
func (am *AddressManager) UpdatePassphrase(addr *core.Address, oldPassphrase, newPassphrase []byte) error {
	key, err := am.ks.GetKey(addr.String(), oldPassphrase)
	if err != nil {
		err = am.loadFile(addr, oldPassphrase)
		if err != nil {
			return err
		}
		key, err = am.ks.GetKey(addr.String(), oldPassphrase)
		if err != nil {
			return err
		}
	}
	defer key.Clear()

	if _, err := am.setKeyStore(key.(keystore.PrivateKey), newPassphrase); err != nil {
		return err
	}
	path, err := am.exportFile(addr, newPassphrase, true)
	if err != nil {
		return err
	}

	am.updateAddressInfo(addr, path)
	return nil
}

// SignHash sign hash
func (am *AddressManager) SignHash(addr *core.Address, hash byteutils.Hash) ([]byte, error) {
	key, err := am.ks.GetUnlocked(addr.String())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":  err,
			"addr": addr,
			"hash": hash,
		}).Error("Failed to get unlocked private key.")
		return nil, ErrAddressIsLocked
	}

	signature, err := crypto.NewSignature()
	if err != nil {
		return nil, err
	}

	if err := signature.InitSign(key.(keystore.PrivateKey)); err != nil {
		return nil, err
	}

	signData, err := signature.Sign(hash)
	if err != nil {
		return nil, err
	}
	return signData.GetData(), nil
}

// GetMnemonic
func (am *AddressManager) GetMnemonic(addr *core.Address, passphrase []byte) (string, error) {
	key, err := am.ks.GetKey(addr.String(), passphrase)
	if err != nil {
		return "", err
	}
	defer key.Clear()

	seed, err := key.(keystore.PrivateKey).Seed()
	if err != nil {
		return "", err
	}

	return bip39.NewMnemonic(seed)
}

func (am *AddressManager) GetPrivateKeyBytMnemonic(memo string) ([]byte, error) {
	if !bip39.IsMnemonicValid(memo) {
		return nil, ErrInvalidMnemonic
	}
	seed, err := bip39.EntropyFromMnemonic(memo)
	if err != nil {
		return nil, ErrInvalidMnemonic
	}
	privKey, err := crypto.NewPrivateKeyFromSeed(seed)
	if err != nil {
		return nil, err
	}
	return privKey.Encoded()
}

func (am *AddressManager) refreshAddresses() error {
	exist, err := util.FileExists(am.keydir)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Can't find the path")
		return err
	}

	if !exist {
		if err := os.MkdirAll(am.keydir, 0700); err != nil {
			panic("Failed to create keystore folder:" + am.keydir + ". err:" + err.Error())
		}
	}

	files, err := ioutil.ReadDir(am.keydir)
	if err != nil {
		return err
	}

	var (
		addresses []*addressInfo
	)

	for _, file := range files {
		acc, err := am.loadKeyFile(file)
		if err != nil {
			// errors have been recorded
			continue
		}
		addresses = append(addresses, acc)
	}
	am.addresses = addresses
	return nil
}

func (am *AddressManager) loadKeyFile(file os.FileInfo) (*addressInfo, error) {
	var (
		keyJSON struct {
			Address string `json:"address"`
		}
	)

	path := filepath.Join(am.keydir, file.Name())

	if file.IsDir() || strings.HasPrefix(file.Name(), ".") || strings.HasSuffix(file.Name(), "~") {
		logging.VLog().WithFields(logrus.Fields{
			"path": path,
		}).Warn("Skipped this key file.")
		return nil, errors.New("File need skip")
	}

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to read the key file")
		return nil, errors.New("Failed to read the key file")
	}

	keyJSON.Address = ""
	err = json.Unmarshal(raw, &keyJSON)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to parse the key file")
		return nil, errors.New("Failed to parse the key file")
	}

	addr, err := core.AddressParse(keyJSON.Address)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":     err,
			"address": keyJSON.Address,
		}).Error("Failed to parse the address.")
		return nil, errors.New("failed to parse the address")
	}
	acc := &addressInfo{addr, path}
	return acc, nil
}

func ZeroBytes(bytes []byte) {
	for i := range bytes {
		bytes[i] = 0
	}
}

// SignBlock sign block with the specified algorithm
func (am *AddressManager) SignBlock(addr *core.Address, block *core.Block) error {
	key, err := am.ks.GetUnlocked(addr.String())
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":   err,
			"block": block,
		}).Error("Failed to get unlocked private key to sign block.")
		return ErrAddressIsLocked
	}

	signature, err := crypto.NewSignature()
	if err != nil {
		return err
	}
	signature.InitSign(key.(keystore.PrivateKey))
	return block.Sign(signature)
}
