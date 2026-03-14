package license

import (
	"testing"
	"time"
)

func TestCacheRoundtrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)

	creds := &BuilderCredentials{APIKey: "k", Secret: "s", ExpiresAt: time.Now().Add(time.Hour)}
	if err := saveCache("test-token", creds); err != nil {
		t.Fatal(err)
	}
	got, err := loadCache("test-token")
	if err != nil {
		t.Fatal(err)
	}
	if got.APIKey != "k" || got.Secret != "s" {
		t.Fatalf("unexpected: %+v", got)
	}
}

func TestCacheExpired(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)

	creds := &BuilderCredentials{APIKey: "k", ExpiresAt: time.Now().Add(-time.Hour)}
	_ = saveCache("test-token", creds)
	_, err := loadCache("test-token")
	if err == nil {
		t.Fatal("expected error for expired cache")
	}
}
