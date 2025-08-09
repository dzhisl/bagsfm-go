# bagsfm-go — Go Client for the Bags API

A Go client for the **Bags API** (bags.fm), enabling complete Solana token launch workflows—token metadata, launches, fee-sharing, analytics, and more.

---

## Features

- **Token Launch**: Create token metadata (with image), generate launch config, build launch transaction
- **Fee Share**: Look up the fee-share wallet by Twitter handle, generate fee-sharing configuration
- **Analytics**: Retrieve lifetime fees collected by a token, list token launch creators
- Built-in handling for `x-api-key` header, JSON encoding, multipart file uploads, and error parsing

---

## Installation

```bash
go get github.com/dzhisl/bagsfm-go
```

Or add the import:

```go
import "github.com/dzhisl/bagsfm-go"
```

---

## Usage

```go
client, err := bags.New("your-api-key", nil)
if err != nil {
    // handle error
}
```

---

## Example: Full Token Launch Flow

```go
ctx := context.Background()

// 1. Upload token metadata and image
infoReq := &bags.CreateTokenInfoRequest{
    Name:           "MyToken",
    Symbol:         "MTK",
    Description:    "My first token launch",
    Website:        "https://example.com",
    Image:          imageReader,
    ImageFilename:  "logo.png",
    ImageMIMEType:  "image/png",
}
info, err := client.CreateTokenInfoAndMetadata(ctx, infoReq)
if err != nil { /* handle error */ }

// 2. Generate launch configuration transaction
cfgReq := &bags.CreateTokenLaunchConfigRequest{ LaunchWallet: walletAddress }
cfgRes, err := client.CreateTokenLaunchConfig(ctx, cfgReq)
if err != nil { /* handle error */ }

// 3. Build final launch transaction (base64 encoded)
txReq := &bags.CreateTokenLaunchTxRequest{
    IPFS:               info.TokenMetadata,
    TokenMint:          info.TokenMint,
    Wallet:             walletAddress,
    InitialBuyLamports: 100_000_000,
    ConfigKey:          cfgRes.ConfigKey,
}
txRes, err := client.CreateTokenLaunchTransaction(ctx, txReq)
if err != nil { /* handle error */ }

// txRes.Transaction contains the base64 transaction to send to Solana.
```

---

## Example: Fee Share Workflow

```go
// Get the fee share wallet by Twitter username
feeWallet, err := client.GetFeeShareWallet(ctx, "alice123")
if err != nil { /* handle error */ }

// Create a fee share configuration
fsReq := &bags.CreateFeeShareConfigRequest{
    WalletA:    yourWallet,
    WalletB:    feeWallet,
    WalletABps: 200,       // 2%
    WalletBBps: 9800,      // 98%
    Payer:      yourWallet,
    BaseMint:   baseMint,
    QuoteMint:  quoteMint,
}
fsRes, err := client.CreateFeeShareConfig(ctx, fsReq)
if err != nil { /* handle error */ }

// fsRes.Tx holds the transaction; fsRes.ConfigKey is the config identifier.
```

---

## Example: Analytics

```go
// Get lifetime fees collected for a token
fees, err := client.GetTokenLifetimeFees(ctx, tokenMint)
if err != nil { /* handle error */ }

// List token launch creators
creators, err := client.GetTokenLaunchCreators(ctx, tokenMint)
if err != nil { /* handle error */ }

for _, c := range creators {
    fmt.Printf("Creator: %s (wallet: %s)\n", c.Username, c.Wallet)
}
```

---

## API Key Management & Best Practices

- All requests must include your API key via `x-api-key` header.
- The default base URL is `https://public-api-v2.bags.fm/api/v1/`, as per Bags API versioning.
- Docs highlight rate limiting at **1,000 requests per hour**, so consider implementing exponential backoff.
  ([docs.bags.fm][1])

---

## License & Contribution

MIT License. Contributions, issues, and forks are welcome—whether it’s extra endpoints, better error handling, or documentation enhancements!

---
