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
	"encoding/json"
	"github.com/BJUT/aqchain/common"
	"github.com/BJUT/aqchain/core/types"
	"github.com/BJUT/aqchain/ethdb"
	"github.com/BJUT/aqchain/log"
	"github.com/BJUT/aqchain/params"
	"github.com/hashicorp/golang-lru"
	"math/big"
	"time"
)

const (
	defaultFullCredit               = 28800 // no punished
	missingPublishCredit            = 100   // punished for missing one block seal
	signRewardCredit                = 10    // seal one block
	autoRewardCredit                = 1     // credit auto recover for each block
	minCalSignerQueueCredit         = 10000 // when calculate the signerQueue

	// the credit of one signer is at least minCalSignerQueueCredit
	candidateStateNormal = 1
	candidateMaxLen      = 500 // if candidateNeedPD is false and candidate is more than candidateMaxLen, then minimum tickets candidates will be remove in each LCRS*loop

	// proposal refund
	proposalRefundDelayLoopCount   = 0
	proposalRefundExpiredLoopCount = proposalRefundDelayLoopCount + 2

	// bug fix
	bugFixBlockNumber = 14456164   // fix bug for header
)

// Snapshot is the state of the authorization voting at a given point in time.
type Snapshot struct {
	config   *params.PoCConfig // Consensus engine parameters to fine tune behavior
	sigcache *lru.ARCCache     // Cache of recent block signatures to speed up ecrecover
	LCRS     uint64            // Loop count to recreate signers from top tally

	CurrentEpoch           uint64  `json:"currentEpoch"`
	Period          uint64                                            `json:"period"`          // Period of seal each block
	Number          uint64                                            `json:"number"`          // Block number where the snapshot was created
	ConfirmedNumber uint64                                            `json:"confirmedNumber"` // Block number confirmed when the snapshot was created
	Hash            common.Hash                                       `json:"hash"`            // Block hash where the snapshot was created
	Validators      []common.Address                                 `json:"validators"`      // Signers queue in current header

	Candidates     map[common.Address]*big.Int              `json:"candidates"`                  // Candidates for Signers (0- adding procedure 1- normal 2- removing procedure)

	MintCount      map[common.Address]uint64           `json:"mintCount"` //
	Punished       map[common.Address]uint64              `json:"punished"`                    // The signer be punished count cause of missing seal
	Proposals       map[common.Hash]*Proposal              `json:"proposals"`       // The Proposals going or success (failed proposal will be removed)

	HeaderTime      uint64                                 `json:"headerTime"`      // Time of the current header
	LoopStartTime   uint64                                 `json:"loopStartTime"`   // Start Time of the current loop
	ProposalRefund  map[uint64]map[common.Address]*big.Int `json:"proposalRefund"`  // Refund proposal deposit

	MinerReward     uint64                                 `json:"minerReward"`     // miner reward per thousand
	MinContribution *big.Int                               `json:"minVoterBalance"` // min voter balance
}

// newSnapshot creates a new snapshot with the specified startup parameters. only ever use if for
// the genesis block.
func newSnapshot(config *params.PoCConfig, sigcache *lru.ARCCache, hash common.Hash, lcrs uint64) *Snapshot {

	snap := &Snapshot{
		config:          config,
		sigcache:        sigcache,
		LCRS:            lcrs,
		CurrentEpoch:(uint64(time.Now().Unix())-config.GenesisTimestamp)/config.Epoch,
		Period:          config.Period,
		Number:          0,
		ConfirmedNumber: 0,
		Hash:            hash,
		Validators:      []common.Address{},
		MintCount:       make(map[common.Address]uint64),
		Punished:        make(map[common.Address]uint64),
		Candidates:      make(map[common.Address]*big.Int),
		Proposals:       make(map[common.Hash]*Proposal),
		HeaderTime:      uint64(time.Now().Unix()) - 1,
		LoopStartTime:   config.GenesisTimestamp,
		ProposalRefund:  make(map[uint64]map[common.Address]*big.Int),
		MinerReward:     minerRewardPerThousand,
		MinContribution: config.MinContribution,
	}

	if len(config.Validators) > 0 {
		for i := 0; i < int(config.MaxValidatorsCount); i++ {
			v:= config.Validators[i%len(config.Validators)]
			snap.Validators = append(snap.Validators, config.Validators[i%len(config.Validators)])
			snap.Candidates[v] = big.NewInt(1000000000000)
		}
	}

	return snap
}

// loadSnapshot loads an existing snapshot from the database.
func loadSnapshot(config *params.PoCConfig, sigcache *lru.ARCCache, db ethdb.Database, hash common.Hash) (*Snapshot, error) {
	blob, err := db.Get(append([]byte("poc-"), hash[:]...))
	if err != nil {
		return nil, err
	}
	snap := new(Snapshot)
	if err := json.Unmarshal(blob, snap); err != nil {
		return nil, err
	}
	snap.config = config
	snap.sigcache = sigcache

	// miner reward per thousand proposal must larger than 0
	// so minerReward is zeron only when update the program
	if snap.MinerReward == 0 {
		snap.MinerReward = minerRewardPerThousand
	}
	if snap.MinContribution == nil {
		snap.MinContribution = new(big.Int).Set(minVoterBalance)
	}
	return snap, nil
}

// store inserts the snapshot into the database.
func (s *Snapshot) store(db ethdb.Database) error {
	blob, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return db.Put(append([]byte("poc-"), s.Hash[:]...), blob)
}

// copy creates a deep copy of the snapshot, though not the individual votes.
func (s *Snapshot) copy() *Snapshot {
	cpy := &Snapshot{
		config:          s.config,
		sigcache:        s.sigcache,
		LCRS:            s.LCRS,
		Period:          s.Period,
		Number:          s.Number,
		ConfirmedNumber: s.ConfirmedNumber,
		Hash:            s.Hash,
		CurrentEpoch: s.CurrentEpoch,

		Validators:    make([]common.Address, len(s.Validators)),
		Candidates:    make(map[common.Address]*big.Int),
		Punished:      make(map[common.Address]uint64),
		Proposals:     make(map[common.Hash]*Proposal),

		HeaderTime:     s.HeaderTime,
		LoopStartTime:  s.LoopStartTime,
		ProposalRefund: make(map[uint64]map[common.Address]*big.Int),

		MinerReward:     s.MinerReward,
		MinContribution: nil,
	}
	copy(cpy.Validators, s.Validators)

	for candidate, state := range s.Candidates {
		cpy.Candidates[candidate] = state
	}
	for signer, cnt := range s.Punished {
		cpy.Punished[signer] = cnt
	}
	for txHash, proposal := range s.Proposals {
		cpy.Proposals[txHash] = proposal.copy()
	}

	for number, refund := range s.ProposalRefund {
		cpy.ProposalRefund[number] = make(map[common.Address]*big.Int)
		for proposer, deposit := range refund {
			cpy.ProposalRefund[number][proposer] = new(big.Int).Set(deposit)
		}
	}
	// miner reward per thousand proposal must larger than 0
	// so minerReward is zeron only when update the program
	if s.MinerReward == 0 {
		cpy.MinerReward = minerRewardPerThousand
	}
	if s.MinContribution == nil {
		cpy.MinContribution = new(big.Int).Set(minVoterBalance)
	} else {
		cpy.MinContribution = new(big.Int).Set(s.MinContribution)
	}

	return cpy
}

// apply creates a new authorization snapshot by applying the given headers to
// the original one.
func (s *Snapshot) apply(headers []*types.Header) (*Snapshot, error) {
	// Allow passing in no headers for cleaner code
	if len(headers) == 0 {
		return s, nil
	}
	// Sanity check that the headers can be applied
	for i := 0; i < len(headers)-1; i++ {
		if headers[i+1].Number.Uint64() != headers[i].Number.Uint64()+1 {
			return nil, errInvalidVotingChain
		}
	}
	if headers[0].Number.Uint64() != s.Number+1 {
		return nil, errInvalidVotingChain
	}
	// Iterate through the headers and create a new snapshot
	snap := s.copy()

	for _, header := range headers {
		// Resolve the authorization key and check against signers
		coinbase, err := ecrecover(header, s.sigcache)
		if err != nil {
			return nil, err
		}
		if coinbase.Str() != header.Coinbase.Str() && header.Number.Cmp(big.NewInt(bugFixBlockNumber)) != 0{
			return nil, errUnauthorized
		}

		headerExtra := HeaderExtra{}
		err = decodeHeaderExtra(s.config, header.Number, header.Extra[extraVanity:len(header.Extra)-extraSeal], &headerExtra)
		if err != nil {
			return nil, err
		}
		snap.HeaderTime = header.Time.Uint64()

		//snap.MintCount[header.Coinbase]++

		// 更新候选人贡献值
		snap.updateCandidates(headerExtra.CurrentBlockJoin)

		// 更新验证人
		snap.updateValidators()

		// deal the snap related with punished
		//snap.updateSnapshotForPunish(headerExtra.SignerMissing, header.Number, header.Coinbase)

		// deal proposals
		snap.updateSnapshotByProposals(headerExtra.CurrentBlockProposals, header.Number)

		// deal declares
		snap.updateSnapshotByDeclares(headerExtra.CurrentBlockDeclares, header.Number)

		// deal trantor upgrade
		if snap.Period == 0 {
			snap.Period = snap.config.Period
		}

		// calculate proposal result
		//snap.calculateProposalResult(header.Number)

		// check the len of candidate if not candidateNeedPD
		if !candidateNeedPD && (snap.Number+1)%(snap.config.MaxValidatorsCount*snap.LCRS) == 0 && len(snap.Candidates) > candidateMaxLen {
			snap.removeExtraCandidate()
		}

	}
	snap.Number += uint64(len(headers))
	snap.Hash = headers[len(headers)-1].Hash()

	return snap, nil
}

func (s *Snapshot) removeExtraCandidate() {
	// remove minimum tickets tally beyond candidateMaxLen
	sorted := s.buildCandidateSlice()
	if len(sorted) > candidateMaxLen {
		removeNeed := sorted[candidateMaxLen:]
		for _, c := range removeNeed {
			delete(s.Candidates, c.addr)
		}
	}
}

func (s *Snapshot) updateSnapshotByDeclares(declares []Declare, headerNumber *big.Int) {
	for _, declare := range declares {
		if proposal, ok := s.Proposals[declare.ProposalHash]; ok {
			// check the proposal enable status and valid block number
			if proposal.ReceivedNumber.Uint64()+proposal.ValidationLoopCnt*s.config.MaxValidatorsCount < headerNumber.Uint64() || !s.isCandidate(declare.Declarer) {
				continue
			}
			// check if this signer already declare on this proposal
			alreadyDeclare := false
			for _, v := range proposal.Declares {
				if v.Declarer.Str() == declare.Declarer.Str() {
					// this declarer already declare for this proposal
					alreadyDeclare = true
					break
				}
			}
			if alreadyDeclare {
				continue
			}
			// add declare to proposal
			s.Proposals[declare.ProposalHash].Declares = append(s.Proposals[declare.ProposalHash].Declares,
				&Declare{declare.ProposalHash, declare.Declarer, declare.Decision})

		}
	}
}

//func (s *Snapshot) calculateProposalResult(headerNumber *big.Int) {
//	// process the expire proposal refund record
//	expiredHeaderNumber := headerNumber.Uint64() - proposalRefundExpiredLoopCount*s.config.MaxValidatorsCount
//	if _, ok := s.ProposalRefund[expiredHeaderNumber]; ok {
//		delete(s.ProposalRefund, expiredHeaderNumber)
//	}
//
//	for hashKey, proposal := range s.Proposals {
//		// the result will be calculate at receiverdNumber + vlcnt + 1
//		if proposal.ReceivedNumber.Uint64()+proposal.ValidationLoopCnt*s.config.MaxValidatorsCount+1 == headerNumber.Uint64() {
//			//return deposit for proposal
//			if _, ok := s.ProposalRefund[headerNumber.Uint64()]; !ok {
//				s.ProposalRefund[headerNumber.Uint64()] = make(map[common.Address]*big.Int)
//			}
//			if _, ok := s.ProposalRefund[headerNumber.Uint64()][proposal.Proposer]; !ok {
//				s.ProposalRefund[headerNumber.Uint64()][proposal.Proposer] = new(big.Int).Set(proposal.CurrentDeposit)
//			} else {
//				s.ProposalRefund[headerNumber.Uint64()][proposal.Proposer].Add(s.ProposalRefund[headerNumber.Uint64()][proposal.Proposer], proposal.CurrentDeposit)
//			}
//
//			// calculate the current stake of this proposal
//			judegmentStake := big.NewInt(0)
//			for _, tally := range s.Tally {
//				judegmentStake.Add(judegmentStake, tally)
//			}
//			judegmentStake.Mul(judegmentStake, big.NewInt(2))
//			judegmentStake.Div(judegmentStake, big.NewInt(3))
//			// calculate declare stake
//			yesDeclareStake := big.NewInt(0)
//			for _, declare := range proposal.Declares {
//				if declare.Decision {
//					if _, ok := s.Tally[declare.Declarer]; ok {
//						yesDeclareStake.Add(yesDeclareStake, s.Tally[declare.Declarer])
//					}
//				}
//			}
//			if yesDeclareStake.Cmp(judegmentStake) > 0 {
//				// process add candidate
//				switch proposal.ProposalType {
//				case proposalTypeCandidateAdd:
//					if candidateNeedPD {
//						s.Candidates[proposal.TargetAddress] = candidateStateNormal
//					}
//				case proposalTypeCandidateRemove:
//					if _, ok := s.Candidates[proposal.TargetAddress]; ok && candidateNeedPD {
//						delete(s.Candidates, proposal.TargetAddress)
//					}
//				default:
//					// todo
//				}
//			}
//
//			// remove all proposal
//			delete(s.Proposals, hashKey)
//		}
//
//	}
//
//}

func (s *Snapshot) updateSnapshotByProposals(proposals []Proposal, headerNumber *big.Int) {
	for _, proposal := range proposals {
		proposal.ReceivedNumber = new(big.Int).Set(headerNumber)
		s.Proposals[proposal.Hash] = &proposal
	}
}

func (s *Snapshot) updateSnapshotForPunish(signerMissing []common.Address, headerNumber *big.Int, coinbase common.Address) {
	// set punished count to half of origin in Epoch
	/*
		if headerNumber.Uint64()%s.config.Epoch == 0 {
			for bePublished := range s.Punished {
				if count := s.Punished[bePublished] / 2; count > 0 {
					s.Punished[bePublished] = count
				} else {
					delete(s.Punished, bePublished)
				}
			}
		}
	*/
	// punish the missing signer
	for _, signerEach := range signerMissing {
		if _, ok := s.Punished[signerEach]; ok {
			// 10 times of defaultFullCredit is big enough for calculate signer order
			if s.Punished[signerEach] <= 10*defaultFullCredit {
				s.Punished[signerEach] += missingPublishCredit
			}
		} else {
			s.Punished[signerEach] = missingPublishCredit
		}
	}
	// reduce the punish of sign signer
	if _, ok := s.Punished[coinbase]; ok {

		if s.Punished[coinbase] > signRewardCredit {
			s.Punished[coinbase] -= signRewardCredit
		} else {
			delete(s.Punished, coinbase)
		}
	}
	// reduce the punish for all punished
	for signerEach := range s.Punished {
		if s.Punished[signerEach] > autoRewardCredit {
			s.Punished[signerEach] -= autoRewardCredit
		} else {
			delete(s.Punished, signerEach)
		}
	}

	s.Punished = make(map[common.Address]uint64)

}

// inturn returns if a signer at a given block height is in-turn or not.
func (s *Snapshot) inturn(signer common.Address, headerTime uint64) bool {
	// if all node stop more than period of one loop
	if signersCount := len(s.Validators); signersCount > 0 {
		if loopIndex := ((headerTime - s.LoopStartTime) / s.config.Period) % uint64(signersCount); s.Validators[loopIndex] == signer {
			return true
		}
	}
	return false

}

// check if address belong to candidate
func (s *Snapshot) isCandidate(address common.Address) bool {
	if _, ok := s.Candidates[address]; ok {
		return true
	}
	return false
}

func (s *Snapshot) calculateProposalRefund() map[common.Address]*big.Int {
	if refund, ok := s.ProposalRefund[s.Number-proposalRefundDelayLoopCount*s.config.MaxValidatorsCount]; ok {
		return refund
	}
	return make(map[common.Address]*big.Int)
}

// 更新验证人
func (s *Snapshot) updateValidators() {
	// 进入下一个时期更新验证人列表
	epoch:= (s.HeaderTime - s.config.GenesisTimestamp) / s.config.Epoch
	if  epoch > s.CurrentEpoch{
		log.Info("进入下个时期 ", "pre", s.CurrentEpoch , "next", epoch)
		// 按照贡献值排序候选人
		var validators []common.Address
		sorted:= s.buildCandidateSlice()
		max:= int(s.config.MaxValidatorsCount)
		for index,val:= range sorted{
			if index >=max{
				break
			}
			validators = append(validators, val.addr)
		}
		s.Validators = validators
		s.CurrentEpoch = epoch
	}
}

func (s *Snapshot) updateCandidates(join []Join) {
	for _,j:= range join{
		s.Candidates[j.Sender]=j.Contribution
	}
}