# Krystal ILO Trader

A small program that help to create trades for Krystal ILO sales.

## Quick Start

```sh
go run ./cmd/app/main.go --config /etc/config.example.yaml
```

**NOTE**:
1. Keystore directory contains encrypted private keys and store in json format.
1. Replace `passphrase` of accounts with correct passphrase to decrypt private keys.
1. Replace `output_token` to sale token.
