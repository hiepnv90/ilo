package blockchain

import (
	"bytes"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

//nolint:gochecknoglobals
var (
	uniswapV3RouterABI   abi.ABI
	uniswapV3Router02ABI abi.ABI
)

//nolint:gochecknoinits
func init() {
	builder := []struct {
		ABI  *abi.ABI
		data []byte
	}{
		{&uniswapV3RouterABI, uniswapV3RouterJSON},
		{&uniswapV3Router02ABI, uniswapV3Router02JSON},
	}

	for _, b := range builder {
		var err error
		*b.ABI, err = abi.JSON(bytes.NewReader(b.data))
		if err != nil {
			panic(err)
		}
	}
}
