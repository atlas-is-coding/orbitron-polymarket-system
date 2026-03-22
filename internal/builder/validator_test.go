// internal/builder/validator_test.go
package builder_test

import (
	"testing"
	"time"

	"github.com/atlasdev/orbitron/internal/builder"
	"github.com/atlasdev/orbitron/internal/license"
	"github.com/rs/zerolog"
)

func nopLog() zerolog.Logger { return zerolog.Nop() }

func TestValidator_NilCreds(t *testing.T) {
	v := builder.NewBuilderKeyValidator(nil, nopLog())
	r := v.Check()
	if r.Valid {
		t.Fatal("nil creds should not be valid")
	}
	if r.Reason == "" {
		t.Fatal("reason should be set")
	}
}

func TestValidator_EmptyKey(t *testing.T) {
	creds := &license.BuilderCredentials{
		APIKey:    "",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	v := builder.NewBuilderKeyValidator(creds, nopLog())
	r := v.Check()
	if r.Valid {
		t.Fatal("empty key should not be valid")
	}
}

func TestValidator_Expired(t *testing.T) {
	creds := &license.BuilderCredentials{
		APIKey:    "testkey",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	v := builder.NewBuilderKeyValidator(creds, nopLog())
	r := v.Check()
	if r.Valid {
		t.Fatal("expired key should not be valid")
	}
	if r.DaysUntilExpiry >= 0 {
		t.Fatalf("DaysUntilExpiry should be negative, got %d", r.DaysUntilExpiry)
	}
}

func TestValidator_Valid(t *testing.T) {
	creds := &license.BuilderCredentials{
		APIKey:    "testkey",
		ExpiresAt: time.Now().Add(10 * 24 * time.Hour),
	}
	v := builder.NewBuilderKeyValidator(creds, nopLog())
	r := v.Check()
	if !r.Valid {
		t.Fatalf("expected valid, got reason: %s", r.Reason)
	}
	if r.DaysUntilExpiry < 9 || r.DaysUntilExpiry > 11 {
		t.Fatalf("expected ~10 days, got %d", r.DaysUntilExpiry)
	}
}

func TestValidator_SoonExpiry(t *testing.T) {
	creds := &license.BuilderCredentials{
		APIKey:    "testkey",
		ExpiresAt: time.Now().Add(3 * 24 * time.Hour),
	}
	v := builder.NewBuilderKeyValidator(creds, nopLog())
	r := v.Check()
	if !r.Valid {
		t.Fatal("key expiring in 3 days should still be valid")
	}
	if r.DaysUntilExpiry > 7 {
		t.Fatal("should be flagged as soon-expiring")
	}
}
