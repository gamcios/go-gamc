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
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-gamc library.  If not, see <http://www.gnu.org/licenses/>.
//
package core

import (
	"gamc.pro/gamcio/go-gamc/conf"
	corepb "gamc.pro/gamcio/go-gamc/core/pb"
	"gamc.pro/gamcio/go-gamc/network"
	"gamc.pro/gamcio/go-gamc/storage/cdb"
	"gamc.pro/gamcio/go-gamc/util/byteutils"
	"gamc.pro/gamcio/go-gamc/util/config"
	"gamc.pro/gamcio/go-gamc/util/logging"
	"github.com/gogo/protobuf/proto"
	lru "github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	// ChunkSize is the size of blocks in a chunk
	ChunkSize = 32
	// Tail Key in storage
	Tail = "blockchain_tail"
	// Fixed in storage
	FIXED = "blockchain_fixed"
)

// BlockChain
type BlockChain struct {
	chainId            uint32
	config             *config.Config
	consensus          Consensus
	sync               Synchronize
	db                 cdb.Storage
	currentHeader      *BlockHeader
	currentBlock       *Block
	txPool             *TxPool
	bkPool             *BlockPool
	genesisBlock       *Block
	tailBlock          *Block
	fixedBlock         *Block
	cachedBlocks       *lru.Cache
	detachedTailBlocks *lru.Cache
	quitCh             chan int
}

// NewBlockChain
func NewBlockChain(config *config.Config, net network.Service, db cdb.Storage) (*BlockChain, error) {

	blockPool, err := NewBlockPool(128)
	if err != nil {
		return nil, err
	}

	chaincfg := conf.GetChainConfig(config)
	txPool := NewTxPool()

	chain := &BlockChain{
		chainId: chaincfg.ChainId,
		config:  config,
		db:      db,
		bkPool:  blockPool,
		txPool:  txPool,
	}

	blockPool.RegisterInNetwork(net)

	chain.cachedBlocks, err = lru.New(128)
	if err != nil {
		return nil, err
	}

	chain.detachedTailBlocks, err = lru.New(128)
	if err != nil {
		return nil, err
	}

	chain.bkPool.setBlockChain(chain)

	return chain, nil
}

func (bc *BlockChain) Setup(gamc gamc) error {
	bc.consensus = gamc.Consensus()

	var err error
	bc.genesisBlock, err = bc.LoadGenesisFromStorage()
	if err != nil {
		return err
	}

	bc.tailBlock, err = bc.LoadTailFromStorage()
	if err != nil {
		return err
	}
	logging.CLog().WithFields(logrus.Fields{
		"tail": bc.tailBlock,
	}).Info("Tail Block.")
	bc.fixedBlock, err = bc.LoadFixedFromStorage()
	if err != nil {
		return err
	}
	logging.CLog().WithFields(logrus.Fields{
		"block": bc.fixedBlock,
	}).Info("Latest Permanent Block.")
	return nil
}

// LoadGenesisFromStorage load genesis
func (bc *BlockChain) LoadGenesisFromStorage() (*Block, error) { // ToRefine, remove or ?
	genesis, err := LoadBlockFromStorage(GenesisHash, bc)
	if err != nil && err != cdb.ErrKeyNotFound {
		return nil, err
	}
	if err == cdb.ErrKeyNotFound {
		genesis, err = NewGenesis(bc.config, bc)
		if err != nil {
			return nil, err
		}
		if err := bc.StoreBlockToStorage(genesis); err != nil {
			return nil, err
		}
		heightKey := byteutils.FromUint64(genesis.Height())
		if err := bc.db.Put(heightKey, genesis.Hash()); err != nil {
			return nil, err
		}
	}
	return genesis, nil
}

// LoadTailFromStorage load tail block
func (bc *BlockChain) LoadTailFromStorage() (*Block, error) {
	hash, err := bc.db.Get([]byte(Tail))
	if err != nil && err != cdb.ErrKeyNotFound {
		return nil, err
	}
	if err == cdb.ErrKeyNotFound {
		genesis, err := bc.LoadGenesisFromStorage()
		if err != nil {
			return nil, err
		}

		if err := bc.StoreTailHashToStorage(genesis); err != nil {
			return nil, err
		}

		return genesis, nil
	}

	return LoadBlockFromStorage(hash, bc)
}

func (bc *BlockChain) StoreTailHashToStorage(block *Block) error {
	return bc.db.Put([]byte(Tail), block.Hash())
}

// LoadFixedFromStorage load FIXED
func (bc *BlockChain) LoadFixedFromStorage() (*Block, error) {
	hash, err := bc.db.Get([]byte(FIXED))
	if err != nil && err != cdb.ErrKeyNotFound {
		return nil, err
	}

	if err == cdb.ErrKeyNotFound {
		if err := bc.StoreFIXEDHashToStorage(bc.genesisBlock); err != nil {
			return nil, err
		}
		return bc.genesisBlock, nil
	}

	return LoadBlockFromStorage(hash, bc)
}

// StoreFIXEDHashToStorage store FIXED block hash
func (bc *BlockChain) StoreFIXEDHashToStorage(block *Block) error {
	return bc.db.Put([]byte(FIXED), block.Hash())
}

func (bc *BlockChain) ChainId() uint32 {
	return bc.chainId
}

func (bc *BlockChain) SetTailBlock(newTail *Block) error {
	if newTail == nil {
		return ErrNilArgument
	}

	if bc.tailBlock != nil {
		oldTail := bc.tailBlock
		if oldTail.Height()+1 != newTail.Height() {
			return errors.New("not invalid tail block")
		}
	}

	if err := bc.db.Put(byteutils.FromUint64(newTail.Height()), newTail.Hash()); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"newtail": newTail,
		}).Debug("Failed to build index by block height.")
	}

	if err := bc.StoreTailHashToStorage(newTail); err != nil { // Refine: rename, delete ToStorage
		return err
	}

	bc.tailBlock = newTail

	return nil
}

// GetBlockOnCanonicalChainByHash check if a block is on canonical chain
func (bc *BlockChain) GetBlockOnCanonicalChainByHash(blockHash byteutils.Hash) *Block {
	blockByHash := bc.GetBlock(blockHash)
	if blockByHash == nil {
		logging.VLog().WithFields(logrus.Fields{
			"hash": blockHash.Hex(),
			"tail": bc.tailBlock,
			"err":  "cannot find block with the given hash in local storage",
		}).Debug("Failed to check a block on canonical chain.")
		return nil
	}
	blockByHeight := bc.GetBlockOnCanonicalChainByHeight(blockByHash.Height())
	if blockByHeight == nil {
		logging.VLog().WithFields(logrus.Fields{
			"height": blockByHash.Height(),
			"tail":   bc.tailBlock,
			"err":    "cannot find block with the given height in local storage",
		}).Debug("Failed to check a block on canonical chain.")
		return nil
	}
	if !blockByHeight.Hash().Equals(blockByHash.Hash()) {
		logging.VLog().WithFields(logrus.Fields{
			"blockByHash":   blockByHash,
			"blockByHeight": blockByHeight,
			"tail":          bc.tailBlock,
			"err":           "block with the given hash isn't on canonical chain",
		}).Debug("Failed to check a block on canonical chain.")
		return nil
	}
	return blockByHeight
}

// GetBlock return block of given hash from local storage and detachedBlocks.
func (bc *BlockChain) GetBlock(hash byteutils.Hash) *Block {
	v, _ := bc.cachedBlocks.Get(hash.Hex())
	if v == nil {
		block, err := LoadBlockFromStorage(hash, bc)
		if err != nil {
			return nil
		}
		return block
	}

	block := v.(*Block)
	return block
}

// GetBlockOnCanonicalChainByHeight return block in given height
func (bc *BlockChain) GetBlockOnCanonicalChainByHeight(height uint64) *Block {

	if height > bc.tailBlock.Height() {
		return nil
	}

	blockHash, err := bc.db.Get(byteutils.FromUint64(height))
	if err != nil {
		return nil
	}
	return bc.GetBlock(blockHash)
}

// PutVerifiedNewBlocks put verified new blocks and tails.
func (bc *BlockChain) putVerifiedNewBlocks(parent *Block, allBlocks, tailBlocks []*Block) error {
	for _, v := range allBlocks {
		bc.cachedBlocks.Add(v.Hash().Hex(), v)
		if err := bc.StoreBlockToStorage(v); err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"block": v,
				"err":   err,
			}).Debug("Failed to store the verified block.")
			return err
		}

		logging.VLog().WithFields(logrus.Fields{
			"block": v,
		}).Info("Accepted the new block on chain")

		//metricsBlockOnchainTimer.Update(time.Duration(time.Now().Unix() - v.Timestamp()))
		//for _, tx := range v.transactions {
		//	metricsTxOnchainTimer.Update(time.Duration(time.Now().Unix() - tx.Timestamp()))
		//}
	}
	for _, v := range tailBlocks {
		bc.detachedTailBlocks.Add(v.Hash().Hex(), v)
	}

	bc.detachedTailBlocks.Remove(parent.Hash().Hex())

	return nil
}

// StoreBlockToStorage store block
func (bc *BlockChain) StoreBlockToStorage(block *Block) error {
	pbBlock, err := block.ToProto()
	if err != nil {
		return err
	}
	value, err := proto.Marshal(pbBlock)
	if err != nil {
		return err
	}
	err = bc.db.Put(block.Hash(), value)
	if err != nil {
		return err
	}
	return nil
}

func (bc *BlockChain) BlockPool() *BlockPool      { return bc.bkPool }
func (bc *BlockChain) TxPool() *TxPool            { return bc.txPool }
func (bc *BlockChain) Consensus() Consensus       { return bc.consensus }
func (bc *BlockChain) TailBlock() *Block          { return bc.tailBlock }
func (bc *BlockChain) GenesisBlock() *Block       { return bc.genesisBlock }
func (bc *BlockChain) FixedBlock() *Block         { return bc.fixedBlock }
func (bc *BlockChain) CurrentBlock() *Block       { return bc.currentBlock }
func (bc *BlockChain) SetFixedBlock(block *Block) { bc.fixedBlock = block }

func (bc *BlockChain) LoadBlockFromStorage(blockHash byteutils.Hash) *Block {
	value, err := bc.db.Get(blockHash)
	if err != nil {
		return nil
	}
	pbBlock := new(corepb.Block)
	block := new(Block)
	if err = proto.Unmarshal(value, pbBlock); err != nil {
		return nil
	}
	if err = block.FromProto(pbBlock); err != nil {
		return nil
	}
	return block
}

// StartActiveSync start active sync task
func (bc *BlockChain) StartActiveSync() bool {
	if bc.sync.StartActiveSync() {
		bc.consensus.SuspendMining()
		go func() {
			bc.sync.WaitingForFinish()
			bc.consensus.ResumeMining()
		}()
		return true
	}
	return false
}

// IsActiveSyncing returns true if being syncing
func (bc *BlockChain) IsActiveSyncing() bool {
	return bc.sync.IsActiveSyncing()
}

// SetSyncEngine set sync engine
func (bc *BlockChain) SetSyncEngine(syncEngine Synchronize) {
	bc.sync = syncEngine
}

// Start start loop.
func (bc *BlockChain) Start() {
	logging.CLog().Info("Starting BlockChain...")
	go bc.loop()
}

func (bc *BlockChain) loop() {
	logging.CLog().Info("Started BlockChain.")
	timerChan := time.NewTicker(5 * time.Second).C
	for {
		select {
		case <-bc.quitCh:
			logging.CLog().Info("Stopped BlockChain.")
			return
		case <-timerChan:
			bc.Consensus().UpdateFixedBlock()
		}
	}
}

// Stop stop loop.
func (bc *BlockChain) Stop() {
	logging.CLog().Info("Stopping BlockChain...")
	bc.quitCh <- 0
}
