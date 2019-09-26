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
	"gamc.pro/gamcio/go-gamc/conf"
	"gamc.pro/gamcio/go-gamc/util/config"
	"gamc.pro/gamcio/go-gamc/util/logging"
	"github.com/btcsuite/btcutil/base58"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/big"
)

const (
	DefaultGenesisPath = "conf/genesis.yaml"
)

// Genesis Block Hash
var (
	GenesisHash      = []byte{10, 107, 126, 98, 237, 120, 159, 139, 240, 67, 134, 227, 127, 108, 206, 197, 236, 51, 176, 26, 218, 146, 194, 126, 194, 149, 216, 63, 18, 108, 55, 102}
	GenesisTimestamp = int64(1561615240)
)

type Token struct {
	Address string `yaml:"address"`
	Value   string `yaml:"value"`
}

type Genesis struct {
	ChainId                uint32   `yaml:"chain_id"`
	Token                  []Token  `yaml:"token"`
	Coinbase               string   `yaml:"coinbase"`
	StandbyNode            []string `yaml:"standby_node"`
	Foundation             Token    `yaml:"foundation"`
	FoundingTeam           Token    `yaml:"founding_team"`
	NodeDeployment         Token    `yaml:"node_deployment"`
	EcologicalConstruction Token    `yaml:"ecological_construction"`
	FoundingCommunity      Token    `yaml:"founding_community"`
}

func LoadGenesisConf(filePath string) (*Genesis, error) {
	in, err := ioutil.ReadFile(filePath)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to read the genesis config file.")
		return nil, err
	}
	genesis := new(Genesis)
	err = yaml.Unmarshal(in, genesis)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to parse genesis file.")
		return nil, err
	}
	return genesis, nil
}

func NewGenesis(cfg *config.Config, chain *BlockChain) (*Block, error) {
	if cfg == nil {
		return nil, ErrNilArgument
	}
	var genesis Block
	// load config
	chainConf := conf.GetChainConfig(cfg)
	genesisPath := chainConf.Genesis
	if len(genesisPath) == 0 {
		genesisPath = DefaultGenesisPath
	}
	genesisConf, err := LoadGenesisConf(genesisPath)
	if err != nil {
		panic("load genesis conf faild: " + err.Error())
	}
	witnesses := make([]*Witness, 0)
	for _, w := range chainConf.Witnesses {
		addr := base58.Decode(w)
		witnesses = append(witnesses, &Witness{master: &Address{addr}, follower: []*Address{{addr}}})
	}
	pd := PsecData{
		term:      0,
		timestamp: GenesisTimestamp,
	}
	worldState, err := NewWorldState(chain.db)
	coinbase, err := AddressParse(genesisConf.Coinbase)

	header := &BlockHeader{
		chainId:       genesisConf.ChainId,
		witnessreward: big.NewInt(0),
		coinbase:      coinbase,
		witnesses:     witnesses,
		psecData:      &pd,
		height:        0,
		timestamp:     GenesisTimestamp,
		sign:          nil,
		extra:         []byte("casc was born."),
	}
	genesis.worldState = worldState
	genesis.header = header
	genesis.db = chain.db

	if err := genesis.Begin(); err != nil {
		return nil, err
	}

	// add token for genesis
	for _, v := range genesisConf.Token {
		addr, err := AddressParse(v.Address)
		if err != nil {
			logging.CLog().WithFields(logrus.Fields{
				"address": v.Address,
				"err":     err,
			}).Error("Found invalid address in genesis token .")
			genesis.RollBack()
			return nil, err
		}
		acc, err := genesis.worldState.GetOrCreateAccount(addr.address)
		if err != nil {
			genesis.RollBack()
			return nil, err
		}
		txsBalance, status := new(big.Int).SetString(v.Value, 10)
		if !status {
			genesis.RollBack()
			return nil, ErrInvalidAmount
		}
		err = acc.AddBalance(txsBalance)
		if err != nil {
			genesis.RollBack()
			return nil, err
		}
	}
	if err := processingDistributionFund(&genesisConf.Foundation, &genesis); err != nil {
		genesis.RollBack()
		return nil, err
	}
	if err := processingDistributionFund(&genesisConf.FoundingTeam, &genesis); err != nil {
		genesis.RollBack()
		return nil, err
	}

	if err := processingDistributionFund(&genesisConf.NodeDeployment, &genesis); err != nil {
		genesis.RollBack()
		return nil, err
	}

	if err := processingDistributionFund(&genesisConf.EcologicalConstruction, &genesis); err != nil {
		genesis.RollBack()
		return nil, err
	}

	if err := processingDistributionFund(&genesisConf.FoundingCommunity, &genesis); err != nil {
		genesis.RollBack()
		return nil, err
	}

	genesis.Commit()

	genesis.header.hash = genesis.CalcHash()

	genesis.header.stateRoot = genesis.WorldState().AccountsRoot()
	genesis.header.txsRoot = genesis.WorldState().TxsRoot()
	return &genesis, nil
}

func processingDistributionFund(token *Token, gBlock *Block) error {
	addr, err := AddressParse(token.Address)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"address": token.Address,
			"err":     err,
		}).Error("Found invalid address in genesis .")
		return err
	}
	acc, err := gBlock.worldState.GetOrCreateAccount(addr.address)
	if err != nil {
		return err
	}
	txsBalance, status := new(big.Int).SetString(token.Value, 10)
	if !status {
		return ErrInvalidAmount
	}
	err = acc.AddBalance(txsBalance)
	if err != nil {
		return err
	}
	return nil
}
