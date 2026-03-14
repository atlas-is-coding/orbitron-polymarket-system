package license

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type licenseRequest struct {
	Token   string `json:"token"`
	Version string `json:"version"`
}

func fetchCredentials(serverURL, token, version string) (*BuilderCredentials, error) {
	body, _ := json.Marshal(licenseRequest{Token: token, Version: version})
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(serverURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("license: request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("license: invalid app token (401)")
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("license: rate limited (429)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("license: server error (%d)", resp.StatusCode)
	}
	var creds BuilderCredentials
	if err := json.NewDecoder(resp.Body).Decode(&creds); err != nil {
		return nil, fmt.Errorf("license: decode response: %w", err)
	}
	return &creds, nil
}
