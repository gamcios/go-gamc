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
	"github.com/gogo/protobuf/proto"
	"math/big"
)

type Witness struct {
	master   *Address
	follower []*Address
}

func (w *Witness) Master() *Address              { return w.master }
func (w *Witness) SetMaster(master *Address)     { w.master = master }
func (w *Witness) AddFollower(follower *Address) { w.follower = append(w.follower, follower) }

// BlockHeader
type BlockHeader struct {
	chainId       uint32
	witnessreward *big.Int
	coinbase      *Address
	witnesses     []*Witness
	stateRoot     []byte
	txsRoot       []byte
	parentHash    []byte
	psecData      *PsecData
	height        uint64
	timestamp     int64
	hash          []byte
	sign          *corepb.Signature
	extra         []byte
}

func (h *BlockHeader) Witnesses() []*Witness   { return h.witnesses }
func (h *BlockHeader) TxsRoot() []byte         { return h.txsRoot }
func (h *BlockHeader) ParentHash() []byte      { return h.parentHash }
func (h *BlockHeader) PsecData() *PsecData     { return h.psecData }
func (h *BlockHeader) Timestamp() int64        { return h.timestamp }
func (h *BlockHeader) WitnessReward() *big.Int { return h.witnessreward }
func (h *BlockHeader) ChainId() uint32         { return h.chainId }
func (h *BlockHeader) Coinbase() *Address      { return h.coinbase }
func (h *BlockHeader) Extra() []byte           { return h.extra }
func (h *BlockHeader) Sign() *corepb.Signature { return h.sign }
func (h *BlockHeader) Hash() []byte            { return h.hash }
func (h *BlockHeader) Height() uint64          { return h.height }

func (h *BlockHeader) SetParentHash(hash []byte)           { h.hash = hash }
func (h *BlockHeader) SetTimestamp(t int64)                { h.timestamp = t }
func (h *BlockHeader) SetHeight(height uint64)             { h.height = height }
func (h *BlockHeader) SetCoinbase(addr Address)            { h.coinbase = &addr }
func (h *BlockHeader) SetPsecData(pd PsecData)             { h.psecData = &pd }
func (h *BlockHeader) SetWitnessReward(reward int64)       { h.witnessreward = big.NewInt(reward) }
func (h *BlockHeader) SetChainId(id uint32)                { h.chainId = id }
func (h *BlockHeader) SetHash(hash byteutils.Hash)         { h.hash = hash }
func (h *BlockHeader) SetAccountsRoot(hash byteutils.Hash) { h.stateRoot = hash }
func (h *BlockHeader) SetTxsRoot(hash byteutils.Hash)      { h.txsRoot = hash }
func (w *Witness) ToProto() (proto.Message, error) {
	followers := make([][]byte, len(w.follower))
	for idx, v := range w.follower {
		followers[idx] = v.address
	}
	return &corepb.Witness{
		Master:    w.master.address,
		Followers: followers,
	}, nil
}

func (w *Witness) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*corepb.Witness); ok {
		if msg != nil {
			master, err := AddressParseFromBytes(msg.Master)
			if err != nil {
				return ErrInvalidProtoToWitness
			}
			w.master = master
			followers := make([]*Address, len(msg.Followers))
			for idx, v := range msg.Followers {
				follower, err := AddressParseFromBytes(v)
				if err != nil {
					return ErrInvalidProtoToWitness
				}
				followers[idx] = follower
			}
			w.follower = followers
			return nil
		}
		return ErrInvalidProtoToWitness
	}
	return ErrInvalidProtoToWitness
}

// FromProto converts proto BlockHeader to domain BlockHeader
func (h *BlockHeader) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*corepb.BlockHeader); ok {
		if msg != nil {
			h.chainId = msg.ChainId
			h.witnessreward = new(big.Int).SetBytes(msg.WitnessReward)
			coinbase, err := AddressParseFromBytes(msg.Coinbase)
			if err != nil {
				return ErrInvalidProtoToBlockHeader
			}
			h.coinbase = coinbase
			witnesses := make([]*Witness, len(msg.Witnesses))
			for idx, v := range msg.Witnesses {
				witness := new(Witness)
				err := witness.FromProto(v)
				if err != nil {
					return err
				}
				witnesses[idx] = witness
			}
			h.witnesses = witnesses
			h.stateRoot = msg.StateRoot
			h.txsRoot = msg.TxsRoot
			h.parentHash = msg.ParentHash
			psecData := new(PsecData)
			err = psecData.FromProto(msg.PsecData)
			if err != nil {
				return err
			}
			h.psecData = psecData
			h.height = msg.Height
			h.timestamp = msg.Timestamp
			h.hash = msg.Hash
			h.sign = msg.Sign
			h.extra = msg.Extra
			return nil
		}
		return ErrInvalidProtoToBlockHeader
	}
	return ErrInvalidProtoToBlockHeader
}

// ToProto converts domain BlockHeader to proto BlockHeader
func (h *BlockHeader) ToProto() (proto.Message, error) {
	witnesses := make([]*corepb.Witness, len(h.witnesses))
	for idx, v := range h.witnesses {
		witness, err := v.ToProto()
		if err != nil {
			return nil, err
		}
		if witness, ok := witness.(*corepb.Witness); ok {
			witnesses[idx] = witness
		} else {
			return nil, ErrInvalidProtoToWitness
		}
	}

	psecData, err := h.psecData.ToProto()
	if err != nil {
		return nil, err
	}

	if psecData, ok := psecData.(*corepb.PsecData); ok {
		return &corepb.BlockHeader{
			Hash:          h.hash,
			ParentHash:    h.parentHash,
			Coinbase:      h.coinbase.address,
			ChainId:       h.chainId,
			Timestamp:     h.timestamp,
			Height:        h.height,
			WitnessReward: h.witnessreward.Bytes(),
			Witnesses:     witnesses,
			StateRoot:     h.stateRoot,
			TxsRoot:       h.txsRoot,
			PsecData:      psecData,
			Sign:          h.sign,
			Extra:         h.extra,
		}, nil
	} else {
		return nil, ErrInvalidProtoToPsecData
	}

}
