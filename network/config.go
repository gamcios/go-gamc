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
	"gamc.pro/gamcio/go-gamc/conf"
	"gamc.pro/gamcio/go-gamc/util/config"
	"fmt"
	"github.com/multiformats/go-multiaddr"
	"time"
)

// const
const (
	DefaultBucketCapacity         = 64
	DefaultRoutingTableMaxLatency = 10
	DefaultPrivateKeyPath         = "conf/network/key"
	DefaultMaxSyncNodes           = 64
	DefaultChainID                = 1
	DefaultRoutingTableDir        = ""
	DefaultMaxStreamNum           = 200
	DefaultReservedStreamNum      = 20
)

// Default Configuration in P2P network
var (
	DefaultListen = []string{"0.0.0.0:8568"}

	RouteTableSyncLoopInterval     = 30 * time.Second
	RouteTableSaveToDiskInterval   = 3 * 60 * time.Second
	RouteTableCacheFileName        = "routetable.cache"
	RouteTableInternalNodeFileName = "conf/internal.txt"

	MaxPeersCountForSyncResp = 32
)

type Config struct {
	Bucketsize           int
	Latency              time.Duration
	BootNodes            []multiaddr.Multiaddr
	PrivateKeyPath       string
	Listen               []string
	MaxSyncNodes         int
	ChainID              uint32
	RoutingTableDir      string
	StreamLimits         uint32
	ReservedStreamLimits uint32
}

type NetConfig struct {
	Seed                    []string `yaml:"seed"`
	Listen                  []string `yaml:"listen"`
	NetworkId               uint32   `yaml:"network_id"`
	PrivateKey              string   `yaml:"private_key"`
	StreamLimits            uint32   `yaml:"stream_limits"`
	ReservedStreamLimits    uint32   `yaml:"reserved_stream_limits"`
	RouteTableCacheFileName string   `yaml:"route_table_cache_filename"`
}

func GetNetConfig(conf *config.Config) *NetConfig {
	netcfg := new(NetConfig)
	conf.GetObject("network", netcfg)
	return netcfg
}

// NewP2PConfig return new config object.
func NewP2PConfig(cfg *config.Config) *Config {
	netcfg := GetNetConfig(cfg)
	if netcfg == nil {
		panic("Failed to find network config in config file.")
	}
	chaincfg := conf.GetChainConfig(cfg)
	if chaincfg == nil {
		panic("Failed to find chain config in config file.")
	}
	config := NewConfigFromDefaults()

	// listen.
	if len(netcfg.Listen) == 0 {
		panic("Missing network.listen config.")
	}
	if err := verifyListenAddress(netcfg.Listen); err != nil {
		panic(fmt.Sprintf("Invalid network.listen config: err is %s, config value is %s.", err, netcfg.Listen))
	}
	config.Listen = netcfg.Listen

	// private key path.
	if checkPathConfig(netcfg.PrivateKey) == false {
		panic(fmt.Sprintf("The network private key path %s is not exist.", netcfg.PrivateKey))
	}
	config.PrivateKeyPath = netcfg.PrivateKey

	// Chain ID.
	config.ChainID = chaincfg.ChainId

	// routing table dir.
	if checkPathConfig(chaincfg.Datadir) == false {
		panic(fmt.Sprintf("The chain data directory %s is not exist.", chaincfg.Datadir))
	}
	config.RoutingTableDir = chaincfg.Datadir

	// seed server address.
	seeds := netcfg.Seed
	if len(seeds) > 0 {
		config.BootNodes = make([]multiaddr.Multiaddr, len(seeds))
		for i, v := range seeds {
			addr, err := multiaddr.NewMultiaddr(v)
			if err != nil {
				panic(fmt.Sprintf("Invalid seed address config: err is %s, config value is %s.", err, v))
			}
			config.BootNodes[i] = addr
		}
	}
	// max stream limits
	if netcfg.StreamLimits > 0 {
		config.StreamLimits = netcfg.StreamLimits
	}

	if netcfg.ReservedStreamLimits > 0 {
		config.ReservedStreamLimits = netcfg.ReservedStreamLimits
	}

	return config
}

// NewConfigFromDefaults return new config from defaults.
func NewConfigFromDefaults() *Config {
	return &Config{
		DefaultBucketCapacity,
		DefaultRoutingTableMaxLatency,
		[]multiaddr.Multiaddr{},
		DefaultPrivateKeyPath,
		DefaultListen,
		DefaultMaxSyncNodes,
		DefaultChainID,
		DefaultRoutingTableDir,
		DefaultMaxStreamNum,
		DefaultReservedStreamNum,
	}
}
