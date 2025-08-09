// Package bags provides a Go client for the Bags API.
package bags

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// DefaultBaseURL is the documented base URL for the Bags API v1.
const DefaultBaseURL = "https://public-api-v2.bags.fm/api/v1/"

// UserAgentDefault is used when no custom User-Agent is provided.
const UserAgentDefault = "bags-go/0.1"

// BagsClient holds configuration for making requests to the Bags API.
type BagsClient struct {
	HTTP      *http.Client
	BaseURL   string
	APIKey    string
	UserAgent string
}

// New creates a new BagsClient with the given API key and defaults.
// The user-provided *http.Client is optional, and if nil will default to one with a 30s timeout.
func New(apiKey string, httpClient *http.Client) (*BagsClient, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("api key is required")
	}
	client := httpClient
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	return &BagsClient{
		HTTP:      client,
		BaseURL:   DefaultBaseURL,
		APIKey:    apiKey,
		UserAgent: UserAgentDefault,
	}, nil
}

// Ping sends a test request to /ping to verify API connectivity.
// It expects a JSON response: { "message": "pong" }.
func (c *BagsClient) Ping(ctx context.Context) error {
	var out struct {
		Message string `json:"message"`
	}
	if err := c.get(ctx, "/ping", &out); err != nil {
		return err
	}
	if strings.ToLower(out.Message) != "pong" {
		return fmt.Errorf("unexpected ping response: %q", out.Message)
	}
	return nil
}

// ------- Internal Helpers -------

func (c *BagsClient) get(ctx context.Context, relPath string, v any) error {
	req, err := c.newRequest(ctx, http.MethodGet, relPath, nil, "")
	if err != nil {
		return err
	}
	return c.do(req, v)
}

func (c *BagsClient) postJSON(ctx context.Context, relPath string, body any, v any) error {
	var rdr io.Reader
	if body != nil {
		buf := &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return fmt.Errorf("encode json: %w", err)
		}
		rdr = buf
	}
	req, err := c.newRequest(ctx, http.MethodPost, relPath, rdr, "application/json")
	if err != nil {
		return err
	}
	return c.do(req, v)
}

func (c *BagsClient) newRequest(ctx context.Context, method, relPath string, body io.Reader, contentType string) (*http.Request, error) {
	base, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}
	base.Path = path.Join(strings.TrimSuffix(base.Path, "/"), relPath)

	req, err := http.NewRequestWithContext(ctx, method, base.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("Accept", "application/json")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if ua := strings.TrimSpace(c.UserAgent); ua != "" {
		req.Header.Set("User-Agent", ua)
	}
	return req, nil
}
func (c *BagsClient) do(req *http.Request, v any) error {
	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var ae apiError
		data, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
		// Now checking ae.Message (field), not ae.Error (method)
		if err := json.Unmarshal(data, &ae); err == nil && (ae.Message != "" || !ae.Success) {
			if ae.Status == 0 {
				ae.Status = res.StatusCode
			}
			return &ae
		}
		bodySnippet := string(data)
		if len(bodySnippet) > 512 {
			bodySnippet = bodySnippet[:512] + "…"
		}
		return fmt.Errorf("bags api error: %s: %s", res.Status, bodySnippet)
	}

	if v != nil {
		return json.NewDecoder(res.Body).Decode(v)
	}
	_, _ = io.Copy(io.Discard, res.Body)
	return nil
}

type apiError struct {
	Success bool   `json:"success"`
	Message string `json:"error"`
	Status  int    `json:"status,omitempty"`
}

func (e *apiError) Error() string {
	status := e.Status
	if status == 0 {
		status = 400
	}
	return fmt.Sprintf("bags api error (%d): %s", status, e.Message)
}
