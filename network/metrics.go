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
	"gamc.pro/gamcio/go-gamc/metrics"
	"fmt"
)

// Metrics map for different in/out network msg types
var (
	metricsPacketsIn = metrics.NewMeter("gamc.net.packets.in")
	metricsBytesIn   = metrics.NewMeter("gamc.net.bytes.in")

	metricsPacketsOut = metrics.NewMeter("gamc.net.packets.out")
	metricsBytesOut   = metrics.NewMeter("gamc.net.bytes.out")
)

func metricsPacketsInByMessageName(messageName string, size uint64) {
	meter := metrics.NewMeter(fmt.Sprintf("gamc.net.packets.in.%s", messageName))
	meter.Mark(1)

	meter = metrics.NewMeter(fmt.Sprintf("gamc.net.bytes.in.%s", messageName))
	meter.Mark(int64(size))
}

func metricsPacketsOutByMessageName(messageName string, size uint64) {
	meter := metrics.NewMeter(fmt.Sprintf("gamc.net.packets.out.%s", messageName))
	meter.Mark(1)

	meter = metrics.NewMeter(fmt.Sprintf("gamc.net.bytes.out.%s", messageName))
	meter.Mark(int64(size))
}
