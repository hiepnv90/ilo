package blockchain

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const (
	methodExactInputSingle = "exactInputSingle"
)

type ExactInputSingleParams struct {
	TokenIn           common.Address
	TokenOut          common.Address
	Fee               *big.Int
	Recipient         common.Address
	Deadline          *big.Int
	AmountIn          *big.Int
	AmountOutMinimum  *big.Int
	SqrtPriceLimitX96 *big.Int
}

func EncodeSwap(
	inputToken common.Address,
	outputToken common.Address,
	recipient common.Address,
	inputAmount *big.Int,
	minOutputAmount *big.Int,
	deadline int64,
	fee *big.Int,
) ([]byte, error) {
	return uniswapV3RouterABI.Pack(
		methodExactInputSingle,
		ExactInputSingleParams{
			TokenIn:           inputToken,
			TokenOut:          outputToken,
			Fee:               fee,
			Recipient:         recipient,
			Deadline:          big.NewInt(deadline),
			AmountIn:          inputAmount,
			AmountOutMinimum:  minOutputAmount,
			SqrtPriceLimitX96: big.NewInt(0),
		},
	)
}
