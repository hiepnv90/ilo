# Krystal ILO Trader

A small program that help to create trades for Krystal ILO sales.

## Quick Start

```sh
go run ./cmd/app/main.go --config internal/config/config.example.yaml
```

Example config file:
```yaml
chain_id: 1
node_rpc: "https://rpc.flashbots.net/fast"
gas_price_endpoint: "https://gas-api.metaswap.codefi.network/networks/1"
krystal_api_endpoint: "https://api.krystal.app/ethereum/v2"
keystore_dir: "keystore"
input_token: "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee" # ETH
output_token: "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48" # USDC
platform_wallet: "0x168E4c3AC8d89B00958B6bE6400B066f0347DDc9" # Krystal Wallet
slippage_bps: 100 # 1%
gas_tip_multiplier: 1.0
#start_time: "2024-08-01T00:00:00Z"
#gas_limit: 500000
#min_return_amount: 9000000000 # 9000 USDC
accounts:
  - address: "0x0000000000000000000001111111111111111111"
    passphrase: "123456"
    amount: 3000000000000000000 # 3 ETH
```

Example keystore file:
```json
{"address":"1111111111111111111111111111111111111111","crypto":{"cipher":"aes-128-ctr","ciphertext":"encrypted_ciphertext","cipherparams":{"iv":"iv"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"salt"},"mac":"mac"},"id":"id","version":3}
```

**NOTE**:
1. Keystore directory contains encrypted private keys and store in json format.
1. Replace `passphrase` of accounts with correct passphrase to decrypt private keys.
1. Replace `output_token` to sale token.
