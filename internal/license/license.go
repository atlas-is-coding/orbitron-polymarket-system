package license

import "time"

const Version = "1.0.0"

// BuilderCredentials holds Polymarket Builder Program credentials.
type BuilderCredentials struct {
	APIKey    string    `json:"builder_api_key"`
	Secret    string    `json:"builder_secret"`
	ExpiresAt time.Time `json:"expires_at"`
}

// LicenseServerURL is the VPS endpoint. Override via ldflags or in tests.
var LicenseServerURL = "https://your-vps-domain.com/api/v1/license"

// Load fetches builder credentials: from cache if fresh, else from server.
// Returns nil (no error) when app token is empty — Builder features simply disabled.
func Load() (*BuilderCredentials, error) {
	token := AppToken()
	if token == "" {
		return nil, nil
	}
	if creds, err := loadCache(token); err == nil {
		return creds, nil
	}
	creds, err := fetchCredentials(LicenseServerURL, token, Version)
	if err != nil {
		return nil, err
	}
	_ = saveCache(token, creds)
	return creds, nil
}
