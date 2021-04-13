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
	"encoding/hex"
	"fmt"
	"github.com/BJUT/aqchain/accounts/abi"
	"github.com/BJUT/aqchain/common/hexutil"
	"github.com/BJUT/aqchain/params"
	"math/big"
	"strconv"
	"strings"

	"github.com/BJUT/aqchain/common"
	"github.com/BJUT/aqchain/consensus"
	"github.com/BJUT/aqchain/core/state"
	"github.com/BJUT/aqchain/core/types"
	"github.com/BJUT/aqchain/rlp"
)

const (
	/*
	 *  ufo:version:category:action/data
	 */
	ufoPrefix           = "ufo"
	ufoVersion          = "1"
	ufoCategoryEvent    = "event"
	ufoEventJoin    = "join"
	ufoEventProposal    = "proposal"
	ufoEventDeclare     = "declare"
	ufoMinSplitLen      = 3
	posPrefix           = 0
	posVersion          = 1
	posCategory         = 2

	posEventJoin = 3
	posEventProposal    = 3
	posEventDeclare     = 3
	/*
	 *  proposal type
	 */
	proposalTypeCandidateAdd                  = 1
	proposalTypeCandidateRemove               = 2

	/*
	 * proposal related
	 */
	maxValidationLoopCnt     = 50000                   // About one month if period = 3 & 21 super nodes
	minValidationLoopCnt     = 4                       // just for test, Note: 12350  About three days if seal each block per second & 21 super nodes
	defaultValidationLoopCnt = 10000                   // About one week if period = 3 & 21 super nodes
	maxProposalDeposit       = 100000                  // If no limit on max proposal deposit and 1 billion TTC deposit success passed, then no new proposal.
)

// RefundGas :
// refund gas to tx sender
type RefundGas map[common.Address]*big.Int

// RefundPair :
type RefundPair struct {
	Sender   common.Address
	GasPrice *big.Int
}

// RefundHash :
type RefundHash map[common.Hash]RefundPair

// Proposal :
// proposal come from  custom tx which data like "ufo:1:event:proposal:candidate:add:address" or "ufo:1:event:proposal:percentage:60"
// proposal only come from the current candidates
// not only candidate add/remove , current signer can proposal for params modify like percentage of reward distribution ...
type Proposal struct {
	Hash                   common.Hash    // tx hash
	ReceivedNumber         *big.Int       // block number of proposal received
	CurrentDeposit         *big.Int       // received deposit for this proposal
	ValidationLoopCnt      uint64         // validation block number length of this proposal from the received block number
	ProposalType           uint64         // type of proposal 1 - add candidate 2 - remove candidate ...
	Proposer               common.Address // proposer
	TargetAddress          common.Address // candidate need to add/remove if candidateNeedPD == true
	MinerRewardPerThousand uint64         // reward of miner + side chain miner
	Declares               []*Declare     // Declare this proposal received (always empty in block header)
	MinVoterBalance        uint64         // value of minVoterBalance , need to mul big.Int(1e+18)
	ProposalDeposit        uint64         // The deposit need to be frozen during before the proposal get final conclusion. (TTC)
}

func (p *Proposal) copy() *Proposal {
	cpy := &Proposal{
		Hash:                   p.Hash,
		ReceivedNumber:         new(big.Int).Set(p.ReceivedNumber),
		CurrentDeposit:         new(big.Int).Set(p.CurrentDeposit),
		ValidationLoopCnt:      p.ValidationLoopCnt,
		ProposalType:           p.ProposalType,
		Proposer:               p.Proposer,
		TargetAddress:          p.TargetAddress,
		MinerRewardPerThousand: p.MinerRewardPerThousand,
		Declares:               make([]*Declare, len(p.Declares)),
		MinVoterBalance:        p.MinVoterBalance,
		ProposalDeposit:        p.ProposalDeposit,
	}

	copy(cpy.Declares, p.Declares)
	return cpy
}

// Declare :
// declare come from custom tx which data like "ufo:1:event:declare:hash:yes"
// proposal only come from the current candidates
// hash is the hash of proposal tx
type Declare struct {
	ProposalHash common.Hash
	Declarer     common.Address
	Decision     bool
}

type Join struct {
	Sender common.Address
	Contribution *big.Int
}

// HeaderExtra is the struct of info in header.Extra[extraVanity:len(header.extra)-extraSeal]
// HeaderExtra is the current struct
type HeaderExtra struct {
	CurrentBlockProposals     []Proposal
	CurrentBlockDeclares      []Declare
	CurrentBlockJoin      []Join
}

// Encode HeaderExtra
func encodeHeaderExtra(config *params.PoCConfig, number *big.Int, val HeaderExtra) ([]byte, error) {

	var headerExtra interface{}
	switch {
	//case config.IsTrantor(number):

	default:
		headerExtra = val
	}
	return rlp.EncodeToBytes(headerExtra)

}

// Decode HeaderExtra
func decodeHeaderExtra(config *params.PoCConfig, number *big.Int, b []byte, val *HeaderExtra) error {
	var err error
	switch {
	//case config.IsTrantor(number):
	default:
		err = rlp.DecodeBytes(b, val)
	}
	return err
}

// Build side chain confirm data
func (a *PoC) buildSCEventConfirmData(scHash common.Hash, headerNumber *big.Int, headerTime *big.Int, lastLoopInfo string, chargingInfo string) []byte {
	return []byte(fmt.Sprintf("%s:%s:%s:%s:%s:%d:%d:%s:%s",
		ufoPrefix, ufoVersion,
		scHash.Hex(), headerNumber.Uint64(), headerTime.Uint64(), lastLoopInfo, chargingInfo))

}

// Calculate Votes from transaction in this block, write into header.Extra
func (a *PoC) processCustomTx(headerExtra HeaderExtra, chain consensus.ChainReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, receipts []*types.Receipt) (HeaderExtra, RefundGas, error) {
	// if predecessor voter make transaction and vote in this block,
	// just process as vote, do it in snapshot.apply
	var (
		snap       *Snapshot
		err        error
		number     uint64
		refundGas  RefundGas
		refundHash RefundHash
	)
	refundGas = make(map[common.Address]*big.Int)
	refundHash = make(map[common.Hash]RefundPair)
	number = header.Number.Uint64()
	if number > 1 {
		snap, err = a.snapshot(chain, number-1, header.ParentHash, nil, defaultLoopCntRecalculateSigners)
		if err != nil {
			return headerExtra, nil, err
		}
	}

	for index, tx := range txs {

		// 访问了指定合约
		if tx.To().Hex() == "0xB28BeC7ab0E007AE4Ad371Da51eF39F54719Be6A"{
			// 反编输入参数和执行结果
			inputHex:= hexutil.Encode(tx.Data())
			//fmt.Println(inputHex)
			// set方法abi码
			fn:= "0xa815ff15"
			// 如果是set方法获取输入参数
			if inputHex[:10] == fn{
				type fnSet struct {
					Fid string
					Owner common.Address
				}
				definition := `[{"name": "set", "type": "event", "inputs": [{"name":"fid", "type":"string"},{"name":"owner", "type":"address"}]}]`
				parsedABI, _ := abi.JSON(strings.NewReader(definition))
				var rst fnSet
				inputHex = inputHex[10:]
				encb, _ := hex.DecodeString(inputHex)
				err = parsedABI.Unpack(&rst, "set", encb)
				//fmt.Println(rst.Fid)
				//fmt.Println(rst.Owner.Hex())
				rcp:= receipts[index]

				if rcp.Status == types.ReceiptStatusSuccessful {
					// 处理贡献值 这里是给文件的创建者贡献值  todo 判断这个账户是否存在？ 贡献值分配
					state.AddContribution(rst.Owner,	big.NewInt(100000000000000000))
				}

				fmt.Println(state.GetContribution(rst.Owner))
			}
		}

		txSender, err := types.Sender(types.NewEIP155Signer(tx.ChainId()), tx)
		if err != nil {
			continue
		}

		if len(string(tx.Data())) >= len(ufoPrefix) {
			txData := string(tx.Data())
			txDataInfo := strings.Split(txData, ":")
			if len(txDataInfo) >= ufoMinSplitLen {
				if txDataInfo[posPrefix] == ufoPrefix {
					if txDataInfo[posVersion] == ufoVersion {
						// process vote event
						if txDataInfo[posCategory] == ufoCategoryEvent {
							if len(txDataInfo) > ufoMinSplitLen {
								if txDataInfo[posEventJoin] == ufoEventJoin  && tx.To() != nil && state.GetContribution(txSender).Cmp(snap.MinContribution) > 0 {
									headerExtra.CurrentBlockJoin = a.processEventJoin(headerExtra.CurrentBlockJoin, state, txSender)
								} else if txDataInfo[posEventProposal] == ufoEventProposal {
									headerExtra.CurrentBlockProposals = a.processEventProposal(headerExtra.CurrentBlockProposals, txDataInfo, state, tx, txSender, snap)
								} else if txDataInfo[posEventDeclare] == ufoEventDeclare && snap.isCandidate(txSender) {
									headerExtra.CurrentBlockDeclares = a.processEventDeclare(headerExtra.CurrentBlockDeclares, txDataInfo, tx, txSender)
								}
							} else {
								// todo : something wrong, leave this transaction to process as normal transaction
							}
						}
					}
				}
			}
		}
	}

	for _, receipt := range receipts {
		if pair, ok := refundHash[receipt.TxHash]; ok && receipt.Status == 1 {
			pair.GasPrice.Mul(pair.GasPrice, big.NewInt(int64(receipt.GasUsed)))
			refundGas = a.refundAddGas(refundGas, pair.Sender, pair.GasPrice)
		}
	}
	return headerExtra, refundGas, nil
}

func (a *PoC) refundAddGas(refundGas RefundGas, address common.Address, value *big.Int) RefundGas {
	if _, ok := refundGas[address]; ok {
		refundGas[address].Add(refundGas[address], value)
	} else {
		refundGas[address] = value
	}

	return refundGas
}

func (a *PoC) processEventProposal(currentBlockProposals []Proposal, txDataInfo []string, state *state.StateDB, tx *types.Transaction, proposer common.Address, snap *Snapshot) []Proposal {
	// sample for add side chain proposal
	// eth.sendTransaction({from:eth.accounts[0],to:eth.accounts[0],value:0,data:web3.toHex("ufo:1:event:proposal:proposal_type:4:sccount:2:screward:50:schash:0x3210000000000000000000000000000000000000000000000000000000000000:vlcnt:4")})
	// sample for declare
	// eth.sendTransaction({from:eth.accounts[0],to:eth.accounts[0],value:0,data:web3.toHex("ufo:1:event:declare:hash:0x853e10706e6b9d39c5f4719018aa2417e8b852dec8ad18f9c592d526db64c725:decision:yes")})
	if len(txDataInfo) <= posEventProposal+2 {
		return currentBlockProposals
	}

	proposal := Proposal{
		Hash:                   tx.Hash(),
		ReceivedNumber:         big.NewInt(0),
		CurrentDeposit:         proposalDeposit, // for all type of deposit
		ValidationLoopCnt:      defaultValidationLoopCnt,
		ProposalType:           proposalTypeCandidateAdd,
		Proposer:               proposer,
		TargetAddress:          common.Address{},
		MinerRewardPerThousand: minerRewardPerThousand,
		Declares:               []*Declare{},
		MinVoterBalance:        new(big.Int).Div(minVoterBalance, big.NewInt(1e+18)).Uint64(),
		ProposalDeposit:        new(big.Int).Div(proposalDeposit, big.NewInt(1e+18)).Uint64(), // default value
	}

	for i := 0; i < len(txDataInfo[posEventProposal+1:])/2; i++ {
		k, v := txDataInfo[posEventProposal+1+i*2], txDataInfo[posEventProposal+2+i*2]
		switch k {
		case "vlcnt":
			// If vlcnt is missing then user default value, but if the vlcnt is beyond the min/max value then ignore this proposal
			if validationLoopCnt, err := strconv.Atoi(v); err != nil || validationLoopCnt < minValidationLoopCnt || validationLoopCnt > maxValidationLoopCnt {
				return currentBlockProposals
			} else {
				proposal.ValidationLoopCnt = uint64(validationLoopCnt)
			}
		case "proposal_type":
			if proposalType, err := strconv.Atoi(v); err != nil {
				return currentBlockProposals
			} else {
				proposal.ProposalType = uint64(proposalType)
			}
		case "candidate":
			// not check here
			proposal.TargetAddress.UnmarshalText([]byte(v))
		case "mrpt":
			// miner reward per thousand
			if mrpt, err := strconv.Atoi(v); err != nil || mrpt <= 0 || mrpt > 1000 {
				return currentBlockProposals
			} else {
				proposal.MinerRewardPerThousand = uint64(mrpt)
			}
		case "mvb":
			// minVoterBalance
			if mvb, err := strconv.Atoi(v); err != nil || mvb <= 0 {
				return currentBlockProposals
			} else {
				proposal.MinVoterBalance = uint64(mvb)
			}
		case "mpd":
			// proposalDeposit
			if mpd, err := strconv.Atoi(v); err != nil || mpd <= 0 || mpd > maxProposalDeposit {
				return currentBlockProposals
			} else {
				proposal.ProposalDeposit = uint64(mpd)
			}
		case "scrt":
			// target address on side chain to charge gas
			proposal.TargetAddress.UnmarshalText([]byte(v))
		}
	}
	// now the proposal is built
	currentProposalPay := new(big.Int).Set(proposalDeposit)

	// check enough balance for deposit
	if state.GetBalance(proposer).Cmp(currentProposalPay) < 0 {
		return currentBlockProposals
	}
	// collection the fee for this proposal (deposit and other fee , sc rent fee ...)
	state.SetBalance(proposer, new(big.Int).Sub(state.GetBalance(proposer), currentProposalPay))

	return append(currentBlockProposals, proposal)
}

func (a *PoC) processEventDeclare(currentBlockDeclares []Declare, txDataInfo []string, tx *types.Transaction, declarer common.Address) []Declare {
	if len(txDataInfo) <= posEventDeclare+2 {
		return currentBlockDeclares
	}
	declare := Declare{
		ProposalHash: common.Hash{},
		Declarer:     declarer,
		Decision:     true,
	}
	for i := 0; i < len(txDataInfo[posEventDeclare+1:])/2; i++ {
		k, v := txDataInfo[posEventDeclare+1+i*2], txDataInfo[posEventDeclare+2+i*2]
		switch k {
		case "hash":
			declare.ProposalHash.UnmarshalText([]byte(v))
		case "decision":
			if v == "yes" {
				declare.Decision = true
			} else if v == "no" {
				declare.Decision = false
			} else {
				return currentBlockDeclares
			}
		}
	}

	return append(currentBlockDeclares, declare)
}

func (a *PoC) processEventJoin(currentBlockJoins []Join,stateDB *state.StateDB, sender common.Address) []Join{
	a.lock.RLock()
	contr := stateDB.GetContribution(sender)
	a.lock.RUnlock()

	currentBlockJoins = append(currentBlockJoins, Join{
		Sender:     sender,
		Contribution:     contr,
	})

	return currentBlockJoins
}

