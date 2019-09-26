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
	"gamc.pro/gamcio/go-gamc/network"
	"gamc.pro/gamcio/go-gamc/storage/cdb"
	"gamc.pro/gamcio/go-gamc/util/config"
	"errors"
	"math/big"
)

var (
	ErrUnknownAncestor = errors.New("unknown ancestor")
	ErrFutureBlock     = errors.New("block in the future")
	ErrInvalidNumber   = errors.New("invalid block number")
)

// ConsensusEngine
type Consensus interface {
	Setup(gamc gamc)
	Start()
	Stop()

	EnableMining()  //
	DisableMining() //
	IsEnable() bool

	ResumeMining()
	SuspendMining()  //
	IsSuspend() bool //

	UpdateFixedBlock()
	VerifyBlock(block *Block) error
	AccumulateRewards(addr Address, reward *big.Int) error
}

// Synchronize interface of sync service
type Synchronize interface {
	Start()
	Stop()

	StartActiveSync() bool
	StopActiveSync()
	WaitingForFinish()
	IsActiveSyncing() bool
}

type AccountManager interface {
	NewAccount(passphrase []byte) (*Address, string, error)
	UpdateAccount(address string, oldPassphrase, newPassphrase []byte) error
	Sign(address *Address, hash []byte) ([]byte, error)
	SignBlock(address *Address, block *Block) error
	Verify(pubKey []byte, message, sig []byte) bool
}

type gamc interface {
	BlockChain() *BlockChain
	NetService() network.Service
	AccountManager() AccountManager
	Consensus() Consensus
	Config() *config.Config
	Storage() cdb.Storage
}
