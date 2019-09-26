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

package sync

import (
	"gamc.pro/gamcio/go-gamc/core"
	net "gamc.pro/gamcio/go-gamc/network"
	"gamc.pro/gamcio/go-gamc/storage/cdb"
	syncpb "gamc.pro/gamcio/go-gamc/sync/pb"
	"gamc.pro/gamcio/go-gamc/trie"
	"gamc.pro/gamcio/go-gamc/util/byteutils"
	"gamc.pro/gamcio/go-gamc/util/logging"
	"errors"
	"github.com/gogo/protobuf/proto"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

// Errors
var (
	ErrInvalidChainSyncMessageData     = errors.New("invalid ChainSync message data")
	ErrInvalidChainGetChunkMessageData = errors.New("invalid ChainGetChunk message data")
)

// Service manage sync tasks
type Service struct {
	blockChain *core.BlockChain
	netService net.Service
	chunk      *Chunk
	quitCh     chan bool
	messageCh  chan net.Message

	activeTask      *Task
	activeTaskMutex sync.Mutex
}

// NewService return new Service.
func NewService(blockChain *core.BlockChain, netService net.Service) *Service {
	return &Service{
		blockChain: blockChain,
		netService: netService,
		chunk:      NewChunk(blockChain),
		quitCh:     make(chan bool, 1),
		activeTask: nil,
		messageCh:  make(chan net.Message, 128),
	}
}

// Start start sync service.
func (ss *Service) Start() {
	logging.VLog().Info("Starting Sync Service.")

	// register the network handler.
	netService := ss.netService
	netService.Register(net.NewSubscriber(ss, ss.messageCh, false, net.ChunkHeadersRequest, net.MessageWeightZero))
	netService.Register(net.NewSubscriber(ss, ss.messageCh, false, net.ChunkHeadersResponse, net.MessageWeightChainChunks))
	netService.Register(net.NewSubscriber(ss, ss.messageCh, false, net.ChunkDataRequest, net.MessageWeightZero))
	netService.Register(net.NewSubscriber(ss, ss.messageCh, false, net.ChunkDataResponse, net.MessageWeightChainChunkData))

	// start loop().
	go ss.startLoop()
}

func (ss *Service) startLoop() {
	logging.CLog().Info("Started Sync Service.")
	timerChan := time.NewTicker(time.Second).C

	for {
		select {
		case <-timerChan:
			metricsCachedSync.Update(int64(len(ss.messageCh)))
		case <-ss.quitCh:
			if ss.activeTask != nil {
				ss.activeTask.Stop()
			}
			logging.CLog().Info("Stopped Sync Service.")
			return
		case message := <-ss.messageCh:
			switch message.MessageType() {
			case net.ChunkHeadersRequest:
				ss.onChunkHeadersRequest(message)
			case net.ChunkHeadersResponse:
				ss.onChunkHeadersResponse(message)
			case net.ChunkDataRequest:
				ss.onChunkDataRequest(message)
			case net.ChunkDataResponse:
				ss.onChunkDataResponse(message)
			default:
				logging.VLog().WithFields(logrus.Fields{
					"messageName": message.MessageType(),
				}).Warn("Received unknown message.")
			}
		}
	}
}

// IsActiveSyncing return if there is active task now
func (ss *Service) IsActiveSyncing() bool {
	if ss.activeTask == nil {
		return false
	}

	return true
}

func (ss *Service) onChunkHeadersRequest(message net.Message) {
	if ss.IsActiveSyncing() {
		return
	}

	// handle ChunkHeadersRequest message.
	chunkSync := new(syncpb.Sync)
	err := proto.Unmarshal(message.Data(), chunkSync)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
			"pid": message.MessageFrom(),
		}).Debug("Invalid ChunkHeadersRequest message data.")
		ss.netService.ClosePeer(message.MessageFrom(), ErrInvalidChainSyncMessageData)
		return
	}

	// generate ChunkHeaders message.
	chunks, err := ss.chunk.generateChunkHeaders(chunkSync.TailBlockHash)
	if err != nil && err != ErrTooSmallGapToSync {
		logging.VLog().WithFields(logrus.Fields{
			"err":  err,
			"pid":  message.MessageFrom(),
			"hash": byteutils.Hex(chunkSync.TailBlockHash),
		}).Debug("Failed to generate chunk headers.")
		return
	}

	ss.chunkHeadersResponse(message.MessageFrom(), chunks)
}

func (c *Chunk) generateChunkHeaders(syncpointHash byteutils.Hash) (*syncpb.ChunkHeaders, error) {
	syncpoint := c.blockChain.GetBlockOnCanonicalChainByHash(syncpointHash)
	if syncpoint == nil {
		logging.VLog().WithFields(logrus.Fields{
			"syncpointHash": syncpointHash.Hex(),
		}).Debug("Failed to find the block on canonical chain")
		return nil, ErrCannotFindBlockByHash
	}
	tail := c.blockChain.TailBlock()
	if int(tail.Height())-int(syncpoint.Height()) <= core.ChunkSize {
		logging.VLog().WithFields(logrus.Fields{
			"err": ErrTooSmallGapToSync,
		}).Debug("Failed to generate sync blocks meta info")
		return &syncpb.ChunkHeaders{}, ErrTooSmallGapToSync
	}

	var chunkHeaders []*syncpb.ChunkHeader
	stor, err := cdb.NewMemoryStorage()
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Debug("Failed to create memory storage")
		return nil, err
	}
	chunksTrie, err := trie.NewTrie(nil, stor, false)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Debug("Failed to create merkle tree")
		return nil, err
	}

	startChunk := (syncpoint.Height() - 1) / core.ChunkSize
	endChunk := (tail.Height() - 1) / core.ChunkSize
	curChunk := startChunk
	for curChunk < endChunk && curChunk-startChunk < MaxChunkPerSyncRequest {
		var headers [][]byte
		blocksTrie, err := trie.NewTrie(nil, stor, false)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err": err,
			}).Debug("Failed to create merkle tree")
			return nil, err
		}

		startHeight := curChunk*core.ChunkSize + 2
		endHeight := (curChunk+1)*core.ChunkSize + 2
		curHeight := startHeight
		for curHeight < endHeight {
			block := c.blockChain.GetBlockOnCanonicalChainByHeight(curHeight)
			if block == nil {
				logging.VLog().WithFields(logrus.Fields{
					"height": curHeight + 1,
				}).Debug("Failed to find the block on canonical chain.")
				return nil, ErrCannotFindBlockByHeight
			}
			headers = append(headers, block.Hash())
			_, _ = blocksTrie.Put(block.Hash(), block.Hash())
			curHeight++
		}
		chunkHeaders = append(chunkHeaders, &syncpb.ChunkHeader{Headers: headers, Root: blocksTrie.RootHash()})
		_, _ = chunksTrie.Put(blocksTrie.RootHash(), blocksTrie.RootHash())

		curChunk++
	}

	logging.VLog().WithFields(logrus.Fields{
		"syncpoint": syncpoint,
		"start":     startChunk,
		"end":       endChunk,
		"limit":     MaxChunkPerSyncRequest,
		"synced":    len(chunkHeaders),
	}).Debug("Succeed to generate chunks meta info.")
	return &syncpb.ChunkHeaders{ChunkHeaders: chunkHeaders, Root: chunksTrie.RootHash()}, nil
}

func (ss *Service) chunkHeadersResponse(peerID string, chunks *syncpb.ChunkHeaders) {
	data, err := proto.Marshal(chunks)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Debug("Failed to marshal syncpb.ChunkHeaders.")
		return
	}

	_ = ss.netService.SendMessageToPeer(net.ChunkHeadersResponse, data, net.MessagePriorityLow, peerID)
}

func (ss *Service) onChunkHeadersResponse(message net.Message) {
	if ss.activeTask == nil {
		return
	}

	ss.activeTask.processChunkHeaders(message)
}

func (ss *Service) onChunkDataRequest(message net.Message) {
	if ss.IsActiveSyncing() {
		return
	}

	// handle ChunkDataRequest message.
	chunkHeader := new(syncpb.ChunkHeader)
	err := proto.Unmarshal(message.Data(), chunkHeader)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
			"pid": message.MessageFrom(),
		}).Debug("Invalid ChainGetChunk message data.")
		ss.netService.ClosePeer(message.MessageFrom(), ErrInvalidChainGetChunkMessageData)
		return
	}

	chunkData, err := ss.chunk.generateChunkData(chunkHeader)
	if err != nil {
		if err == ErrWrongChunkHeaderRootHash {
			ss.netService.ClosePeer(message.MessageFrom(), err)
		}
		return
	}

	ss.chunkDataResponse(message.MessageFrom(), chunkData)
}

func (ss *Service) chunkDataResponse(peerID string, chunkData *syncpb.ChunkData) {
	data, err := proto.Marshal(chunkData)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Debug("Failed to marshal syncpb.ChunkData.")
		return
	}

	_ = ss.netService.SendMessageToPeer(net.ChunkDataResponse, data, net.MessagePriorityLow, peerID)
}

func (ss *Service) onChunkDataResponse(message net.Message) {
	if ss.activeTask == nil {
		return
	}

	ss.activeTask.processChunkData(message)
}

// StartActiveSync starts an active sync task
func (ss *Service) StartActiveSync() bool {
	// lock.
	ss.activeTaskMutex.Lock()
	defer ss.activeTaskMutex.Unlock()

	if ss.IsActiveSyncing() {
		return false
	}

	ss.activeTask = NewTask(ss.blockChain, ss.netService, ss.chunk)
	ss.activeTask.Start()

	logging.CLog().WithFields(logrus.Fields{
		"syncpoint": ss.activeTask.syncPointBlock,
	}).Info("Started Active Sync Task.")
	return true
}

// Stop stop sync service.
func (ss *Service) Stop() {
	// deregister the network handler.
	netService := ss.netService
	netService.Deregister(net.NewSubscriber(ss, ss.messageCh, false, net.ChunkHeadersRequest, net.MessageWeightZero))
	netService.Deregister(net.NewSubscriber(ss, ss.messageCh, false, net.ChunkHeadersResponse, net.MessageWeightChainChunks))
	netService.Deregister(net.NewSubscriber(ss, ss.messageCh, false, net.ChunkDataRequest, net.MessageWeightZero))
	netService.Deregister(net.NewSubscriber(ss, ss.messageCh, false, net.ChunkDataResponse, net.MessageWeightChainChunkData))

	ss.StopActiveSync()

	ss.quitCh <- true
}

// StopActiveSync stops current sync task
func (ss *Service) StopActiveSync() {
	if ss.activeTask == nil {
		return
	}

	ss.activeTask.Stop()
	ss.activeTask = nil
}

// WaitingForFinish wait for finishing current sync task
func (ss *Service) WaitingForFinish() {
	if ss.activeTask == nil {
		return
	}

	<-ss.activeTask.statusCh

	logging.CLog().WithFields(logrus.Fields{
		"tail": ss.blockChain.TailBlock(),
	}).Info("Active Sync Task Finished.")

	ss.activeTask = nil
}
