package store

import (
	"context"
	"github.com/BJUT/aqchain/accounts/abi/bind"
	"github.com/BJUT/aqchain/common"
	"github.com/BJUT/aqchain/core/types"
	"math/big"
)

type Backend interface {
	bind.ContractBackend
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	BalanceAt(ctx context.Context, address common.Address, blockNum *big.Int) (*big.Int, error)
}
