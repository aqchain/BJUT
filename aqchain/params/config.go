// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package params

import (
	"fmt"
	"github.com/BJUT/aqchain/common"
	"math/big"
)

var (
	MainnetGenesisHash = common.HexToHash("t0c6f20167950874a396ca06b717df8cd96597bc849b4935003cda98598c5e98be") // Mainnet genesis hash to enforce below configs on
	TestnetGenesisHash = common.HexToHash("t01fa3701d9164c47b587e335e9d4c2a34360dd77c84859c818ba7ea684bdf6ee1") // Testnet genesis hash to enforce below configs on
)

var (
	// MainnetChainConfig is the chain parameters to run a node on the main network.
	MainnetChainConfig = &ChainConfig{
		ChainId:             big.NewInt(9999),
		HomesteadBlock:      big.NewInt(1),
		EIP150Block:         big.NewInt(2),
		EIP150Hash:          common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
		EIP155Block:         big.NewInt(3),
		EIP158Block:         big.NewInt(3),
		ByzantiumBlock:      big.NewInt(4),
		ConstantinopleBlock: nil,
		PoC: &PoCConfig{
			Period:             3,
			Epoch:              201600,
			MaxValidatorsCount: 21,
			MinContribution:    new(big.Int).Mul(big.NewInt(100), big.NewInt(1e+18)),
			GenesisTimestamp:   1554004800,
			Validators: []common.Address{
				common.HexToAddress("0xfc97a1a683fb0108299d5725bc9d72899c9a86b4"),
				common.HexToAddress("0xb01d3dd946d3e2e1f3bf7ba86ff03864f8af47cd"),
				common.HexToAddress("0xbfa6f96861798532a334fd0df82e9f26c626dc59"),
			},
		},
	}

	// MainnetChainConfig is the chain parameters to run a node on the main network.
	RinkebyChainConfig = &ChainConfig{
		ChainId:             big.NewInt(9999),
		HomesteadBlock:      big.NewInt(1),
		EIP150Block:         big.NewInt(2),
		EIP150Hash:          common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
		EIP155Block:         big.NewInt(3),
		EIP158Block:         big.NewInt(3),
		ByzantiumBlock:      big.NewInt(4),
		ConstantinopleBlock: nil,
		PoC: &PoCConfig{
			Period:             3,
			Epoch:              201600,
			MaxValidatorsCount: 21,
			MinContribution:    new(big.Int).Mul(big.NewInt(100), big.NewInt(1e+18)),
			GenesisTimestamp:   1554004800,
			Validators: []common.Address{
				common.HexToAddress("0xfc97a1a683fb0108299d5725bc9d72899c9a86b4"),
				common.HexToAddress("0xb01d3dd946d3e2e1f3bf7ba86ff03864f8af47cd"),
				common.HexToAddress("0xbfa6f96861798532a334fd0df82e9f26c626dc59"),
			},
		},
	}

	// TestnetChainConfig contains the chain parameters to run a node on the Ropsten test network.
	TestnetChainConfig = &ChainConfig{
		ChainId:             big.NewInt(9998),
		HomesteadBlock:      big.NewInt(1),
		EIP150Block:         big.NewInt(2),
		EIP150Hash:          common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
		EIP155Block:         big.NewInt(3),
		EIP158Block:         big.NewInt(3),
		ByzantiumBlock:      big.NewInt(4),
		ConstantinopleBlock: nil,
		PoC: &PoCConfig{
			Period:             3,
			Epoch:              201600,
			MaxValidatorsCount: 21,
			MinContribution:    new(big.Int).Mul(big.NewInt(100), big.NewInt(1e+18)),
			GenesisTimestamp:   1554004800,
			Validators: []common.Address{
				common.HexToAddress("0xfc97a1a683fb0108299d5725bc9d72899c9a86b4"),
				common.HexToAddress("0xb01d3dd946d3e2e1f3bf7ba86ff03864f8af47cd"),
				common.HexToAddress("0xbfa6f96861798532a334fd0df82e9f26c626dc59"),
			},
		},
	}

	// AllEthashProtocolChanges contains every protocol change (EIPs) introduced
	// and accepted by the Ethereum core developers into the Ethash consensus.
	//
	// This configuration is intentionally not using keyed fields to force anyone
	// adding flags to the config to also have to set these fields.
	AllEthashProtocolChanges = &ChainConfig{big.NewInt(1337), big.NewInt(0), big.NewInt(0), common.Hash{}, big.NewInt(0), big.NewInt(0), big.NewInt(0), nil, new(EthashConfig), nil, nil}

	// AllCliqueProtocolChanges contains every protocol change (EIPs) introduced
	// and accepted by the Ethereum core developers into the Clique consensus.
	//
	// This configuration is intentionally not using keyed fields to force anyone
	// adding flags to the config to also have to set these fields.
	AllCliqueProtocolChanges = &ChainConfig{big.NewInt(1337), big.NewInt(0), big.NewInt(0), common.Hash{}, big.NewInt(0), big.NewInt(0), big.NewInt(0), nil, nil, &CliqueConfig{Period: 0, Epoch: 30000}, nil}

	// AllPoCProtocolChanges contains every protocol change (EIPs) introduced
	// and accepted by the Ethereum core developers into the PoC consensus.
	//
	// This configuration is intentionally not using keyed fields to force anyone
	// adding flags to the config to also have to set these fields.
	AllPoCProtocolChanges = &ChainConfig{big.NewInt(1337), big.NewInt(0), big.NewInt(0), common.Hash{}, big.NewInt(0), big.NewInt(0), big.NewInt(0), nil, nil, nil, &PoCConfig{Period: 3, Epoch: 30000, MaxValidatorsCount: 21, MinContribution: big.NewInt(1000000000000000000), GenesisTimestamp: 0, Validators: []common.Address{}}}

	TestChainConfig = &ChainConfig{big.NewInt(1), big.NewInt(0), big.NewInt(0), common.Hash{}, big.NewInt(0), big.NewInt(0), big.NewInt(0), nil, new(EthashConfig), nil, nil}
	TestRules       = TestChainConfig.Rules(new(big.Int))
)

// ChainConfig is the core config which determines the blockchain settings.
//
// ChainConfig is stored in the database on a per block basis. This means
// that any network, identified by its genesis block, can have its own
// set of configuration options.
type ChainConfig struct {
	ChainId *big.Int `json:"chainId"` // Chain id identifies the current chain and is used for replay protection

	HomesteadBlock *big.Int `json:"homesteadBlock,omitempty"` // Homestead switch block (nil = no fork, 0 = already homestead)

	// EIP150 implements the Gas price changes (https://github.com/ethereum/EIPs/issues/150)
	EIP150Block *big.Int    `json:"eip150Block,omitempty"` // EIP150 HF block (nil = no fork)
	EIP150Hash  common.Hash `json:"eip150Hash,omitempty"`  // EIP150 HF hash (needed for header only clients as only gas pricing changed)

	EIP155Block *big.Int `json:"eip155Block,omitempty"` // EIP155 HF block
	EIP158Block *big.Int `json:"eip158Block,omitempty"` // EIP158 HF block

	ByzantiumBlock      *big.Int `json:"byzantiumBlock,omitempty"`      // Byzantium switch block (nil = no fork, 0 = already on byzantium)
	ConstantinopleBlock *big.Int `json:"constantinopleBlock,omitempty"` // Constantinople switch block (nil = no fork, 0 = already activated)

	// Various consensus engines
	Ethash *EthashConfig `json:"ethash,omitempty"`
	Clique *CliqueConfig `json:"clique,omitempty"`
	PoC    *PoCConfig    `json:"poc,omitempty"`
}

// EthashConfig is the consensus engine configs for proof-of-work based sealing.
type EthashConfig struct{}

// String implements the stringer interface, returning the consensus engine details.
func (c *EthashConfig) String() string {
	return "ethash"
}

// CliqueConfig is the consensus engine configs for proof-of-authority based sealing.
type CliqueConfig struct {
	Period uint64 `json:"period"` // Number of seconds between blocks to enforce
	Epoch  uint64 `json:"epoch"`  // Epoch length to reset votes and checkpoint
}

// String implements the stringer interface, returning the consensus engine details.
func (c *CliqueConfig) String() string {
	return "clique"
}

type GenesisAccount struct {
	Balance string `json:"balance"`
}

// PoCLightConfig is the config for light node of poc
type PoCLightConfig struct {
	Alloc map[common.UnprefixedAddress]GenesisAccount `json:"alloc"`
}

// PoCConfig is the consensus engine configs for delegated-proof-of-stake based sealing.
type PoCConfig struct {
	Period             uint64           `json:"period"`            // Number of seconds between blocks to enforce
	Epoch              uint64           `json:"epoch"`             // Epoch length to reset votes and checkpoint
	MaxValidatorsCount uint64           `json:"maxValidatorsCount"` // Max count of validators
	MinContribution    *big.Int         `json:"minContribution"`
	GenesisTimestamp   uint64           `json:"genesisTimestamp"` // The LoopStartTime of first Block
	Validators         []common.Address `json:"validators"`          // Signers vote by themselves to seal the block, make sure the signer accounts are pre-funded
	LightConfig        *PoCLightConfig  `json:"lightConfig,omitempty"`
}

// String implements the stringer interface, returning the consensus engine details.
func (a *PoCConfig) String() string {
	return "poc"
}

// String implements the fmt.Stringer interface.
func (c *ChainConfig) String() string {
	var engine interface{}
	switch {
	case c.Ethash != nil:
		engine = c.Ethash
	case c.Clique != nil:
		engine = c.Clique
	case c.PoC != nil:
		engine = c.PoC
	default:
		engine = "unknown"
	}
	return fmt.Sprintf("{ChainID: %v Homestead: %v EIP150: %v EIP155: %v EIP158: %v Byzantium: %v Constantinople: %v Engine: %v}",
		c.ChainId,
		c.HomesteadBlock,
		c.EIP150Block,
		c.EIP155Block,
		c.EIP158Block,
		c.ByzantiumBlock,
		c.ConstantinopleBlock,
		engine,
	)
}

// IsHomestead returns whether num is either equal to the homestead block or greater.
func (c *ChainConfig) IsHomestead(num *big.Int) bool {
	return isForked(c.HomesteadBlock, num)
}

func (c *ChainConfig) IsEIP150(num *big.Int) bool {
	return isForked(c.EIP150Block, num)
}

func (c *ChainConfig) IsEIP155(num *big.Int) bool {
	return isForked(c.EIP155Block, num)
}

func (c *ChainConfig) IsEIP158(num *big.Int) bool {
	return isForked(c.EIP158Block, num)
}

func (c *ChainConfig) IsByzantium(num *big.Int) bool {
	return isForked(c.ByzantiumBlock, num)
}

func (c *ChainConfig) IsConstantinople(num *big.Int) bool {
	return isForked(c.ConstantinopleBlock, num)
}

// GasTable returns the gas table corresponding to the current phase (homestead or homestead reprice).
//
// The returned GasTable's fields shouldn't, under any circumstances, be changed.
func (c *ChainConfig) GasTable(num *big.Int) GasTable {
	if num == nil {
		return GasTableHomestead
	}
	switch {
	case c.IsEIP158(num):
		return GasTableEIP158
	case c.IsEIP150(num):
		return GasTableEIP150
	default:
		return GasTableHomestead
	}
}

// CheckCompatible checks whether scheduled fork transitions have been imported
// with a mismatching chain configuration.
func (c *ChainConfig) CheckCompatible(newcfg *ChainConfig, height uint64) *ConfigCompatError {
	bhead := new(big.Int).SetUint64(height)

	// Iterate checkCompatible to find the lowest conflict.
	var lasterr *ConfigCompatError
	for {
		err := c.checkCompatible(newcfg, bhead)
		if err == nil || (lasterr != nil && err.RewindTo == lasterr.RewindTo) {
			break
		}
		lasterr = err
		bhead.SetUint64(err.RewindTo)
	}
	return lasterr
}

func (c *ChainConfig) checkCompatible(newcfg *ChainConfig, head *big.Int) *ConfigCompatError {
	if isForkIncompatible(c.HomesteadBlock, newcfg.HomesteadBlock, head) {
		return newCompatError("Homestead fork block", c.HomesteadBlock, newcfg.HomesteadBlock)
	}
	if isForkIncompatible(c.EIP150Block, newcfg.EIP150Block, head) {
		return newCompatError("EIP150 fork block", c.EIP150Block, newcfg.EIP150Block)
	}
	if isForkIncompatible(c.EIP155Block, newcfg.EIP155Block, head) {
		return newCompatError("EIP155 fork block", c.EIP155Block, newcfg.EIP155Block)
	}
	if isForkIncompatible(c.EIP158Block, newcfg.EIP158Block, head) {
		return newCompatError("EIP158 fork block", c.EIP158Block, newcfg.EIP158Block)
	}
	if c.IsEIP158(head) && !configNumEqual(c.ChainId, newcfg.ChainId) {
		return newCompatError("EIP158 chain ID", c.EIP158Block, newcfg.EIP158Block)
	}
	if isForkIncompatible(c.ByzantiumBlock, newcfg.ByzantiumBlock, head) {
		return newCompatError("Byzantium fork block", c.ByzantiumBlock, newcfg.ByzantiumBlock)
	}
	if isForkIncompatible(c.ConstantinopleBlock, newcfg.ConstantinopleBlock, head) {
		return newCompatError("Constantinople fork block", c.ConstantinopleBlock, newcfg.ConstantinopleBlock)
	}
	return nil
}

// isForkIncompatible returns true if a fork scheduled at s1 cannot be rescheduled to
// block s2 because head is already past the fork.
func isForkIncompatible(s1, s2, head *big.Int) bool {
	return (isForked(s1, head) || isForked(s2, head)) && !configNumEqual(s1, s2)
}

// isForked returns whether a fork scheduled at block s is active at the given head block.
func isForked(s, head *big.Int) bool {
	if s == nil || head == nil {
		return false
	}
	return s.Cmp(head) <= 0
}

func configNumEqual(x, y *big.Int) bool {
	if x == nil {
		return y == nil
	}
	if y == nil {
		return x == nil
	}
	return x.Cmp(y) == 0
}

// ConfigCompatError is raised if the locally-stored blockchain is initialised with a
// ChainConfig that would alter the past.
type ConfigCompatError struct {
	What string
	// block numbers of the stored and new configurations
	StoredConfig, NewConfig *big.Int
	// the block number to which the local chain must be rewound to correct the error
	RewindTo uint64
}

func newCompatError(what string, storedblock, newblock *big.Int) *ConfigCompatError {
	var rew *big.Int
	switch {
	case storedblock == nil:
		rew = newblock
	case newblock == nil || storedblock.Cmp(newblock) < 0:
		rew = storedblock
	default:
		rew = newblock
	}
	err := &ConfigCompatError{what, storedblock, newblock, 0}
	if rew != nil && rew.Sign() > 0 {
		err.RewindTo = rew.Uint64() - 1
	}
	return err
}

func (err *ConfigCompatError) Error() string {
	return fmt.Sprintf("mismatching %s in database (have %d, want %d, rewindto %d)", err.What, err.StoredConfig, err.NewConfig, err.RewindTo)
}

// Rules wraps ChainConfig and is merely syntatic sugar or can be used for functions
// that do not have or require information about the block.
//
// Rules is a one time interface meaning that it shouldn't be used in between transition
// phases.
type Rules struct {
	ChainId                                   *big.Int
	IsHomestead, IsEIP150, IsEIP155, IsEIP158 bool
	IsByzantium                               bool
}

func (c *ChainConfig) Rules(num *big.Int) Rules {
	chainId := c.ChainId
	if chainId == nil {
		chainId = new(big.Int)
	}
	return Rules{ChainId: new(big.Int).Set(chainId), IsHomestead: c.IsHomestead(num), IsEIP150: c.IsEIP150(num), IsEIP155: c.IsEIP155(num), IsEIP158: c.IsEIP158(num), IsByzantium: c.IsByzantium(num)}
}
