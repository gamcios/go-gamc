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
	"gamc.pro/gamcio/go-gamc/crypto/keystore"
	"gamc.pro/gamcio/go-gamc/storage/cdb"
	"gamc.pro/gamcio/go-gamc/util/byteutils"
	"gamc.pro/gamcio/go-gamc/util/logging"
	"github.com/gogo/protobuf/proto"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
	"time"
)

var (
	BlockHashLength = 32
)

// Block
type Block struct {
	header       *BlockHeader
	transactions []*Transaction //
	worldState   WorldState
	db           cdb.Storage
}

// NewBlock
func NewBlock(header *BlockHeader, txs []*Transaction) *Block {
	block := Block{
		header: header,
	}

	return &block
}

// CalcHash
func (b *Block) CalcHash() byteutils.Hash {
	h, _ := b.calcHash()
	return h
}

func (b *Block) Header() *BlockHeader          { return b.header }
func (b *Block) Hash() byteutils.Hash          { return b.header.hash }
func (b *Block) Timestamp() int64              { return b.header.Timestamp() }
func (b *Block) Height() uint64                { return b.header.Height() }
func (b *Block) SetHeight(height uint64)       { b.header.height = height }
func (b *Block) SetTimestamp(time int64)       { b.header.timestamp = time }
func (b *Block) SetParent(hash byteutils.Hash) { b.header.parentHash = hash }
func (b *Block) Witness() []*Witness           { return b.header.Witnesses() }
func (b *Block) SetWorldState(parent *Block)   { b.worldState, _ = parent.WorldState().Clone() }

// CopyHeader creates a deep copy of a block header to prevent side effects from
// modifying a header variable.
func CopyHeader(h *BlockHeader) *BlockHeader {
	cpy := *h

	if len(h.Extra()) > 0 {
		cpy.extra = make([]byte, len(h.Extra()))
		copy(cpy.extra, h.Extra())
	}
	return &cpy
}

//
func NewBlockWithHeader(header *BlockHeader) *Block {
	return &Block{header: CopyHeader(header)}
}

// LoadBlockFromStorage return a block from storage
func LoadBlockFromStorage(hash byteutils.Hash, chain *BlockChain) (*Block, error) {
	if chain == nil {
		return nil, ErrNilArgument
	}

	value, err := chain.db.Get(hash)
	if err != nil {
		return nil, err
	}
	pbBlock := new(corepb.Block)
	block := new(Block)
	if err = proto.Unmarshal(value, pbBlock); err != nil {
		return nil, err
	}
	if err = block.FromProto(pbBlock); err != nil {
		return nil, err
	}
	block.worldState, err = NewWorldState(chain.db)
	if err != nil {
		return nil, err
	}
	if err := block.WorldState().LoadAccountsRoot(block.StateRoot()); err != nil {
		return nil, err
	}
	if err := block.WorldState().LoadTxsRoot(block.TxsRoot()); err != nil {
		return nil, err
	}

	block.db = chain.db
	return block, nil
}

// ToProto converts domain Block into proto Block
func (b *Block) ToProto() (proto.Message, error) {
	header, err := b.header.ToProto()
	if err != nil {
		return nil, err
	}
	if header, ok := header.(*corepb.BlockHeader); ok {
		txs := make([]*corepb.Transaction, len(b.transactions))
		for idx, v := range b.transactions {
			tx, err := v.ToProto()
			if err != nil {
				return nil, err
			}
			if tx, ok := tx.(*corepb.Transaction); ok {
				txs[idx] = tx
			} else {
				return nil, ErrInvalidProtoToTransaction
			}
		}
		return &corepb.Block{
			Hash:   b.Hash(),
			Header: header,
			Body:   txs,
		}, nil
	}
	return nil, ErrInvalidProtoToBlock
}

// HashPbBlock return the hash of pb block.
func HashPbBlock(pbBlock *corepb.Block) (byteutils.Hash, error) {
	block := new(Block)
	if err := block.FromProto(pbBlock); err != nil {
		return nil, err
	}
	return block.calcHash()
}

// CalcHash calculate the hash of block.
func (b *Block) calcHash() (byteutils.Hash, error) {
	hasher := sha3.New256()

	pbPsec, err := b.header.psecData.ToProto()
	if err != nil {
		return nil, err
	}
	psecData, err := proto.Marshal(pbPsec)
	if err != nil {
		return nil, err
	}

	hasher.Write(b.ParentHash())
	hasher.Write(b.Coinbase().Bytes())
	hasher.Write(byteutils.FromUint32(b.header.chainId))
	hasher.Write(byteutils.FromInt64(b.header.timestamp))
	hasher.Write(b.header.witnessreward.Bytes())
	for _, v := range b.header.witnesses {
		pbWitness, err := v.ToProto()
		if err != nil {
			return nil, err
		}
		witness, err := proto.Marshal(pbWitness)
		if err != nil {
			return nil, err
		}
		hasher.Write(witness)
	}
	hasher.Write(b.StateRoot())
	hasher.Write(b.TxsRoot())
	hasher.Write(psecData)
	hasher.Write(b.header.extra)

	for _, tx := range b.transactions {
		hasher.Write(tx.Hash())
	}

	return hasher.Sum(nil), nil
}

// FromProto converts proto Block to domain Block
func (b *Block) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*corepb.Block); ok {
		if msg != nil {
			b.header = new(BlockHeader)
			if err := b.header.FromProto(msg.Header); err != nil {
				return err
			}
			b.transactions = make(Transactions, len(msg.Body))
			for idx, v := range msg.Body {
				if v != nil {
					tx := new(Transaction)
					if err := tx.FromProto(v); err != nil {
						return err
					}
					b.transactions[idx] = tx
				} else {
					return ErrInvalidProtoToTransaction
				}
			}
			return nil
		}
		return ErrInvalidProtoToBlock
	}
	return ErrInvalidProtoToBlock
}

// VerifyIntegrity verify block's hash, txs' integrity and consensus acceptable.
func (b *Block) VerifyIntegrity(chainId uint32, consensus Consensus) error {
	if consensus == nil {
		//metricsInvalidBlock.Inc(1)
		return ErrNilArgument
	}

	// check ChainID.
	if b.header.chainId != chainId {
		logging.VLog().WithFields(logrus.Fields{
			"expect": chainId,
			"actual": b.header.chainId,
		}).Info("Failed to check chainid.")
		//metricsInvalidBlock.Inc(1)
		return ErrInvalidBlockHeaderChainID
	}

	// verify transactions integrity.
	for _, tx := range b.transactions {
		if err := tx.VerifyIntegrity(b.header.chainId); err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"tx":  tx,
				"err": err,
			}).Info("Failed to verify tx's integrity.")
			//metricsInvalidBlock.Inc(1)
			return err
		}
	}

	// verify block hash.
	wantedHash, err := b.calcHash()
	if err != nil {
		return err
	}
	if !wantedHash.Equals(b.Hash()) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": wantedHash,
			"actual": b.Hash(),
			"err":    err,
		}).Info("Failed to check block's hash.")
		//metricsInvalidBlock.Inc(1)
		return ErrInvalidBlockHash
	}

	//verify the block is acceptable by consensus.
	if err := consensus.VerifyBlock(b); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"block": b,
			"err":   err,
		}).Info("Failed to verify block.")
		metricsInvalidBlock.Inc(1)
		return err
	}

	return nil
}

// LinkParentBlock link parent block, return true if hash is the same; false otherwise.
func (b *Block) LinkParentBlock(chain *BlockChain, parentBlock *Block) error {
	if !b.ParentHash().Equals(parentBlock.Hash()) {
		return ErrLinkToWrongParentBlock
	}

	var err error
	if b.worldState, err = parentBlock.WorldState().Clone(); err != nil {
		return ErrCloneAccountState
	}

	//elapsedSecond := block.Timestamp() - parentBlock.Timestamp()
	//consensusState, err := parentBlock.worldState.NextConsensusState(elapsedSecond)
	//if err != nil {
	//	return err
	//}
	//block.WorldState().SetConsensusState(consensusState)

	b.header.height = parentBlock.header.height + 1
	b.db = parentBlock.db

	return nil
}

// VerifyExecution execute the block and verify the execution result.
func (b *Block) VerifyExecution() error {
	startAt := time.Now().Unix()

	if err := b.Begin(); err != nil {
		return err
	}

	beganAt := time.Now().Unix()

	if err := b.execute(); err != nil {
		b.RollBack()
		return err
	}

	executedAt := time.Now().Unix()

	if err := b.verifyState(); err != nil {
		b.RollBack()
		return err
	}

	commitAt := time.Now().Unix()

	b.Commit()

	endAt := time.Now().Unix()

	logging.VLog().WithFields(logrus.Fields{
		"start":        startAt,
		"end":          endAt,
		"commit":       commitAt,
		"diff-all":     endAt - startAt,
		"diff-commit":  endAt - commitAt,
		"diff-begin":   beganAt - startAt,
		"diff-execute": executedAt - startAt,
		"diff-verify":  commitAt - executedAt,
		"block":        b,
		"txs":          len(b.Transactions()),
	}).Info("Verify txs.")

	return nil
}

type verifyCtx struct {
	mergeCh chan bool
	block   *Block
}

// Execute block and return result.
func (b *Block) execute() error {
	startAt := time.Now().UnixNano()

	//TODO Reward
	//if err := block.rewardCoinbaseForMint(); err != nil {
	//	return err
	//}

	//context := &verifyCtx{
	//	mergeCh: make(chan bool, 1),
	//	block:   block,
	//}

	//dispatcher := dag.NewDispatcher(block.dependency, parallelNum, int64(VerifyExecutionTimeout), context, func(node *dag.Node, context interface{}) error { // TODO: if system occurs, the block won't be retried any more
	//	ctx := context.(*verifyCtx)
	//	block := ctx.block
	//	mergeCh := ctx.mergeCh
	//
	//	idx := node.Index()
	//	if idx < 0 || idx > len(block.transactions)-1 {
	//		return ErrInvalidDagBlock
	//	}
	//	tx := block.transactions[idx]
	//
	//	logging.VLog().WithFields(logrus.Fields{
	//		"tx.hash": tx.hash,
	//	}).Debug("execute tx.")
	//
	//	metricsTxExecute.Mark(1)
	//
	//	mergeCh <- true
	//	txWorldState, err := block.WorldState().Prepare(tx.Hash().String())
	//	if err != nil {
	//		<-mergeCh
	//		return err
	//	}
	//	<-mergeCh
	//
	//	if _, err = block.ExecuteTransaction(tx, txWorldState); err != nil {
	//		return err
	//	}
	//
	//	mergeCh <- true
	//	if _, err := txWorldState.CheckAndUpdate(); err != nil {
	//		return err
	//	}
	//	<-mergeCh
	//
	//	return nil
	//})

	start := time.Now().UnixNano()

	// TODO Verify block transactions
	//if err := dispatcher.Run(); err != nil {
	//	transactions := []string{}
	//	for k, tx := range block.transactions {
	//		txInfo := fmt.Sprintf("{Index: %d, Tx: %s}", k, tx.String())
	//		transactions = append(transactions, txInfo)
	//	}
	//	logging.VLog().WithFields(logrus.Fields{
	//		"dag": block.dependency.String(),
	//		"txs": transactions,
	//		"err": err,
	//	}).Info("Failed to verify txs in block.")
	//	return err
	//}
	end := time.Now().UnixNano()

	if len(b.transactions) != 0 {
		metricsTxVerifiedTime.Update((end - start) / int64(len(b.transactions)))
	} else {
		metricsTxVerifiedTime.Update(0)
	}

	if err := b.WorldState().Flush(); err != nil {
		return err
	}

	endAt := time.Now().UnixNano()
	metricsBlockVerifiedTime.Update(endAt - startAt)
	metricsTxsInBlock.Update(int64(len(b.transactions)))

	return nil
}

// RollBack a batch task
func (b *Block) RollBack() {
	if err := b.WorldState().RollBack(); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Fatal("Failed to rollback the block")
	}
}

// verifyState return state verify result.
func (b *Block) verifyState() error {
	// verify state root.
	if !byteutils.Equal(b.WorldState().AccountsRoot(), b.StateRoot()) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": b.StateRoot(),
			"actual": b.WorldState().AccountsRoot(),
		}).Info("Failed to verify state.")
		return ErrInvalidBlockStateRoot
	}

	// verify transaction root.
	if !byteutils.Equal(b.WorldState().TxsRoot(), b.TxsRoot()) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": b.TxsRoot(),
			"actual": b.WorldState().TxsRoot(),
		}).Info("Failed to verify txs.")
		return ErrInvalidBlockTxsRoot
	}

	return nil
}

// Commit a batch task
func (b *Block) Commit() {
	if err := b.WorldState().Commit(); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Fatal("Failed to commit the block")
	}
}

// Begin a batch task
func (b *Block) Begin() error {
	return b.WorldState().Begin()
}

// WorldState return the world state of the block
func (b *Block) WorldState() WorldState {
	return b.worldState
}

// StateRoot return state root hash.
func (b *Block) StateRoot() byteutils.Hash {
	return b.header.stateRoot
}

// TxsRoot return txs root hash.
func (b *Block) TxsRoot() byteutils.Hash {
	return b.header.txsRoot
}

// ParentHash return parent hash.
func (b *Block) ParentHash() byteutils.Hash {
	return b.header.parentHash
}

// Coinbase return coinbase
func (b *Block) Coinbase() *Address {
	return b.header.coinbase
}

// Transactions returns block transactions
func (b *Block) Transactions() Transactions {
	return b.transactions
}

// Signature return block's signature
func (b *Block) Signature() *corepb.Signature {
	return b.header.sign
}

// SignHash return block's sign hash
func (b *Block) SignHash() byteutils.Hash {
	return b.header.sign.GetData()
}

func (b *Block) Sign(signature keystore.Signature) error {
	if signature == nil {
		return ErrNilArgument
	}
	sign, err := signature.Sign(b.header.hash)
	if err != nil {
		return err
	}
	b.header.sign = &corepb.Signature{
		Signer: sign.GetSigner(),
		Data:   sign.GetData(),
	}
	return nil
}
