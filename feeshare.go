// feeshare.go
package bags

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// -------------------- Get Fee Share Wallet --------------------

// GetFeeShareWallet resolves the fee share wallet address associated with a
// Twitter username.
//
// API Reference (Bags): "Get Fee Share Wallet"
// - Method: GET
// - Path: /token-launch/fee-share/wallet/twitter
// - Base URL: https://public-api-v2.bags.fm/api/v1
// - Security: header "x-api-key: <YOUR_API_KEY>"
// - Query params:
//   - twitterUsername (string): Twitter username/handle (without @)
//
// Example request:
//
//	GET /token-launch/fee-share/wallet/twitter?twitterUsername=elonmusk
//	x-api-key: <YOUR_API_KEY>
//
// Example response (200):
//
//	{"success": true, "response": "<string>"}
//
// Error responses:
//
//	400/401/500: {"success": false, "error": "<string>"}
//
// Returns the wallet address as a string on success.
func (c *BagsClient) GetFeeShareWallet(ctx context.Context, twitterUsername string) (string, error) {
	handle := strings.TrimSpace(twitterUsername)
	if handle == "" {
		return "", fmt.Errorf("twitterUsername is required")
	}

	// Build query: token-launch/fee-share/wallet/twitter?twitterUsername=<handle>
	rel := "token-launch/fee-share/wallet/twitter"
	q := url.Values{}
	q.Set("twitterUsername", handle)
	relWithQuery := rel + "?" + q.Encode()

	req, err := c.newRequest(ctx, http.MethodGet, relWithQuery, nil, "")
	if err != nil {
		return "", err
	}

	var env struct {
		Success  bool   `json:"success"`
		Response string `json:"response"`
	}
	if err := c.do(req, &env); err != nil {
		return "", err
	}
	if !env.Success || strings.TrimSpace(env.Response) == "" {
		return "", fmt.Errorf("unexpected response")
	}
	return env.Response, nil
}

// -------------------- Create Fee Share Config --------------------

// CreateFeeShareConfigRequest is the request body for
// POST /token-launch/fee-share/create-config.
//
// API Reference (Bags): "Create Fee Share Config creation transaction"
// - Description: Create a custom fee sharing config between two wallets.
// - Security: header "x-api-key: <YOUR_API_KEY>"
// - Body fields:
//   - walletA (string): First wallet address (base58 public key)
//   - walletB (string): Second wallet address (base58 public key)
//   - walletABps (int): Basis points allocated to walletA (0-10000)
//   - walletBBps (int): Basis points allocated to walletB (0-10000)
//   - payer (string): Payer wallet (public key) that will cover fees
//   - baseMint (string): Token mint public key to which fees apply
//   - quoteMint (string): Quote mint public key; must be wSOL mint currently
//
// Example body:
//
//	{
//	  "walletA": "<pubkeyA>",
//	  "walletB": "<pubkeyB>",
//	  "walletABps": 1000,
//	  "walletBBps": 9000,
//	  "payer": "<payerPubkey>",
//	  "baseMint": "<tokenMint>",
//	  "quoteMint": "So11111111111111111111111111111111111111112"
//	}
type CreateFeeShareConfigRequest struct {
	WalletA    string `json:"walletA"`    // First wallet address (base58)
	WalletB    string `json:"walletB"`    // Second wallet address (base58)
	WalletABps int64  `json:"walletABps"` // Basis points for walletA (0-10000)
	WalletBBps int64  `json:"walletBBps"` // Basis points for walletB (0-10000)
	Payer      string `json:"payer"`      // Payer wallet public key
	BaseMint   string `json:"baseMint"`   // Token mint public key
	QuoteMint  string `json:"quoteMint"`  // Quote mint public key (must be wSOL mint at the moment)
}

// CreateFeeShareConfigResult matches the Bags response "response" payload.
//
// Example success envelope:
//
//	{"success": true, "response": {"tx": "<string>", "configKey": "<string>"}}
//
// When the configuration already exists, the "tx" field may be empty or omitted.
type CreateFeeShareConfigResult struct {
	Tx        string `json:"tx"`
	ConfigKey string `json:"configKey"`
}

// CreateFeeShareConfig creates a custom fee sharing configuration between two
// wallets for a given token mint.
//
// API Reference (Bags): "Create Fee Share Config creation transaction"
// - Method: POST
// - Path: /token-launch/fee-share/create-config
// - Base URL: https://public-api-v2.bags.fm/api/v1
// - Security: header "x-api-key: <YOUR_API_KEY>"
// - Request body: JSON as described in CreateFeeShareConfigRequest
//
// Example response (200):
//
//	{"success": true, "response": {"tx": "<string>", "configKey": "<string>"}}
//
// Error responses:
//
//	400/401/500: {"success": false, "error": "<string>"}
func (c *BagsClient) CreateFeeShareConfig(ctx context.Context, in *CreateFeeShareConfigRequest) (*CreateFeeShareConfigResult, error) {
	if in == nil {
		return nil, fmt.Errorf("nil request")
	}
	// Minimal validation; the API ultimately enforces correctness.
	if strings.TrimSpace(in.WalletA) == "" ||
		strings.TrimSpace(in.WalletB) == "" ||
		strings.TrimSpace(in.Payer) == "" ||
		strings.TrimSpace(in.BaseMint) == "" ||
		strings.TrimSpace(in.QuoteMint) == "" {
		return nil, fmt.Errorf("walletA, walletB, payer, baseMint, and quoteMint are required")
	}

	var env struct {
		Success  bool                        `json:"success"`
		Response *CreateFeeShareConfigResult `json:"response"`
	}
	if err := c.postJSON(ctx, "token-launch/fee-share/create-config", in, &env); err != nil {
		return nil, err
	}
	if !env.Success || env.Response == nil {
		return nil, fmt.Errorf("unexpected response")
	}
	return env.Response, nil
}
