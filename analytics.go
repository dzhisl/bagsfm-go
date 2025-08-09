// analytics.go
package bags

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// -------------------- Analytics: Get Token Lifetime Fees --------------------

// Retrieve the total lifetime fees collected for a specific token.
//
// GET /token-launch/lifetime-fees?tokenMint=<string>
// Authorization: x-api-key header required.
//
// Response:
//
//	{
//	  "success": true,
//	  "response": "<string>"
//	}
func (c *BagsClient) GetTokenLifetimeFees(ctx context.Context, tokenMint string) (string, error) {
	if tm := strings.TrimSpace(tokenMint); tm == "" {
		return "", fmt.Errorf("tokenMint is required")
	}

	req, err := c.newRequest(ctx, http.MethodGet,
		"token-launch/lifetime-fees?tokenMint="+url.QueryEscape(tokenMint), nil, "")
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
	if !env.Success {
		return "", fmt.Errorf("unexpected response")
	}
	return env.Response, nil
}

// -------------------- Analytics: Get Token Launch Creators --------------------

// Retrieve the creators/deployers of a specific token launch.
//
// GET /token-launch/creator/v2?tokenMint=<string>
// Authorization: x-api-key header required.
//
// Response:
//
//	{
//	  "success": true,
//	  "response": [
//	    {
//	      "username": "<string>",
//	      "pfp": "<string>",
//	      "twitterUsername": "<string>",
//	      "royaltyBps": 123,
//	      "isCreator": true,
//	      "wallet": "<string>"
//	    }
//	  ]
//	}
func (c *BagsClient) GetTokenLaunchCreators(ctx context.Context, tokenMint string) ([]TokenCreator, error) {
	if tm := strings.TrimSpace(tokenMint); tm == "" {
		return nil, fmt.Errorf("tokenMint is required")
	}

	req, err := c.newRequest(ctx, http.MethodGet,
		"token-launch/creator/v2?tokenMint="+url.QueryEscape(tokenMint), nil, "")
	if err != nil {
		return nil, err
	}

	var env struct {
		Success  bool           `json:"success"`
		Response []TokenCreator `json:"response"`
	}
	if err := c.do(req, &env); err != nil {
		return nil, err
	}
	if !env.Success {
		return nil, fmt.Errorf("unexpected response")
	}
	return env.Response, nil
}

// TokenCreator matches the "response" object in the Get Token Launch Creators call.
type TokenCreator struct {
	Username        string `json:"username"`
	Pfp             string `json:"pfp"`
	TwitterUsername string `json:"twitterUsername"`
	RoyaltyBps      int    `json:"royaltyBps"`
	IsCreator       bool   `json:"isCreator"`
	Wallet          string `json:"wallet"`
}
