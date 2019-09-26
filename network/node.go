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
	"context"
	"gamc.pro/gamcio/go-gamc/util/logging"
	"errors"
	"fmt"
	csms "github.com/libp2p/go-conn-security-multistream"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/metrics"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	secio "github.com/libp2p/go-libp2p-secio"
	swarm "github.com/libp2p/go-libp2p-swarm"
	tptu "github.com/libp2p/go-libp2p-transport-upgrader"
	yamux "github.com/libp2p/go-libp2p-yamux"
	basichost "github.com/libp2p/go-libp2p/p2p/host/basic"
	msmux "github.com/libp2p/go-stream-muxer-multistream"
	"github.com/libp2p/go-tcp-transport"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"net"
)

// Error types
var (
	ErrPeerIsNotConnected = errors.New("peer is not connected")
)

// Node the node can be used as both the client and the server
type Node struct {
	synchronizing bool
	quitCh        chan bool
	netService    *gamcService
	config        *Config
	context       context.Context
	id            peer.ID
	networkKey    crypto.PrivKey
	network       network.Network
	host          *basichost.BasicHost
	streamManager *StreamManager
	routeTable    *RouteTable
}

// NewNode return new Node according to the config.
func NewNode(config *Config) (*Node, error) {
	// check Listen port.
	if err := checkPortAvailable(config.Listen); err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err":    err,
			"listen": config.Listen,
		}).Error("Failed to check port.")
		return nil, err
	}

	node := &Node{
		quitCh:        make(chan bool, 10),
		config:        config,
		context:       context.Background(),
		streamManager: NewStreamManager(config),
		synchronizing: false,
	}

	initP2PNetworkKey(config, node)
	initP2PRouteTable(config, node)

	if err := initP2PSwarmNetwork(config, node); err != nil {
		return nil, err
	}

	return node, nil
}

func initP2PNetworkKey(config *Config, node *Node) {
	// init p2p network key.
	networkKey, err := LoadNetworkKeyFromFileOrCreateNew(config.PrivateKeyPath)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err":        err,
			"NetworkKey": config.PrivateKeyPath,
		}).Warn("Failed to load network private key from file.")
	}

	node.networkKey = networkKey
	node.id, err = peer.IDFromPublicKey(networkKey.GetPublic())
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err":        err,
			"NetworkKey": config.PrivateKeyPath,
		}).Warn("Failed to generate ID from network key file.")
	}
}

func initP2PRouteTable(config *Config, node *Node) error {
	// init p2p route table.
	node.routeTable = NewRouteTable(config, node)
	return nil
}

func (node *Node) startHost() error {
	// add nat manager
	options := &basichost.HostOpts{}
	options.NATManager = basichost.NewNATManager
	host, err := basichost.NewHost(node.context, node.network, options)
	if err != nil {
		logging.CLog().WithFields(logrus.Fields{
			"err":            err,
			"listen address": node.config.Listen,
		}).Error("Failed to start node.")
		return err
	}

	host.SetStreamHandler(gamcProtocolID, node.onStreamConnected)
	node.host = host

	return nil
}

// GenUpgrader creates a new connection upgrader for use with this swarm.
func GenUpgrader(n *swarm.Swarm) *tptu.Upgrader {
	id := n.LocalPeer()
	pk := n.Peerstore().PrivKey(id)
	secMuxer := new(csms.SSMuxer)
	secMuxer.AddTransport(secio.ID, &secio.Transport{
		LocalID:    id,
		PrivateKey: pk,
	})

	stMuxer := msmux.NewBlankTransport()
	stMuxer.AddTransport(gamcProtocolID, yamux.DefaultTransport)

	return &tptu.Upgrader{
		Secure:  secMuxer,
		Muxer:   stMuxer,
		Filters: n.Filters,
	}

}

func initP2PSwarmNetwork(config *Config, node *Node) error {
	// init p2p multiaddr and swarm network.
	swarm := swarm.NewSwarm(
		node.context,
		node.id,
		node.routeTable.peerStore,
		metrics.NewBandwidthCounter(),
	)

	tcpTransport := tcp.NewTCPTransport(GenUpgrader(swarm))
	if err := swarm.AddTransport(tcpTransport); err != nil {
		panic(err)
	}

	for _, v := range node.config.Listen {
		tcpAddr, err := net.ResolveTCPAddr("tcp", v)
		if err != nil {
			logging.CLog().WithFields(logrus.Fields{
				"err":    err,
				"listen": v,
			}).Error("Failed to bind node socket.")
			return err
		}

		addr, err := multiaddr.NewMultiaddr(
			fmt.Sprintf(
				"/ip4/%s/tcp/%d",
				tcpAddr.IP,
				tcpAddr.Port,
			),
		)
		if err != nil {
			logging.CLog().WithFields(logrus.Fields{
				"err":    err,
				"listen": v,
			}).Error("Failed to bind node socket.")
			return err
		}
		swarm.Listen(addr)
	}
	swarm.Peerstore().AddAddrs(node.id, swarm.ListenAddresses(), peerstore.PermanentAddrTTL)
	node.network = swarm
	return nil
}

// Start host & route table discovery
func (node *Node) Start() error {
	logging.CLog().Info("Starting gamcService Node...")

	node.streamManager.Start()

	if err := node.startHost(); err != nil {
		return err
	}

	node.routeTable.Start()

	logging.CLog().WithFields(logrus.Fields{
		"id":                node.ID(),
		"listening address": node.host.Addrs(),
	}).Info("Started gamcService Node.")

	return nil
}

// Stop stop a node.
func (node *Node) Stop() {
	logging.CLog().WithFields(logrus.Fields{
		"id":                node.ID(),
		"listening address": node.host.Addrs(),
	}).Info("Stopping gamcService Node...")

	node.routeTable.Stop()
	node.stopHost()
	node.streamManager.Stop()
}

func (node *Node) stopHost() {
	node.network.Close()

	if node.host == nil {
		return
	}

	node.host.Close()
}

// BroadcastMessage broadcast message.
func (node *Node) BroadcastMessage(messageName string, data Serializable, priority int) {
	// node can not broadcast or relay message if it is in synchronizing.
	if node.synchronizing {
		return
	}

	node.streamManager.BroadcastMessage(messageName, data, priority)
}

// RelayMessage relay message.
func (node *Node) RelayMessage(messageName string, data Serializable, priority int) {
	// node can not broadcast or relay message if it is in synchronizing.
	if node.synchronizing {
		return
	}

	node.streamManager.RelayMessage(messageName, data, priority)
}

// SendMessageToPeer send message to a peer.
func (node *Node) SendMessageToPeer(messageName string, data []byte, priority int, peerID string) error {
	stream := node.streamManager.FindByPeerID(peerID)
	if stream == nil {
		logging.VLog().WithFields(logrus.Fields{
			"pid": peerID,
			"err": ErrPeerIsNotConnected,
		}).Debug("Failed to locate peer's stream")
		return ErrPeerIsNotConnected
	}

	return stream.SendMessage(messageName, data, priority)
}

// SetgamcService set netService
func (node *Node) SetgamcService(ns *gamcService) {
	node.netService = ns
}

// ID return node ID.
func (node *Node) ID() string {
	return node.id.Pretty()
}

func (node *Node) onStreamConnected(s network.Stream) {
	node.streamManager.Add(s, node)
}
