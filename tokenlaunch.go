// tokenlaunch.go
package bags

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
)

// -------------------- Types (from docs) --------------------

// CreateTokenInfoRequest carries token metadata and the image to upload.
// Docs fields: name, symbol, description, telegram, twitter, website, image (file).
// POST https://public-api-v2.bags.fm/api/v1/token-launch/create-token-info (multipart/form-data)
// Auth header: x-api-key
// Ref: https://bags.mintlify.app/api-reference/create-token-info
type CreateTokenInfoRequest struct {
	Name        string
	Symbol      string
	Description string
	Telegram    string
	Twitter     string
	Website     string

	// Image is required; filename is sent in Content-Disposition.
	Image         io.Reader
	ImageFilename string
	ImageMIMEType string // optional; defaults to application/octet-stream when empty
}

type CreateTokenInfoResult struct {
	TokenMint     string         `json:"tokenMint"`
	TokenMetadata string         `json:"tokenMetadata"`
	TokenLaunch   TokenLaunchObj `json:"tokenLaunch"`
}

type TokenLaunchObj struct {
	UserID       string `json:"userId"`
	Name         string `json:"name"`
	Symbol       string `json:"symbol"`
	Description  string `json:"description"`
	Telegram     string `json:"telegram"`
	Twitter      string `json:"twitter"`
	Website      string `json:"website"`
	Image        string `json:"image"`
	TokenMint    string `json:"tokenMint"`
	Status       string `json:"status"` // e.g., "PRE_LAUNCH"
	LaunchWallet string `json:"launchWallet"`
	LaunchSig    string `json:"launchSignature"`
	URI          string `json:"uri"`
	CreatedAtISO string `json:"createdAt"`
	UpdatedAtISO string `json:"updatedAt"`
}

// CreateTokenLaunchConfigRequest/Result for config creation.
// POST https://public-api-v2.bags.fm/api/v1/token-launch/create-config (application/json)
// Auth header: x-api-key
// Ref: https://bags.mintlify.app/api-reference/create-token-launch-configuration
type CreateTokenLaunchConfigRequest struct {
	LaunchWallet string `json:"launchWallet"`
}
type CreateTokenLaunchConfigResult struct {
	Tx        string `json:"tx"`
	ConfigKey string `json:"configKey"`
}

// CreateTokenLaunchTxRequest/Result for final transaction.
// POST https://public-api-v2.bags.fm/api/v1/token-launch/create-launch-transaction (application/json)
// Auth header: x-api-key
// Ref: https://bags.mintlify.app/api-reference/create-token-launch-transaction
type CreateTokenLaunchTxRequest struct {
	IPFS               string `json:"ipfs"`
	TokenMint          string `json:"tokenMint"`
	Wallet             string `json:"wallet"`
	InitialBuyLamports int64  `json:"initialBuyLamports"`
	ConfigKey          string `json:"configKey"`
}
type CreateTokenLaunchTxResult struct {
	Transaction string // "response" is a plain string (base64 tx)
}

// -------------------- Methods --------------------

// CreateTokenInfoAndMetadata uploads metadata + image and returns created info.
// Endpoint: POST token-launch/create-token-info (multipart/form-data)
func (c *BagsClient) CreateTokenInfoAndMetadata(ctx context.Context, in *CreateTokenInfoRequest) (*CreateTokenInfoResult, error) {
	if in == nil {
		return nil, fmt.Errorf("nil request")
	}
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Symbol) == "" {
		return nil, fmt.Errorf("name and symbol are required")
	}
	if in.Image == nil || strings.TrimSpace(in.ImageFilename) == "" {
		return nil, fmt.Errorf("image and image filename are required")
	}

	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)

	// stream multipart body
	go func() {
		defer mw.Close()
		defer pw.Close()

		writeField := func(k, v string) error {
			if strings.TrimSpace(v) == "" {
				return nil
			}
			return mw.WriteField(k, v)
		}
		_ = writeField("name", in.Name)
		_ = writeField("symbol", in.Symbol)
		_ = writeField("description", in.Description)
		_ = writeField("telegram", in.Telegram)
		_ = writeField("twitter", in.Twitter)
		_ = writeField("website", in.Website)

		ctype := in.ImageMIMEType
		if strings.TrimSpace(ctype) == "" {
			ctype = "application/octet-stream"
		}
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="image"; filename="%s"`, in.ImageFilename))
		h.Set("Content-Type", ctype)

		part, err := mw.CreatePart(h)
		if err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		if _, err := io.Copy(part, in.Image); err != nil {
			_ = pw.CloseWithError(err)
			return
		}
	}()

	// IMPORTANT: path is relative (no leading slash) to avoid clobbering BaseURL path.
	req, err := c.newRequest(ctx, http.MethodPost, "token-launch/create-token-info", pr, mw.FormDataContentType())
	if err != nil {
		return nil, err
	}

	var env struct {
		Success  bool                   `json:"success"`
		Response *CreateTokenInfoResult `json:"response"`
	}
	if err := c.do(req, &env); err != nil {
		return nil, err
	}
	if !env.Success || env.Response == nil {
		return nil, fmt.Errorf("unexpected response")
	}
	return env.Response, nil
}

// CreateTokenLaunchConfig creates the config-creation transaction for a wallet.
// Endpoint: POST token-launch/create-config (application/json)
func (c *BagsClient) CreateTokenLaunchConfig(ctx context.Context, in *CreateTokenLaunchConfigRequest) (*CreateTokenLaunchConfigResult, error) {
	if in == nil || strings.TrimSpace(in.LaunchWallet) == "" {
		return nil, fmt.Errorf("launchWallet is required")
	}
	var env struct {
		Success  bool                           `json:"success"`
		Response *CreateTokenLaunchConfigResult `json:"response"`
	}
	if err := c.postJSON(ctx, "token-launch/create-config", in, &env); err != nil {
		return nil, err
	}
	if !env.Success || env.Response == nil {
		return nil, fmt.Errorf("unexpected response")
	}
	return env.Response, nil
}

// CreateTokenLaunchTransaction builds the final launch transaction (signed with token mint).
// Endpoint: POST token-launch/create-launch-transaction (application/json)
func (c *BagsClient) CreateTokenLaunchTransaction(ctx context.Context, in *CreateTokenLaunchTxRequest) (*CreateTokenLaunchTxResult, error) {
	if in == nil {
		return nil, fmt.Errorf("nil request")
	}
	if strings.TrimSpace(in.IPFS) == "" ||
		strings.TrimSpace(in.TokenMint) == "" ||
		strings.TrimSpace(in.Wallet) == "" ||
		strings.TrimSpace(in.ConfigKey) == "" {
		return nil, fmt.Errorf("ipfs, tokenMint, wallet, and configKey are required")
	}

	var env struct {
		Success  bool   `json:"success"`
		Response string `json:"response"`
	}
	if err := c.postJSON(ctx, "token-launch/create-launch-transaction", in, &env); err != nil {
		return nil, err
	}
	if !env.Success || strings.TrimSpace(env.Response) == "" {
		return nil, fmt.Errorf("unexpected response")
	}
	return &CreateTokenLaunchTxResult{Transaction: env.Response}, nil
}
