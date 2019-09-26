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
	"gamc.pro/gamcio/go-gamc/util/byteutils"
	"gamc.pro/gamcio/go-gamc/util/logging"
	"encoding/hex"
	"github.com/gogo/protobuf/proto"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
	"math/big"
	"time"
)

// Transaction
type Transaction struct {
	hash      byteutils.Hash
	from      *Address
	to        *Address
	value     *big.Int
	nonce     uint64
	chainId   uint32
	fee       *big.Int
	timestamp int64
	data      *corepb.Data
	priority  uint32
	sign      *corepb.Signature
}

// Transactions is an alias of Transaction array.
type Transactions []*Transaction

// NewTransaction
func NewTransaction(nonce uint64, from, to *Address, amount *big.Int) *Transaction {
	if to == nil { // create contract tx
		return newTransaction(nonce, from, nil, amount)
	}
	return newTransaction(nonce, from, to, amount)
}

// NewContractCreation
func NewContractCreation(nonce uint64, from *Address, amount *big.Int) *Transaction {
	return newTransaction(nonce, from, nil, amount)
}

// newTransaction
func newTransaction(nonce uint64, from, to *Address, amount *big.Int) *Transaction {
	tx := Transaction{
		nonce:     nonce,
		value:     new(big.Int),
		from:      from,
		to:        to,
		timestamp: time.Now().Unix(),
	}
	if amount != nil {
		tx.value.Set(amount)
	}

	return &tx
}

func (tx *Transaction) Nonce() uint64        { return tx.nonce }
func (tx *Transaction) Hash() byteutils.Hash { return tx.hash }
func (tx *Transaction) Timestamp() int64     { return tx.timestamp }

// TxFrom
func (tx *Transaction) From() *Address {
	if tx.from == nil {
		return nil
	}
	from := *tx.from
	return &from
}

// TxTo
func (tx *Transaction) To() *Address {
	if tx.to == nil {
		return nil
	}
	to := *tx.to
	return &to
}

// ToProto converts domain Tx to proto Tx
func (tx *Transaction) ToProto() (proto.Message, error) {
	return &corepb.Transaction{
		Hash:      tx.hash,
		From:      tx.from.address,
		To:        tx.to.address,
		Value:     tx.value.Bytes(),
		Nonce:     tx.nonce,
		ChainId:   tx.chainId,
		Fee:       tx.fee.Bytes(),
		Timestamp: tx.timestamp,
		Data:      tx.data,
		Priority:  tx.priority,
		Sign:      tx.sign,
	}, nil
}

// FromProto converts proto Tx to domain Tx
func (tx *Transaction) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*corepb.Transaction); ok {
		if msg != nil {
			tx.hash = msg.Hash
			from, err := AddressParseFromBytes(msg.From)
			if err != nil {
				return ErrInvalidProtoToTransaction
			}
			tx.from = from
			to, err := AddressParseFromBytes(msg.To)
			if err != nil {
				return ErrInvalidProtoToTransaction
			}
			tx.to = to
			tx.value = new(big.Int).SetBytes(msg.Value)
			tx.nonce = msg.Nonce
			tx.chainId = msg.ChainId
			tx.fee = new(big.Int).SetBytes(msg.Fee)
			tx.timestamp = msg.Timestamp
			tx.data = msg.Data
			tx.priority = msg.Priority
			tx.sign = msg.Sign
		}
		return ErrInvalidProtoToTransaction
	}
	return ErrInvalidProtoToTransaction
}

// VerifyIntegrity return transaction verify result, including Hash and Signature.
func (tx *Transaction) VerifyIntegrity(chainId uint32) error {
	// check ChainID.
	if tx.chainId != chainId {
		return ErrInvalidChainID
	}

	// check Hash.
	wantedHash, err := tx.calcHash()
	if err != nil {
		return err
	}
	if wantedHash.Equals(tx.hash) == false {
		return ErrInvalidTransactionHash
	}

	// check Signature.
	return tx.verifySign()

}

func (tx *Transaction) verifySign() error {
	signer, err := NewAddressFromPublicKey(tx.sign.Signer)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"tx.sign.Signer": hex.EncodeToString(tx.sign.Signer),
		}).Debug("Failed to verify tx's sign.")
		return ErrInvalidPublicKey
	}
	if !tx.from.Equals(signer) {
		logging.VLog().WithFields(logrus.Fields{
			"signer":  signer.String(),
			"tx.from": tx.from,
		}).Debug("Failed to verify tx's sign.")
		return ErrInvalidTransactionSigner
	}
	return nil
}

// HashTransaction hash the transaction.
func (tx *Transaction) calcHash() (byteutils.Hash, error) {
	hasher := sha3.New256()

	data, err := proto.Marshal(tx.data)
	if err != nil {
		return nil, err
	}

	hasher.Write(tx.from.address)
	hasher.Write(tx.to.address)
	hasher.Write(tx.value.Bytes())
	hasher.Write(byteutils.FromUint64(tx.nonce))
	hasher.Write(byteutils.FromUint32(tx.chainId))
	hasher.Write(tx.fee.Bytes())
	hasher.Write(byteutils.FromInt64(tx.timestamp))
	hasher.Write(data)
	hasher.Write(byteutils.FromUint32(tx.priority))

	return hasher.Sum(nil), nil
}
