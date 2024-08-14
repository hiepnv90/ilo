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
keystore_dir: "keystore"
router_address: "0x68b3465833fb72a70ecdf485e0e4c7bd8665fc45" # Uniswap v3 router address
input_token: "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee" # ETH
output_token: "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48" # USDC
fee_tier: 500 # 0.05%, fee tier of uniswap v3 pool
gas_tip_multiplier: 1.0
#start_time: "2024-08-01T00:00:00Z" # Run immediately if omitted.
#gas_limit: 300000 # Call node to estimate gas if omitted.
#min_return_amount: 7000000000 # 7000 USDC
weth: "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"
accounts:
  - address: "0x0000000000000000000001111111111111111111"
    passphrase: "123456"
    amount: 3000000000000000000 # 3 ETH
    priv_key: "" # optional, set this empty to use keystore
    #recipient: "" # recipient wallet, default is account address.
    max_gas_fee: 200000000000000000 # 0.2 ETH, default is estimated from metamask API.
    #min_return_amount: 12000000000 # 12000 USDC, if omitted, use global value set above.
```

Example keystore file:
```json
{"address":"1111111111111111111111111111111111111111","crypto":{"cipher":"aes-128-ctr","ciphertext":"encrypted_ciphertext","cipherparams":{"iv":"iv"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"salt"},"mac":"mac"},"id":"id","version":3}
```

**NOTE**:
1. Keystore directory contains encrypted private keys and store in json format.
1. Replace `passphrase` of accounts with correct passphrase to decrypt private keys.
1. Replace `output_token` to sale token.
1. Router address is different between chains. For base, the address is `0x2626664c2603336e57b271c5c0b26f421741e481`.
1. Weth address is different between chains. For base, the address is `0x4200000000000000000000000000000000000006`.
1. Need to find the correct fee tier for uniswap v3 pool, so the router can find the correct pool for swap.
