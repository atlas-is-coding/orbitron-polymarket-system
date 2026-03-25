package clob

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/atlasdev/orbitron/internal/auth"
)

// APIKeyCreds — ответ на /auth/derive-api-key и /auth/api-key (POST).
type APIKeyCreds struct {
	APIKey     string `json:"apiKey"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

// DeriveAPIKey retrieves or creates L2 credentials via L1 signature.
// First attempts GET /auth/derive-api-key (nonce=0); if the key doesn't exist
// (HTTP 400), falls back to POST /auth/api-key to create a new one.
func (c *Client) DeriveAPIKey(l1 *auth.L1Signer, chainID int64) (*auth.L2Credentials, error) {
	ts, err := c.GetServerTime()
	if err != nil {
		return nil, fmt.Errorf("clob: DeriveAPIKey: get server time: %w", err)
	}

	nonce := "0"
	headers, err := l1.L1Headers(strconv.FormatInt(ts, 10), nonce, chainID)
	if err != nil {
		return nil, fmt.Errorf("clob: DeriveAPIKey: sign: %w", err)
	}

	// Try to derive an existing key first.
	resp, err := c.http.Get("/auth/derive-api-key", headers)
	if err != nil {
		return nil, fmt.Errorf("clob: DeriveAPIKey: %w", err)
	}

	if resp.StatusCode == 400 {
		// Key doesn't exist yet — create one.
		resp, err = c.http.Post("/auth/api-key", nil, headers)
		if err != nil {
			return nil, fmt.Errorf("clob: CreateAPIKey: %w", err)
		}
		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("clob: CreateAPIKey: HTTP %d: %s", resp.StatusCode, resp.Body)
		}
	} else if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("clob: DeriveAPIKey: HTTP %d: %s", resp.StatusCode, resp.Body)
	}

	var creds APIKeyCreds
	if err := json.Unmarshal(resp.Body, &creds); err != nil {
		return nil, fmt.Errorf("clob: DeriveAPIKey: decode: %w", err)
	}
	return &auth.L2Credentials{
		APIKey:     creds.APIKey,
		APISecret:  creds.Secret,
		Passphrase: creds.Passphrase,
		Address:    l1.Address(),
	}, nil
}

func (c *Client) GetServerTime() (int64, error) {
	resp, err := c.http.Get("/time", nil)
	if err != nil {
		return 0, fmt.Errorf("clob: GetServerTime: %w", err)
	}

	if resp.StatusCode >= 400 {
		return 0, fmt.Errorf("clob: GetServerTime: HTTP %d: %s", resp.StatusCode, resp.Body)
	}

	var timestamp int64
	if err := json.Unmarshal(resp.Body, &timestamp); err != nil {
		return 0, fmt.Errorf("clob: GetServerTime: decode: %w", err)
	}

	return timestamp, nil
}
