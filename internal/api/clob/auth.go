package clob

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/atlasdev/polytrade-bot/internal/auth"
)

// APIKeyCreds — ответ на /auth/derive-api-key и /auth/api-key (POST).
type APIKeyCreds struct {
	APIKey     string `json:"apiKey"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

// DeriveAPIKey получает существующие L2 credentials через L1 подпись.
// Вызывает GET /auth/derive-api-key с L1 заголовками (nonce=0).
func (c *Client) DeriveAPIKey(l1 *auth.L1Signer) (*auth.L2Credentials, error) {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := "0"
	headers, err := l1.L1Headers(ts, nonce)
	if err != nil {
		return nil, fmt.Errorf("clob: DeriveAPIKey: sign: %w", err)
	}
	resp, err := c.http.Get("/auth/derive-api-key", headers)
	if err != nil {
		return nil, fmt.Errorf("clob: DeriveAPIKey: %w", err)
	}
	if resp.StatusCode >= 400 {
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
