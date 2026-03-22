// internal/builder/validator.go
package builder

import (
	"time"

	"github.com/atlasdev/orbitron/internal/license"
	"github.com/rs/zerolog"
)

// ValidationResult holds the result of a builder key validation check.
type ValidationResult struct {
	Valid           bool
	DaysUntilExpiry int    // negative if expired
	Reason          string // empty if valid
}

// BuilderKeyValidator checks that builder credentials are present and not expired.
type BuilderKeyValidator struct {
	creds  *license.BuilderCredentials
	logger zerolog.Logger
}

// NewBuilderKeyValidator creates a validator. creds may be nil (no license token configured).
func NewBuilderKeyValidator(creds *license.BuilderCredentials, log zerolog.Logger) *BuilderKeyValidator {
	return &BuilderKeyValidator{creds: creds, logger: log}
}

// Check validates the credentials and logs the result. Always non-fatal.
func (v *BuilderKeyValidator) Check() ValidationResult {
	if v.creds == nil {
		v.logger.Debug().Msg("builder: no license token configured — builder attribution disabled")
		return ValidationResult{Reason: "no license token configured"}
	}
	if v.creds.APIKey == "" {
		v.logger.Error().Msg("builder: API key is empty — orders will NOT be attributed")
		return ValidationResult{Reason: "API key is empty"}
	}

	now := time.Now()
	days := int(v.creds.ExpiresAt.Sub(now).Hours() / 24)

	if now.After(v.creds.ExpiresAt) {
		// Ensure days is negative even when expired < 24h ago (int truncation gives 0).
		if days >= 0 {
			days = -1
		}
		v.logger.Error().
			Int("days_expired", -days).
			Msg("builder: API key has EXPIRED — orders are NOT attributed")
		return ValidationResult{DaysUntilExpiry: days, Reason: "key expired"}
	}

	if days < 7 {
		v.logger.Warn().
			Int("days_until_expiry", days).
			Msg("builder: API key expiring soon — renew via Polymarket")
	} else {
		v.logger.Info().
			Int("days_until_expiry", days).
			Str("key_prefix", v.creds.APIKey[:min(4, len(v.creds.APIKey))]+"***").
			Msg("builder: API key valid")
	}

	return ValidationResult{Valid: true, DaysUntilExpiry: days}
}
