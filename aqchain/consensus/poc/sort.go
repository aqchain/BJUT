// Copyright 2018 The aqchain Authors
// This file is part of the aqchain library.
//
// The aqchain library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The aqchain library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the aqchain library. If not, see <http://www.gnu.org/licenses/>.

// Package poc implements the delegated-proof-of-stake consensus engine.

package poc

import (
	"bytes"
	"github.com/BJUT/aqchain/common"
	"math/big"
	"sort"
)

type ScoreItem struct {
	addr  common.Address
	score *big.Int
}
type CandidateSlice []ScoreItem

func (s CandidateSlice) Len() int      { return len(s) }
func (s CandidateSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s CandidateSlice) Less(i, j int) bool {
	//we need sort reverse, so ...
	isLess := s[i].score.Cmp(s[j].score)
	if isLess > 0 {
		return true

	} else if isLess < 0 {
		return false
	}
	// if the stake equal
	return bytes.Compare(s[i].addr.Bytes(), s[j].addr.Bytes()) > 0
}

func (s *Snapshot) buildCandidateSlice() CandidateSlice {
	var candidateSlice CandidateSlice
	for address, contribution := range s.Candidates {
		if !candidateNeedPD || s.isCandidate(address) {
			if _, ok := s.Punished[address]; ok {
				var creditWeight uint64
				if s.Punished[address] > defaultFullCredit-minCalSignerQueueCredit {
					creditWeight = minCalSignerQueueCredit
				} else {
					creditWeight = defaultFullCredit - s.Punished[address]
				}
				candidateSlice = append(candidateSlice, ScoreItem{address, new(big.Int).Mul(contribution, big.NewInt(int64(creditWeight)))})
			} else {
				candidateSlice = append(candidateSlice, ScoreItem{address, new(big.Int).Mul(contribution, big.NewInt(defaultFullCredit))})
			}
		}
	}
	sort.Sort(candidateSlice)
	return candidateSlice
}
