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
	"container/heap"
	corepb "gamc.pro/gamcio/go-gamc/core/pb"
	"gamc.pro/gamcio/go-gamc/network"
	"gamc.pro/gamcio/go-gamc/util/logging"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sync"
)

const (
	maxPendingSize = 4096
)

var (
	ErrTxPoolFull = errors.New("tx pool is full")
)

// TxPool
type TxPool struct {
	pending   timeHeap
	queued    timeHeap
	txFilter  map[string]map[uint64]struct{} // address --> map[txNonce]struct{}
	quitCh    chan int
	recvMsgCh chan network.Message
	rw        sync.RWMutex
}

func NewTxPool() *TxPool {
	return &TxPool{
		pending:  timeHeap{},
		queued:   timeHeap{},
		txFilter: make(map[string]map[uint64]struct{}),
		quitCh:   make(chan int),
	}
}

func (pool *TxPool) Start() {
	logging.CLog().WithFields(logrus.Fields{}).Info("Starting TransactionPool...")

	go pool.loop()
}

func (pool *TxPool) loop() {
	for {
		select {
		case <-pool.quitCh:
			logging.CLog().WithFields(logrus.Fields{}).Info("Stopped TxPool.")
			return
		case msg := <-pool.recvMsgCh:
			if msg.MessageType() != MessageTypeNewTx {
				logging.VLog().WithFields(logrus.Fields{
					"messageType": msg.MessageType(),
					"message":     msg,
					"err":         "not new tx msg",
				}).Debug("Received unregistered message.")
				continue
			}
			tx := new(Transaction)
			pbTx := new(corepb.Transaction)
			if err := proto.Unmarshal(msg.Data(), pbTx); err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"msgType": msg.MessageType(),
					"msg":     msg,
					"err":     err,
				}).Debug("Failed to unmarshal data.")
				continue
			}
			if err := tx.FromProto(pbTx); err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"msgType": msg.MessageType(),
					"msg":     msg,
					"err":     err,
				}).Debug("Failed to recover a tx from proto data.")
				continue
			}
			if err := pool.add(tx, true); err != nil {
				logging.VLog().WithFields(logrus.Fields{
					"func":        "TxPool.loop",
					"messageType": msg.MessageType(),
					"transaction": tx,
					"err":         err,
				}).Debug("Failed to push a tx into tx pool.")
				continue
			}

		}
	}
}

// GetAccountTxAmount
func (pool *TxPool) GetAccountTxAmount(address Address) int {
	return len(pool.txFilter[address.String()])
}

// Get
func (pool *TxPool) Get(size int) []*Transaction {
	pool.rw.Lock()
	defer pool.rw.Unlock()
	return pool.pending.PopTxs(size)
}

// AddRemote
func (pool *TxPool) AddRemote(txs []*Transaction) []error {
	return pool.addTxs(txs, false)
}

// AddLocal
func (pool *TxPool) AddLocal(txs []*Transaction) []error {
	return pool.addTxs(txs, true)
}

// Promote
func (pool *TxPool) Promote(size int) {
	pool.promote(size)
}

// isExist
func (pool *TxPool) isExist(tx *Transaction) bool {
	from := *tx.From()
	if _, ok := pool.txFilter[from.String()][tx.Nonce()]; ok {
		return true
	}

	return false
}

// addTxs
func (pool *TxPool) addTxs(txs []*Transaction, local bool) []error {
	errs := make([]error, 0)
	for _, tx := range txs {
		if pool.isExist(tx) {
			continue
		}
		if err := pool.add(tx, local); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

// add
func (pool *TxPool) add(tx *Transaction, local bool) error {
	if len(pool.pending) >= maxPendingSize {
		return ErrTxPoolFull
	}

	if local {
		heap.Push(&pool.pending, tx)
	} else {
		heap.Push(&pool.queued, tx)
	}

	return nil
}

// promote
func (pool *TxPool) promote(size int) {
	for i := 0; i < size; i++ {
		tx := heap.Pop(&pool.queued)
		heap.Push(&pool.pending, tx)
	}
}

// removeTx
func (pool *TxPool) removeTx(tx *Transaction) {
	if !pool.isExist(tx) {
		return
	}
	from := *tx.From()
	delete(pool.txFilter[from.String()], tx.Nonce())
}
