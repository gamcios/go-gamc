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
	"container/heap"
)

type timeHeap []*Transaction

func (h timeHeap) Len() int      { return len(h) }
func (h timeHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h timeHeap) Less(i, j int) bool {
	return h[i].Timestamp() < h[j].Timestamp()
}

func (h *timeHeap) Push(x interface{}) {
	*h = append(*h, x.(*Transaction))
}

func (h *timeHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// PopTxs
func (h *timeHeap) PopTxs(size int) []*Transaction {
	res := make([]*Transaction, 0)
	if size <= 0 {
		return res
	}
	count := 0
	for count < size {
		tx := heap.Pop(h)
		res = append(res, tx.(*Transaction))
		count++
	}
	return res
}
