package blockchain

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"
)

func TestEncodeSwap(t *testing.T) {
	encodedData, err := EncodeSwap(
		common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
		common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
		common.HexToAddress("0x0000000000000000000111111111111111111111"),
		big.NewInt(3000000000000000),
		big.NewInt(1),
		1723275906,
		big.NewInt(500),
	)
	require.NoError(t, err)
	t.Log(hexutil.Encode(encodedData))
}

func TestEncodeSwap02(t *testing.T) {
	minReturnAmount, _ := new(big.Int).SetString("10000000000000000000000000000", 10)
	encodedData, err := EncodeSwap02(
		common.HexToAddress("0x4200000000000000000000000000000000000006"),
		common.HexToAddress("0x6b9bb36519538e0c073894e964e90172e1c0b41f"),
		common.HexToAddress("0x719911dCe2e792b93D74370c188f0E4AEc0860ec"),
		big.NewInt(3000000000000000),
		minReturnAmount,
		big.NewInt(10000),
	)
	require.NoError(t, err)
	t.Log(hexutil.Encode(encodedData))
}
