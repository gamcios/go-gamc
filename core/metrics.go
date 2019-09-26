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

import "gamc.pro/gamcio/go-gamc/metrics"

var (
	metricsDuplicatedBlock = metrics.NewCounter("gamc.block.duplicated")
	metricsInvalidBlock    = metrics.NewCounter("gamc.block.invalid")

	metricsTxVerifiedTime    = metrics.NewGauge("gamc.tx.executed")
	metricsTxsInBlock        = metrics.NewGauge("gamc.block.txs")
	metricsBlockVerifiedTime = metrics.NewGauge("gamc.block.executed")

	metricsBlockOnchainTimer = metrics.NewTimer("gamc.block.onchain")
	metricsTxOnchainTimer    = metrics.NewTimer("gamc.transaction.onchain")

	// block_pool metrics
	metricsCachedNewBlock      = metrics.NewGauge("gamc.block.new.cached")
	metricsCachedDownloadBlock = metrics.NewGauge("gamc.block.download.cached")
	metricsLruPoolCacheBlock   = metrics.NewGauge("gamc.block.lru.poolcached")
	metricsLruCacheBlock       = metrics.NewGauge("gamc.block.lru.blocks")
	metricsLruTailBlock        = metrics.NewGauge("gamc.block.lru.tailblock")
)
