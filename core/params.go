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

import "errors"

// MessageType
const (
	MessageTypeNewBlock                   = "newblock"
	MessageTypeParentBlockDownloadRequest = "dlblock"
	MessageTypeBlockDownloadResponse      = "dlreply"
	MessageTypeNewTx                      = "newtx"
)

var (
	ErrInvalidAddress         = errors.New("address: invalid address")
	ErrInvalidAddressFormat   = errors.New("address: invalid address format")
	ErrInvalidAddressType     = errors.New("address: invalid address type")
	ErrInvalidAddressChecksum = errors.New("address: invalid address checksum")

	ErrInvalidArgument           = errors.New("invalid argument(s)")
	ErrNilArgument               = errors.New("argument(s) is nil")
	ErrInvalidAmount             = errors.New("invalid amount")
	ErrInvalidProtoToBlock       = errors.New("protobuf message cannot be converted into Block")
	ErrInvalidProtoToBlockHeader = errors.New("protobuf message cannot be converted into BlockHeader")
	ErrInvalidProtoToTransaction = errors.New("protobuf message cannot be converted into Transaction")
	ErrInvalidProtoToWitness     = errors.New("protobuf message cannot be converted into Witness")
	ErrInvalidProtoToPsecData    = errors.New("protobuf message cannot be converted into PsecData")

	ErrDuplicatedBlock           = errors.New("duplicated block")
	ErrInvalidChainID            = errors.New("invalid transaction chainID")
	ErrInvalidTransactionHash    = errors.New("invalid transaction hash")
	ErrInvalidBlockHeaderChainID = errors.New("invalid block header chainId")
	ErrInvalidBlockHash          = errors.New("invalid block hash")

	ErrInvalidTransactionSigner = errors.New("invalid transaction signer")
	ErrInvalidPublicKey         = errors.New("invalid public key")

	ErrMissingParentBlock                                = errors.New("cannot find the block's parent block in storage")
	ErrInvalidBlockCannotFindParentInLocalAndTrySync     = errors.New("invalid block received, sync its parent from others")
	ErrInvalidBlockCannotFindParentInLocalAndTryDownload = errors.New("invalid block received, download its parent from others")
	ErrLinkToWrongParentBlock                            = errors.New("link the block to a block who is not its parent")
	ErrCloneAccountState                                 = errors.New("failed to clone account state")

	ErrInvalidBlockStateRoot = errors.New("invalid block state root hash")
	ErrInvalidBlockTxsRoot   = errors.New("invalid block txs root hash")
)
