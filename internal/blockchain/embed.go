package blockchain

import _ "embed"

//go:embed abis/UniswapV3Router.abi.json
var uniswapV3RouterJSON []byte

//go:embed abis/UniswapV3Router02.abi.json
var uniswapV3Router02JSON []byte
