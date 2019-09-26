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
	"gamc.pro/gamcio/go-gamc/util/config"
	"gamc.pro/gamcio/go-gamc/util/logging"
	"github.com/sirupsen/logrus"
)

// gamcService service for gamc p2p network
type gamcService struct {
	node       *Node
	dispatcher *Dispatcher
}

// NewgamcService create netService
func NewgamcService(conf *config.Config) (*gamcService, error) {
	netcfg := GetNetConfig(conf)

	if netcfg == nil {
		logging.CLog().Fatal("Failed to find network config in config file")
		return nil, ErrConfigLackNetWork
	}

	node, err := NewNode(NewP2PConfig(conf))
	if err != nil {
		return nil, err
	}

	ns := &gamcService{
		node:       node,
		dispatcher: NewDispatcher(),
	}
	node.SetgamcService(ns)

	return ns, nil
}

// PutMessage put message to dispatcher.
func (ns *gamcService) PutMessage(msg Message) {
	ns.dispatcher.PutMessage(msg)
}

// Start start p2p manager.
func (ns *gamcService) Start() error {
	logging.CLog().Info("Starting gamcService...")

	// start dispatcher.
	ns.dispatcher.Start()

	// start node.
	if err := ns.node.Start(); err != nil {
		ns.dispatcher.Stop()
		logging.CLog().WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to start gamcService.")
		return err
	}

	logging.CLog().Info("Started gamcService.")
	return nil
}

// Stop stop p2p manager.
func (ns *gamcService) Stop() {
	logging.CLog().Info("Stopping gamcService...")

	ns.node.Stop()
	ns.dispatcher.Stop()
}

// Register register the subscribers.
func (ns *gamcService) Register(subscribers ...*Subscriber) {
	ns.dispatcher.Register(subscribers...)
}

// Deregister Deregister the subscribers.
func (ns *gamcService) Deregister(subscribers ...*Subscriber) {
	ns.dispatcher.Deregister(subscribers...)
}

// Broadcast message.
func (ns *gamcService) Broadcast(name string, msg Serializable, priority int) {
	ns.node.BroadcastMessage(name, msg, priority)
}

// Relay message.
func (ns *gamcService) Relay(name string, msg Serializable, priority int) {
	ns.node.RelayMessage(name, msg, priority)
}

// SendMessage send message to a peer.
func (ns *gamcService) SendMessage(msgName string, msg []byte, target string, priority int) error {
	return ns.node.SendMessageToPeer(msgName, msg, priority, target)
}

// SendMessageToPeers send message to peers.
func (ns *gamcService) SendMessageToPeers(messageName string, data []byte, priority int, filter PeerFilterAlgorithm) []string {
	return ns.node.streamManager.SendMessageToPeers(messageName, data, priority, filter)
}

// SendMessageToPeer send message to a peer.
func (ns *gamcService) SendMessageToPeer(messageName string, data []byte, priority int, peerID string) error {
	return ns.node.SendMessageToPeer(messageName, data, priority, peerID)
}

// ClosePeer close the stream to a peer.
func (ns *gamcService) ClosePeer(peerID string, reason error) {
	ns.node.streamManager.CloseStream(peerID, reason)
}

// Node return the peer node
func (ns *gamcService) Node() *Node {
	return ns.node
}
