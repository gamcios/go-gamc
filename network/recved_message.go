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

package network

import (
	"gamc.pro/gamcio/go-gamc/util/logging"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/willf/bloom"
	"sync"
)

const (
	// according to https://krisives.github.io/bloom-calculator/
	// Count (n) = 100000, Error (p) = 0.001
	maxCountOfRecvMessageInBloomFiler = 1000000
	bloomFilterOfRecvMessageArgM      = 14377588
	bloomFilterOfRecvMessageArgK      = 10
)

var (
	bloomFilterOfRecvMessage        = bloom.New(bloomFilterOfRecvMessageArgM, bloomFilterOfRecvMessageArgK)
	bloomFilterMutex                sync.Mutex
	countOfRecvMessageInBloomFilter = 0
)

// RecordKey add key to bloom filter.
func RecordKey(key string) {
	bloomFilterMutex.Lock()
	defer bloomFilterMutex.Unlock()

	countOfRecvMessageInBloomFilter++
	if countOfRecvMessageInBloomFilter > maxCountOfRecvMessageInBloomFiler {
		// reset.
		logging.VLog().WithFields(logrus.Fields{
			"countOfRecvMessageInBloomFilter": countOfRecvMessageInBloomFilter,
		}).Debug("reset bloom filter.")
		countOfRecvMessageInBloomFilter = 0
		bloomFilterOfRecvMessage = bloom.New(bloomFilterOfRecvMessageArgM, bloomFilterOfRecvMessageArgK)
	}

	bloomFilterOfRecvMessage.AddString(key)
}

// RecordRecvMessage records received message
func RecordRecvMessage(s *Stream, hash uint32) {
	RecordKey(fmt.Sprintf("%s-%d", s.pid, hash))
}

// HasKey use bloom filter to check if the key exists quickly
func HasKey(key string) bool {
	bloomFilterMutex.Lock()
	defer bloomFilterMutex.Unlock()

	return bloomFilterOfRecvMessage.TestString(key)
}

// HasRecvMessage check if the received message exists before
func HasRecvMessage(s *Stream, hash uint32) bool {
	return HasKey(fmt.Sprintf("%s-%d", s.pid, hash))
}
