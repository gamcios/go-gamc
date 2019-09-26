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

import (
	corepb "gamc.pro/gamcio/go-gamc/core/pb"
	"gamc.pro/gamcio/go-gamc/util/byteutils"
)

// ConsensusState interface of consensus state
type ConsensusState interface {
	RootHash() *corepb.ConsensusRoot
	String() string
	Clone() (ConsensusState, error)
	Replay(ConsensusState) error
	Proposer() byteutils.Hash
	Timestamp() int64
	NextConsensusState(int64, WorldState) (ConsensusState, error)
	Term() ([]byteutils.Hash, error)
	TermRoot() byteutils.Hash
}
