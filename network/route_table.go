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
	"bufio"
	netpb "gamc.pro/gamcio/go-gamc/network/pb"
	"gamc.pro/gamcio/go-gamc/util/logging"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	kbucket "github.com/libp2p/go-libp2p-kbucket"
	pstoremem "github.com/libp2p/go-libp2p-peerstore/pstoremem"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"path"
	"reflect"
	"strings"
	"time"
)

// Route Table Errors
var (
	ErrExceedMaxSyncRouteResponse = errors.New("too many sync route table response")
)

// RouteTable route table struct.
type RouteTable struct {
	quitCh                   chan bool
	peerStore                peerstore.Peerstore
	routeTable               *kbucket.RoutingTable
	maxPeersCountForSyncResp int
	maxPeersCountToSync      int
	cacheFilePath            string
	seedNodes                []ma.Multiaddr
	node                     *Node
	streamManager            *StreamManager
	latestUpdatedAt          int64
	internalNodeList         []string
}

// NewRouteTable new route table.
func NewRouteTable(config *Config, node *Node) *RouteTable {
	table := &RouteTable{
		quitCh:                   make(chan bool, 1),
		peerStore:                pstoremem.NewPeerstore(),
		maxPeersCountForSyncResp: MaxPeersCountForSyncResp,
		maxPeersCountToSync:      config.MaxSyncNodes,
		cacheFilePath:            path.Join(config.RoutingTableDir, RouteTableCacheFileName),
		seedNodes:                config.BootNodes,
		node:                     node,
		streamManager:            node.streamManager,
		latestUpdatedAt:          0,
	}

	table.routeTable = kbucket.NewRoutingTable(
		config.Bucketsize,
		kbucket.ConvertPeerID(node.id),
		config.Latency,
		table.peerStore,
	)

	table.routeTable.Update(node.id)
	table.peerStore.AddPubKey(node.id, node.networkKey.GetPublic())
	table.peerStore.AddPrivKey(node.id, node.networkKey)

	return table
}

// RemovePeerStream remove peerStream from peerStore.
func (table *RouteTable) RemovePeerStream(s *Stream) {
	table.peerStore.AddAddr(s.pid, s.addr, 0)
	table.routeTable.Remove(s.pid)
	table.onRouteTableChange()
}

func (table *RouteTable) onRouteTableChange() {
	table.latestUpdatedAt = time.Now().Unix()
}

// AddPeerStream add peer stream to peerStore.
func (table *RouteTable) AddPeerStream(s *Stream) {
	table.peerStore.AddAddr(
		s.pid,
		s.addr,
		peerstore.PermanentAddrTTL,
	)
	table.routeTable.Update(s.pid)
	table.onRouteTableChange()
}

// GetRandomPeers get random peers
func (table *RouteTable) GetRandomPeers(pid peer.ID) []peer.AddrInfo {

	// change sync route algorithm from `NearestPeers` to `randomPeers`
	var peers []peer.ID
	allPeers := table.routeTable.ListPeers()
	// Do not accept internal node synchronization routing requests.
	if inArray(pid.Pretty(), table.internalNodeList) {
		return []peer.AddrInfo{}
	}

	for _, v := range allPeers {
		if inArray(v.Pretty(), table.internalNodeList) == false {
			peers = append(peers, v)
		}
	}
	peers = shufflePeerID(peers)
	if len(peers) > table.maxPeersCountForSyncResp {
		peers = peers[:table.maxPeersCountForSyncResp]
	}
	ret := make([]peer.AddrInfo, len(peers))
	for i, v := range peers {
		ret[i] = table.peerStore.PeerInfo(v)
	}
	return ret
}

// AddPeerInfo add peer to route table.
func (table *RouteTable) AddPeerInfo(prettyID string, addrStr []string) error {
	pid, err := peer.IDB58Decode(prettyID)
	if err != nil {
		return nil
	}

	addrs := make([]ma.Multiaddr, len(addrStr))
	for i, v := range addrStr {
		addrs[i], err = ma.NewMultiaddr(v)
		if err != nil {
			return err
		}
	}

	if table.routeTable.Find(pid) != "" {
		table.peerStore.SetAddrs(pid, addrs, peerstore.PermanentAddrTTL)
	} else {
		table.peerStore.AddAddrs(pid, addrs, peerstore.PermanentAddrTTL)
	}
	table.routeTable.Update(pid)
	table.onRouteTableChange()

	return nil
}

// AddPeers add peers to route table
func (table *RouteTable) AddPeers(pid string, peers *netpb.Peers) {
	// recv too many peers info. say Bye.
	if len(peers.Peers) > table.maxPeersCountForSyncResp {
		table.streamManager.CloseStream(pid, ErrExceedMaxSyncRouteResponse)
	}
	for _, v := range peers.Peers {
		table.AddPeerInfo(v.Id, v.Addrs)
	}
}

// Start start route table syncLoop.
func (table *RouteTable) Start() {
	logging.CLog().Info("Starting gamcService RouteTable Sync...")

	go table.syncLoop()
}

func (table *RouteTable) syncLoop() {
	// Load Route Table.
	table.LoadSeedNodes()
	table.LoadRouteTableFromFile()
	table.LoadInternalNodeList()

	// trigger first sync.
	table.SyncRouteTable()

	logging.CLog().Info("Started gamcService RouteTable Sync.")

	syncLoopTicker := time.NewTicker(RouteTableSyncLoopInterval)
	saveRouteTableToDiskTicker := time.NewTicker(RouteTableSaveToDiskInterval)
	latestUpdatedAt := table.latestUpdatedAt

	for {
		select {
		case <-table.quitCh:
			logging.CLog().Info("Stopped gamcService RouteTable Sync.")
			return
		case <-syncLoopTicker.C:
			table.SyncRouteTable()
		case <-saveRouteTableToDiskTicker.C:
			if latestUpdatedAt < table.latestUpdatedAt {
				table.SaveRouteTableToFile()
				latestUpdatedAt = table.latestUpdatedAt
			}
		}
	}
}

// LoadSeedNodes load seed nodes.
func (table *RouteTable) LoadSeedNodes() {
	for _, ipfsAddr := range table.seedNodes {
		table.AddIPFSPeerAddr(ipfsAddr)
	}
}

// AddIPFSPeerAddr add a peer to route table with ipfs address.
func (table *RouteTable) AddIPFSPeerAddr(addr ma.Multiaddr) {
	id, addr, err := ParseFromIPFSAddr(addr)
	if err != nil {
		return
	}
	table.AddPeer(id, addr)
}

// AddPeer add peer to route table.
func (table *RouteTable) AddPeer(pid peer.ID, addr ma.Multiaddr) {
	logging.VLog().Debugf("Adding Peer: %s,%s", pid.Pretty(), addr.String())
	table.peerStore.AddAddr(pid, addr, peerstore.PermanentAddrTTL)
	table.routeTable.Update(pid)
	table.onRouteTableChange()

}

// LoadRouteTableFromFile load route table from file.
func (table *RouteTable) LoadRouteTableFromFile() {
	file, err := os.Open(table.cacheFilePath)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"cacheFilePath": table.cacheFilePath,
			"err":           err,
		}).Warn("Failed to open Route Table Cache file.")
		return
	}
	defer file.Close()

	// read line by line.
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}

		addr, err := ma.NewMultiaddr(line)
		if err != nil {
			// ignore.
			logging.VLog().WithFields(logrus.Fields{
				"err":  err,
				"text": line,
			}).Warn("Invalid address in Route Table Cache file.")
			continue
		}

		table.AddIPFSPeerAddr(addr)
	}
}

//LoadInternalNodeList Load Internal Node list from file
func (table *RouteTable) LoadInternalNodeList() {
	file, err := os.Open(RouteTableInternalNodeFileName)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Warn("Failed to open internal list file.")
		return
	}
	defer file.Close()

	// read line by line.
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 0 {
			table.internalNodeList = append(table.internalNodeList, line)
		}
	}

	logging.VLog().WithFields(logrus.Fields{
		"internalNodeList": table.internalNodeList,
	}).Info("Loaded internal node list.")
}

// SyncRouteTable sync route table.
func (table *RouteTable) SyncRouteTable() {
	syncedPeers := make(map[peer.ID]bool)

	// sync with seed nodes.
	for _, ipfsAddr := range table.seedNodes {
		pid, _, err := ParseFromIPFSAddr(ipfsAddr)
		if err != nil {
			continue
		}
		table.SyncWithPeer(pid)
		syncedPeers[pid] = true
	}

	// random peer selection.
	peers := table.routeTable.ListPeers()
	peersCount := len(peers)
	if peersCount <= 1 {
		return
	}

	peersCountToSync := table.maxPeersCountToSync

	if peersCount < peersCountToSync {
		peersCountToSync = peersCount
	}
	selectedPeersIdx := make(map[int]bool)
	for i := 0; i < peersCountToSync/2; i++ {
		ri := 0

		for {
			ri = rand.Intn(peersCountToSync)
			if selectedPeersIdx[ri] == false {
				break
			}
		}

		selectedPeersIdx[ri] = true
		pid := peers[ri]

		if syncedPeers[pid] == false {
			table.SyncWithPeer(pid)
			syncedPeers[pid] = true
		}
	}
}

// SyncWithPeer sync route table with a peer.
func (table *RouteTable) SyncWithPeer(pid peer.ID) {
	if pid == table.node.id {
		return
	}

	stream := table.streamManager.Find(pid)

	if stream == nil {
		stream = NewStreamFromPID(pid, table.node)
		table.streamManager.AddStream(stream)
	}

	stream.SyncRoute()
}

// SaveRouteTableToFile save route table to file.
func (table *RouteTable) SaveRouteTableToFile() {
	file, err := os.Create(table.cacheFilePath)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"cacheFilePath": table.cacheFilePath,
			"err":           err,
		}).Warn("Failed to open Route Table Cache file.")
		return
	}
	defer file.Close()

	// write header.
	file.WriteString(fmt.Sprintf("# %s\n", time.Now().String()))

	peers := table.routeTable.ListPeers()
	for _, v := range peers {
		for _, addr := range table.peerStore.Addrs(v) {
			line := fmt.Sprintf("%s/ipfs/%s\n", addr, v.Pretty())
			file.WriteString(line)
		}
	}
}

// Stop quit route table syncLoop.
func (table *RouteTable) Stop() {
	logging.CLog().Info("Stopping gamcService RouteTable Sync...")

	table.quitCh <- true
}

func inArray(obj interface{}, array interface{}) bool {
	arrayValue := reflect.ValueOf(array)
	if reflect.TypeOf(array).Kind() == reflect.Array || reflect.TypeOf(array).Kind() == reflect.Slice {
		for i := 0; i < arrayValue.Len(); i++ {
			if arrayValue.Index(i).Interface() == obj {
				return true
			}
		}
	}
	return false
}

func shufflePeerID(pids []peer.ID) []peer.ID {

	r := rand.New(rand.NewSource(time.Now().Unix()))
	ret := make([]peer.ID, len(pids))
	perm := r.Perm(len(pids))
	for i, randIndex := range perm {
		ret[i] = pids[randIndex]
	}
	return ret
}
