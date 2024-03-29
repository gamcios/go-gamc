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
	netpb "gamc.pro/gamcio/go-gamc/network/pb"
	"gamc.pro/gamcio/go-gamc/util/logging"
	"errors"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

// Stream Message Type
const (
	ClientVersion  = "0.1.0"
	gamcProtocolID = "/gamc/1.0.0"
	HELLO          = "hello"
	OK             = "ok"
	BYE            = "bye"
	SYNCROUTE      = "syncroute"
	ROUTETABLE     = "routetable"
	RECVEDMSG      = "recvedmsg"
	CurrentVersion = 0x0
)

const (
	streamStatusInit = iota
	streamStatusHandshakeSucceed
	streamStatusClosed
)

// Stream Errors
var (
	ErrShouldCloseConnectionAndExitLoop = errors.New("should close connection and exit loop")
	ErrStreamIsNotConnected             = errors.New("stream is not connected")
)

// Stream define the structure of a stream in p2p network
type Stream struct {
	syncMutex                 sync.Mutex
	pid                       peer.ID
	addr                      ma.Multiaddr
	stream                    network.Stream
	node                      *Node
	handshakeSucceedCh        chan bool
	messageNotifChan          chan int
	highPriorityMessageChan   chan *gamcMessage
	normalPriorityMessageChan chan *gamcMessage
	lowPriorityMessageChan    chan *gamcMessage
	quitWriteCh               chan bool
	status                    int
	connectedAt               int64
	latestReadAt              int64
	latestWriteAt             int64
	msgCount                  map[string]int
	reservedFlag              []byte
}

// NewStream return a new Stream
func NewStream(stream network.Stream, node *Node) *Stream {
	return newStreamInstance(stream.Conn().RemotePeer(), stream.Conn().RemoteMultiaddr(), stream, node)
}

func newStreamInstance(pid peer.ID, addr ma.Multiaddr, stream network.Stream, node *Node) *Stream {
	return &Stream{
		pid:                       pid,
		addr:                      addr,
		stream:                    stream,
		node:                      node,
		handshakeSucceedCh:        make(chan bool, 1),
		messageNotifChan:          make(chan int, 6*1024),
		highPriorityMessageChan:   make(chan *gamcMessage, 2*1024),
		normalPriorityMessageChan: make(chan *gamcMessage, 2*1024),
		lowPriorityMessageChan:    make(chan *gamcMessage, 2*1024),
		quitWriteCh:               make(chan bool, 1),
		status:                    streamStatusInit,
		connectedAt:               time.Now().Unix(),
		latestReadAt:              0,
		latestWriteAt:             0,
		msgCount:                  make(map[string]int),
		reservedFlag:              DefaultReserved,
	}
}

// IsConnected return if the stream is connected
func (s *Stream) IsConnected() bool {
	return s.stream != nil
}

// Connect to the stream
func (s *Stream) Connect() error {
	logging.VLog().WithFields(logrus.Fields{
		"stream": s.String(),
	}).Debug("Connecting to peer.")

	// connect to host.
	stream, err := s.node.host.NewStream(
		s.node.context,
		s.pid,
		gamcProtocolID,
	)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"stream": s.String(),
			"err":    err,
		}).Debug("Failed to connect to host.")
		return err
	}
	s.stream = stream
	s.addr = stream.Conn().RemoteMultiaddr()

	return nil
}

func (s *Stream) handleMessage(message *gamcMessage) error {
	messageName := message.MessageName()
	s.msgCount[messageName]++

	switch messageName {
	case HELLO:
		return s.onHello(message)
	case OK:
		return s.onOk(message)
	case BYE:
		return s.onBye(message)
	}

	// check handshake status.
	if s.status != streamStatusHandshakeSucceed {
		return ErrShouldCloseConnectionAndExitLoop
	}

	switch messageName {
	case SYNCROUTE:
		return s.onSyncRoute(message)
	case ROUTETABLE:
		return s.onRouteTable(message)
	default:
		data, err := s.getData(message)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err":         err,
				"messageName": message.MessageName(),
			}).Info("Handle message data occurs error.")
			return err
		}
		s.node.netService.PutMessage(NewBaseMessage(message.MessageName(), s.pid.Pretty(), data))
		// record recv message.
		RecordRecvMessage(s, message.DataCheckSum())
	}

	return nil
}

// Close close the stream
func (s *Stream) close(reason error) {
	// Add lock & close flag to prevent multi call.
	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()

	if s.status == streamStatusClosed {
		return
	}
	s.status = streamStatusClosed

	logging.VLog().WithFields(logrus.Fields{
		"stream": s.String(),
		"reason": reason,
	}).Debug("Closing stream.")

	// cleanup.
	s.node.streamManager.RemoveStream(s)
	s.node.routeTable.RemovePeerStream(s)

	// quit.
	s.quitWriteCh <- true

	// close stream.
	if s.stream != nil {
		s.stream.Close()
	}
}

func (s *Stream) Hello() error {
	msg := &netpb.Hello{
		NodeId:        s.node.id.String(),
		ClientVersion: ClientVersion,
	}
	return s.WriteProtoMessage(HELLO, msg, ReservedCompressionClientFlag)
}

func (s *Stream) onHello(message *gamcMessage) error {
	msg, err := netpb.HelloMessageFromProto(message.OriginalData())
	if err != nil {
		return ErrShouldCloseConnectionAndExitLoop
	}

	if msg.NodeId != s.pid.String() || !CheckClientVersionCompatibility(ClientVersion, msg.ClientVersion) {
		// invalid client, bye().
		logging.VLog().WithFields(logrus.Fields{
			"pid":               s.pid.Pretty(),
			"address":           s.addr,
			"ok.node_id":        msg.NodeId,
			"ok.client_version": msg.ClientVersion,
		}).Warn("Invalid NodeId or incompatible client version.")
		return ErrShouldCloseConnectionAndExitLoop
	}

	if (message.Reserved()[2] & ReservedCompressionClientFlag) > 0 {
		s.reservedFlag = CompressionReserved
	}

	// add to route table.
	s.node.routeTable.AddPeerStream(s)

	// handshake finished.
	s.finishHandshake()

	return s.Ok()
}

// Ok say ok in the stream
func (s *Stream) Ok() error {
	// send OK.
	resp := &netpb.OK{
		NodeId:        s.node.id.String(),
		ClientVersion: ClientVersion,
	}

	return s.WriteProtoMessage(OK, resp, ReservedCompressionClientFlag)
}

func (s *Stream) onOk(message *gamcMessage) error {
	msg, err := netpb.OKMessageFromProto(message.OriginalData())
	if err != nil {
		return ErrShouldCloseConnectionAndExitLoop
	}

	if msg.NodeId != s.pid.String() || !CheckClientVersionCompatibility(ClientVersion, msg.ClientVersion) {
		// invalid client, bye().
		logging.VLog().WithFields(logrus.Fields{
			"pid":               s.pid.Pretty(),
			"address":           s.addr,
			"ok.node_id":        msg.NodeId,
			"ok.client_version": msg.ClientVersion,
		}).Warn("Invalid NodeId or incompatible client version.")
		return ErrShouldCloseConnectionAndExitLoop
	}

	if (message.Reserved()[2] & ReservedCompressionClientFlag) > 0 {
		s.reservedFlag = CompressionReserved
	}

	// add to route table.
	s.node.routeTable.AddPeerStream(s)

	// handshake finished.
	s.finishHandshake()

	return nil
}

// Bye say bye in the stream
func (s *Stream) Bye() {
	s.WriteMessage(BYE, []byte{}, DefaultReservedFlag)
	s.close(errors.New("bye: force close"))
}

func (s *Stream) onBye(message *gamcMessage) error {
	logging.VLog().WithFields(logrus.Fields{
		"stream": s.String(),
	}).Debug("Received Bye message, close the connection.")
	return ErrShouldCloseConnectionAndExitLoop
}

// SyncRoute send sync route request
func (s *Stream) SyncRoute() error {
	return s.SendMessage(SYNCROUTE, []byte{}, MessagePriorityHigh)
}

func (s *Stream) onSyncRoute(message *gamcMessage) error {
	return s.RouteTable()
}

// RouteTable send sync table request
func (s *Stream) RouteTable() error {
	// get random peers from routeTable
	peers := s.node.routeTable.GetRandomPeers(s.pid)

	// prepare the protobuf message.
	msg := &netpb.Peers{
		Peers: make([]*netpb.PeerInfo, len(peers)),
	}

	for i, v := range peers {
		pi := &netpb.PeerInfo{
			Id:    v.ID.Pretty(),
			Addrs: make([]string, len(v.Addrs)),
		}
		for j, addr := range v.Addrs {
			pi.Addrs[j] = addr.String()
		}
		msg.Peers[i] = pi
	}

	logging.VLog().WithFields(logrus.Fields{
		"stream":          s.String(),
		"routetableCount": len(peers),
	}).Debug("Replied sync route message.")

	return s.SendProtoMessage(ROUTETABLE, msg, MessagePriorityHigh)
}

func (s *Stream) onRouteTable(message *gamcMessage) error {
	data, err := s.getData(message)
	if err != nil {
		return err
	}

	peers := new(netpb.Peers)
	if err := proto.Unmarshal(data, peers); err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err": err,
		}).Debug("Invalid Peers proto message.")
		return ErrShouldCloseConnectionAndExitLoop
	}

	s.node.routeTable.AddPeers(s.node.ID(), peers)

	return nil
}

func (s *Stream) finishHandshake() {
	logging.VLog().WithFields(logrus.Fields{
		"stream": s.String(),
	}).Debug("Finished handshake.")

	s.status = streamStatusHandshakeSucceed
	s.handshakeSucceedCh <- true
}

// SendProtoMessage send proto msg to buffer
func (s *Stream) SendProtoMessage(messageName string, pb proto.Message, priority int) error {
	data, err := proto.Marshal(pb)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":         err,
			"messageName": messageName,
			"stream":      s.String(),
		}).Debug("Failed to marshal proto message.")
		return err
	}

	return s.SendMessage(messageName, data, priority)
}

// WriteProtoMessage write proto msg in the stream
func (s *Stream) WriteProtoMessage(messageName string, pb proto.Message, reservedClientFlag byte) error {
	data, err := proto.Marshal(pb)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":         err,
			"messageName": messageName,
			"stream":      s.String(),
		}).Debug("Failed to marshal proto message.")
		return err
	}

	return s.WriteMessage(messageName, data, reservedClientFlag)
}

// WriteMessage write raw msg in the stream
func (s *Stream) WriteMessage(messageName string, data []byte, reservedClientFlag byte) error {
	// hello and ok messages come with the client flag bit.
	var reserved = make([]byte, len(s.reservedFlag))
	copy(reserved, s.reservedFlag)

	if reservedClientFlag == ReservedCompressionClientFlag {
		reserved[2] = s.reservedFlag[2] | reservedClientFlag
	}

	message, err := NewgamcMessage(s.node.config.ChainID, reserved, CurrentVersion, messageName, data)
	if err != nil {
		return err
	}

	return s.WritegamcMessage(message)
}

// WritegamcMessage write gamc msg in the stream
func (s *Stream) WritegamcMessage(message *gamcMessage) error {
	// metrics.
	metricsPacketsOutByMessageName(message.MessageName(), message.Length())

	err := s.Write(message.Content())
	message.FlagWriteMessageAt()

	return err
}

// SendMessage send msg to buffer
func (s *Stream) SendMessage(messageName string, data []byte, priority int) error {
	message, err := NewgamcMessage(s.node.config.ChainID, s.reservedFlag, CurrentVersion, messageName, data)
	if err != nil {
		return err
	}

	// metrics.
	metricsPacketsOutByMessageName(messageName, message.Length())

	// send to pool.
	message.FlagSendMessageAt()

	// use a non-blocking channel to avoid blocking when the channel is full.
	switch priority {
	case MessagePriorityHigh:
		s.highPriorityMessageChan <- message
	case MessagePriorityNormal:
		select {
		case s.normalPriorityMessageChan <- message:
		default:
			logging.VLog().WithFields(logrus.Fields{
				"normalPriorityMessageChan.len": len(s.normalPriorityMessageChan),
				"stream":                        s.String(),
			}).Debug("Received too many normal priority message.")
			return nil
		}
	default:
		select {
		case s.lowPriorityMessageChan <- message:
		default:
			logging.VLog().WithFields(logrus.Fields{
				"lowPriorityMessageChan.len": len(s.lowPriorityMessageChan),
				"stream":                     s.String(),
			}).Debug("Received too many low priority message.")
			return nil
		}
	}
	select {
	case s.messageNotifChan <- 1:
	default:
		logging.VLog().WithFields(logrus.Fields{
			"messageNotifChan.len": len(s.messageNotifChan),
			"stream":               s.String(),
		}).Debug("Received too many message notifChan.")
		return nil
	}
	return nil
}

func (s *Stream) Write(data []byte) error {
	if s.stream == nil {
		s.close(ErrStreamIsNotConnected)
		return ErrStreamIsNotConnected
	}

	// at least 5kb/s to write message
	deadline := time.Now().Add(time.Duration(len(data)/1024/5+1) * time.Second)
	if err := s.stream.SetWriteDeadline(deadline); err != nil {
		return err
	}
	n, err := s.stream.Write(data)
	if err != nil {
		logging.VLog().WithFields(logrus.Fields{
			"err":    err,
			"stream": s.String(),
		}).Warn("Failed to send message to peer.")
		s.close(err)
		return err
	}
	s.latestWriteAt = time.Now().Unix()

	// metrics.
	metricsPacketsOut.Mark(1)
	metricsBytesOut.Mark(int64(n))

	return nil
}

// StartLoop start stream handling loop.
func (s *Stream) StartLoop() {
	go s.writeLoop()
	go s.readLoop()
}

func (s *Stream) readLoop() {
	// send Hello to host if stream is not connected.
	if !s.IsConnected() {
		if err := s.Connect(); err != nil {
			s.close(err)
			return
		}
		if err := s.Hello(); err != nil {
			s.close(err)
			return
		}
	}

	// loop.
	buf := make([]byte, 1024*4)
	messageBuffer := make([]byte, 0)

	var message *gamcMessage

	for {
		n, err := s.stream.Read(buf)
		if err != nil {
			logging.VLog().WithFields(logrus.Fields{
				"err":    err,
				"stream": s.String(),
			}).Debug("Error occurred when reading data from network connection.")
			s.close(err)
			return
		}

		messageBuffer = append(messageBuffer, buf[:n]...)
		s.latestReadAt = time.Now().Unix()

		for {
			if message == nil {
				var err error

				// waiting for header data.
				if len(messageBuffer) < gamcMessageHeaderLength {
					// continue reading.
					break
				}

				message, err = ParsegamcMessage(messageBuffer)
				if err != nil {
					s.Bye()
					return
				}

				// check ChainID.
				if s.node.config.ChainID != message.ChainID() {
					logging.VLog().WithFields(logrus.Fields{
						"err":             err,
						"stream":          s.String(),
						"conf.chainID":    s.node.config.ChainID,
						"message.chainID": message.ChainID(),
					}).Warn("Invalid chainID, disconnect the connection.")
					s.Bye()
					return
				}

				// remove header from buffer.
				messageBuffer = messageBuffer[gamcMessageHeaderLength:]
			}

			// waiting for data.
			if len(messageBuffer) < int(message.DataLength()) {
				// continue reading.
				break
			}

			if err := message.ParseMessageData(messageBuffer); err != nil {
				s.Bye()
				return
			}

			// remove data from buffer.
			messageBuffer = messageBuffer[message.DataLength():]

			// metrics.
			metricsPacketsIn.Mark(1)
			metricsBytesIn.Mark(int64(message.Length()))
			metricsPacketsInByMessageName(message.MessageName(), message.Length())

			// handle message.
			if err := s.handleMessage(message); err == ErrShouldCloseConnectionAndExitLoop {
				s.Bye()
				return
			}

			// reset message.
			message = nil
		}
	}
}

func (s *Stream) writeLoop() {
	// waiting for handshake succeed.
	handshakeTimeoutTicker := time.NewTicker(30 * time.Second)
	select {
	case <-s.handshakeSucceedCh:
		// handshake succeed.
	case <-s.quitWriteCh:
		logging.VLog().WithFields(logrus.Fields{
			"stream": s.String(),
		}).Debug("Quiting Stream Write Loop.")
		return
	case <-handshakeTimeoutTicker.C:
		logging.VLog().WithFields(logrus.Fields{
			"stream": s.String(),
		}).Debug("Handshaking Stream timeout, quiting.")
		s.close(errors.New("Handshake timeout"))
		return
	}

	for {
		select {
		case <-s.quitWriteCh:
			logging.VLog().WithFields(logrus.Fields{
				"stream": s.String(),
			}).Debug("Quiting Stream Write Loop.")
			return
		case <-s.messageNotifChan:
			select {
			case message := <-s.highPriorityMessageChan:
				s.WritegamcMessage(message)
				continue
			default:
			}

			select {
			case message := <-s.normalPriorityMessageChan:
				s.WritegamcMessage(message)
				continue
			default:
			}

			select {
			case message := <-s.lowPriorityMessageChan:
				s.WritegamcMessage(message)
				continue
			default:
			}
		}
	}
}

// CheckClientVersionCompatibility if two clients are compatible
// If the clientVersion of node A is X.Y.Z, then node B must be X.Y.{} to be compatible with A.
func CheckClientVersionCompatibility(v1, v2 string) bool {
	s1 := strings.Split(v1, ".")
	s2 := strings.Split(v1, ".")

	if len(s1) != 3 || len(s2) != 3 {
		return false
	}

	if s1[0] != s2[0] || s1[1] != s2[1] {
		return false
	}
	return true
}

func (s *Stream) getData(message *gamcMessage) ([]byte, error) {
	var data []byte
	if ByteSliceEqualBCE(s.reservedFlag, CompressionReserved) {
		var err error
		data, err = message.Data()
		if err != nil {
			return nil, err
		}
	} else {
		data = message.OriginalData()
	}
	return data, nil
}

// ByteSliceEqualBCE determines whether two byte arrays are equal.
func ByteSliceEqualBCE(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) {
		return false
	}

	b = b[:len(a)]
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

// IsHandshakeSucceed return if the handshake in the stream succeed
func (s *Stream) IsHandshakeSucceed() bool {
	return s.status == streamStatusHandshakeSucceed
}

// NewStreamFromPID return a new Stream based on the pid
func NewStreamFromPID(pid peer.ID, node *Node) *Stream {
	return newStreamInstance(pid, nil, nil, node)
}

func (s *Stream) String() string {
	addrStr := ""
	if s.addr != nil {
		addrStr = s.addr.String()
	}

	return fmt.Sprintf("Peer Stream: %s,%s", s.pid.Pretty(), addrStr)
}
