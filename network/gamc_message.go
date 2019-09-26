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
	"bytes"
	"gamc.pro/gamcio/go-gamc/util/byteutils"
	"gamc.pro/gamcio/go-gamc/util/logging"
	"errors"
	"github.com/golang/snappy"
	"github.com/sirupsen/logrus"
	"hash/crc32"
	"time"
)

/*
gamcMessage defines protocol in gamc, we define our own wire protocol, as the following:

 0               1               2               3              (bytes)
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                         Magic Number                          |
+-----------------------------------------------+---------------+
|                         Chain ID              |    Reserved   |
+-------------------------------+---------------+---------------+
|           Reserved            |            Version            |
+-------------------------------+-------------------------------+
|                                                               |
+                                                               +
|                         Message Name                          |
+                                                               +
|                                                               |
+---------------------------------------------------------------+
|                         Data Length                           |
+---------------------------------------------------------------+
|                         Data Checksum                         |
+---------------------------------------------------------------+
|                         Header Checksum                       |
|---------------------------------------------------------------+
|                                                               |
+                         Data                                  +
.                                                               .
|                                                               |
+---------------------------------------------------------------+
*/

const (
	gamcMessageMagicNumberEndIdx    = 4
	gamcMessageChainIDEndIdx        = 7
	gamcMessageReservedEndIdx       = 10
	gamcMessageVersionIndex         = 10
	gamcMessageVersionEndIdx        = 12
	gamcMessageNameEndIdx           = 24
	gamcMessageDataLengthEndIdx     = 28
	gamcMessageDataCheckSumEndIdx   = 32
	gamcMessageHeaderCheckSumEndIdx = 36
	gamcMessageHeaderLength         = 36

	// Consider that a block is too large in sync.
	MaxgamcMessageDataLength = 512 * 1024 * 1024 // 512m.
	MaxgamcMessageNameLength = 24 - 12           // 12.

	DefaultReservedFlag           = 0x0
	ReservedCompressionEnableFlag = 0x80
	ReservedCompressionClientFlag = 0x40
)

var (
	MagicNumber         = []byte{0x47, 0x41, 0x4D, 0x43}
	DefaultReserved     = []byte{DefaultReservedFlag, DefaultReservedFlag, DefaultReservedFlag}
	CompressionReserved = []byte{DefaultReservedFlag, DefaultReservedFlag, DefaultReservedFlag | ReservedCompressionEnableFlag}

	ErrInsufficientMessageHeaderLength = errors.New("insufficient message header length")
	ErrInsufficientMessageDataLength   = errors.New("insufficient message data length")
	ErrInvalidMagicNumber              = errors.New("invalid magic number")
	ErrInvalidHeaderCheckSum           = errors.New("invalid header checksum")
	ErrInvalidDataCheckSum             = errors.New("invalid data checksum")
	ErrExceedMaxDataLength             = errors.New("exceed max data length")
	ErrExceedMaxMessageNameLength      = errors.New("exceed max message name length")
	ErrUncompressMessageFailed         = errors.New("uncompress message failed")
)

//gamcMessage struct
type gamcMessage struct {
	content     []byte
	messageName string

	// debug fields.
	sendMessageAt  int64
	writeMessageAt int64
}

// MagicNumber return magicNumber
func (message *gamcMessage) MagicNumber() []byte {
	return message.content[0:gamcMessageMagicNumberEndIdx]
}

// ChainID return chainID
func (message *gamcMessage) ChainID() uint32 {
	chainIdData := make([]byte, 4)
	copy(chainIdData[1:], message.content[gamcMessageMagicNumberEndIdx:gamcMessageChainIDEndIdx])
	return byteutils.Uint32(chainIdData)
}

// Reserved return reserved
func (message *gamcMessage) Reserved() []byte {
	return message.content[gamcMessageChainIDEndIdx:gamcMessageReservedEndIdx]
}

// Version return version
func (message *gamcMessage) Version() uint16 {
	return byteutils.Uint16(message.content[gamcMessageVersionIndex:gamcMessageVersionEndIdx])
}

// MessageName return message name
func (message *gamcMessage) MessageName() string {
	if message.messageName == "" {
		data := message.content[gamcMessageVersionEndIdx:gamcMessageNameEndIdx]
		pos := bytes.IndexByte(data, 0)
		if pos != -1 {
			message.messageName = string(data[0:pos])
		} else {
			message.messageName = string(data)
		}
	}
	return message.messageName
}

// DataLength return dataLength
func (message *gamcMessage) DataLength() uint32 {
	return byteutils.Uint32(message.content[gamcMessageNameEndIdx:gamcMessageDataLengthEndIdx])
}

// DataCheckSum return data checkSum
func (message *gamcMessage) DataCheckSum() uint32 {
	return byteutils.Uint32(message.content[gamcMessageDataLengthEndIdx:gamcMessageDataCheckSumEndIdx])
}

// HeaderCheckSum return header checkSum
func (message *gamcMessage) HeaderCheckSum() uint32 {
	return byteutils.Uint32(message.content[gamcMessageDataCheckSumEndIdx:gamcMessageHeaderCheckSumEndIdx])
}

// HeaderWithoutCheckSum return header without checkSum
func (message *gamcMessage) HeaderWithoutCheckSum() []byte {
	return message.content[:gamcMessageDataCheckSumEndIdx]
}

// Data return data
func (message *gamcMessage) Data() ([]byte, error) {
	reserved := message.Reserved()
	data := message.content[gamcMessageHeaderLength:]
	if (reserved[2] & ReservedCompressionEnableFlag) > 0 {
		var err error
		data, err = snappy.Decode(nil, data)
		//dstData := make([]byte, MaxgamcMessageDataLength)
		//l, err := lz4.UncompressBlock(data, dstData)
		if err != nil {
			return nil, ErrUncompressMessageFailed
		}
		//if l > 0 {
		//	data = make([]byte, l)
		//	data = dstData[:l]
		//}
	}
	return data, nil
}

// OriginalData return original data
func (message *gamcMessage) OriginalData() []byte {
	return message.content[gamcMessageHeaderLength:]
}

// Content return message content
func (message *gamcMessage) Content() []byte {
	return message.content
}

// Length return message Length
func (message *gamcMessage) Length() uint64 {
	return uint64(len(message.content))
}

// NewgamcMessage new gamc message
func NewgamcMessage(chainID uint32, reserved []byte, version uint16, messageName string, data []byte) (*gamcMessage, error) {

	// Process message compression
	if ((reserved[2] & ReservedCompressionClientFlag) == 0) && ((reserved[2] & ReservedCompressionEnableFlag) > 0) {
		data = snappy.Encode(nil, data)
		//dstData := make([]byte, len(data))
		//ht := make([]int, 64<<10)
		//l, err := lz4.CompressBlock(data, dstData, ht)
		//if err != nil {
		//	panic(err)
		//}
		//if l > 0 {
		//	data = make([]byte, l)
		//	data = dstData[:l]
		//}
	}

	if len(data) > MaxgamcMessageDataLength {
		logging.VLog().WithFields(logrus.Fields{
			"messageName": messageName,
			"dataLength":  len(data),
			"limits":      MaxgamcMessageDataLength,
		}).Debug("Exceeded max data length.")
		return nil, ErrExceedMaxDataLength
	}

	if len(messageName) > MaxgamcMessageNameLength {
		logging.VLog().WithFields(logrus.Fields{
			"messageName":      messageName,
			"len(messageName)": len(messageName),
			"limits":           MaxgamcMessageNameLength,
		}).Debug("Exceeded max message name length.")
		return nil, ErrExceedMaxMessageNameLength
	}

	dataCheckSum := crc32.ChecksumIEEE(data)

	message := &gamcMessage{
		content: make([]byte, gamcMessageHeaderLength+len(data)),
	}

	// copy fields.
	copy(message.content[0:gamcMessageMagicNumberEndIdx], MagicNumber)
	chainIdData := byteutils.FromUint32(chainID)
	copy(message.content[gamcMessageMagicNumberEndIdx:gamcMessageChainIDEndIdx], chainIdData[1:])
	copy(message.content[gamcMessageChainIDEndIdx:gamcMessageReservedEndIdx], reserved)
	copy(message.content[gamcMessageVersionIndex:gamcMessageVersionEndIdx], byteutils.FromUint16(version))
	copy(message.content[gamcMessageVersionEndIdx:gamcMessageNameEndIdx], []byte(messageName))
	copy(message.content[gamcMessageNameEndIdx:gamcMessageDataLengthEndIdx], byteutils.FromUint32(uint32(len(data))))
	copy(message.content[gamcMessageDataLengthEndIdx:gamcMessageDataCheckSumEndIdx], byteutils.FromUint32(dataCheckSum))

	// header checksum.
	headerCheckSum := crc32.ChecksumIEEE(message.HeaderWithoutCheckSum())
	copy(message.content[gamcMessageDataCheckSumEndIdx:gamcMessageHeaderCheckSumEndIdx], byteutils.FromUint32(headerCheckSum))

	// copy data.
	copy(message.content[gamcMessageHeaderCheckSumEndIdx:], data)

	return message, nil
}

// ParsegamcMessage parse gamc message
func ParsegamcMessage(data []byte) (*gamcMessage, error) {
	if len(data) < gamcMessageHeaderLength {
		return nil, ErrInsufficientMessageHeaderLength
	}

	message := &gamcMessage{
		content: make([]byte, gamcMessageHeaderLength),
	}
	copy(message.content, data)

	if err := message.VerifyHeader(); err != nil {
		return nil, err
	}

	return message, nil
}

// ParseMessageData parse gamc message data
func (message *gamcMessage) ParseMessageData(data []byte) error {
	if uint32(len(data)) < message.DataLength() {
		return ErrInsufficientMessageDataLength
	}

	message.content = append(message.content, data[:message.DataLength()]...)
	return message.VerifyData()
}

// VerifyHeader verify message header
func (message *gamcMessage) VerifyHeader() error {
	if !byteutils.Equal(MagicNumber, message.MagicNumber()) {
		logging.VLog().WithFields(logrus.Fields{
			"expect": MagicNumber,
			"actual": message.MagicNumber(),
			"err":    "invalid magic number",
		}).Debug("Failed to verify header.")
		return ErrInvalidMagicNumber
	}

	expectedCheckSum := crc32.ChecksumIEEE(message.HeaderWithoutCheckSum())
	if expectedCheckSum != message.HeaderCheckSum() {
		logging.VLog().WithFields(logrus.Fields{
			"expect": expectedCheckSum,
			"actual": message.HeaderCheckSum(),
			"err":    "invalid header checksum",
		}).Debug("Failed to verify header.")
		return ErrInvalidHeaderCheckSum
	}

	if message.DataLength() > MaxgamcMessageDataLength {
		logging.VLog().WithFields(logrus.Fields{
			"messageName": message.MessageName(),
			"dataLength":  message.DataLength(),
			"limit":       MaxgamcMessageDataLength,
			"err":         "exceeded max data length",
		}).Debug("Failed to verify header.")
		return ErrExceedMaxDataLength
	}

	return nil
}

// VerifyData verify message data
func (message *gamcMessage) VerifyData() error {
	expectedCheckSum := crc32.ChecksumIEEE(message.OriginalData())
	if expectedCheckSum != message.DataCheckSum() {
		logging.VLog().WithFields(logrus.Fields{
			"expect": expectedCheckSum,
			"actual": message.DataCheckSum(),
			"err":    "invalid data checksum",
		}).Debug("Failed to verify data")
		return ErrInvalidDataCheckSum
	}
	return nil
}

// FlagWriteMessageAt flag of write message time
func (message *gamcMessage) FlagWriteMessageAt() {
	message.writeMessageAt = time.Now().UnixNano()
}

// FlagSendMessageAt flag of send message time
func (message *gamcMessage) FlagSendMessageAt() {
	message.sendMessageAt = time.Now().UnixNano()
}

// LatencyFromSendToWrite latency from sendMessage to writeMessage
func (message *gamcMessage) LatencyFromSendToWrite() int64 {
	if message.sendMessageAt == 0 {
		return -1
	} else if message.writeMessageAt == 0 {
		message.FlagWriteMessageAt()
	}

	// convert from nano to millisecond.
	return (message.writeMessageAt - message.sendMessageAt) / int64(time.Millisecond)
}
