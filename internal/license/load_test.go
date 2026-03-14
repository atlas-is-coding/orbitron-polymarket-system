package license

// XDG_CACHE_HOME isolation relies on os.UserCacheDir() respecting $XDG_CACHE_HOME on Linux.
// These tests do not skip on other platforms; they simply may not isolate the real cache dir there.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLoad_EmptyToken(t *testing.T) {
	orig := rawToken
	defer func() { rawToken = orig }()
	rawToken = ""

	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	origURL := LicenseServerURL
	defer func() { LicenseServerURL = origURL }()
	LicenseServerURL = srv.URL

	creds, err := Load()
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if creds != nil {
		t.Fatalf("expected nil creds, got: %+v", creds)
	}
	if called {
		t.Fatal("expected server NOT to be called for empty token")
	}
}

func TestLoad_CacheMiss_FetchAndSave(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)

	orig := rawToken
	defer func() { rawToken = orig }()
	rawToken = encodeToken("test-token")

	requestCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		resp := BuilderCredentials{
			APIKey:    "key123",
			Secret:    "sec456",
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	origURL := LicenseServerURL
	defer func() { LicenseServerURL = origURL }()
	LicenseServerURL = srv.URL

	creds, err := Load()
	if err != nil {
		t.Fatalf("first Load() error: %v", err)
	}
	if creds == nil {
		t.Fatal("first Load() returned nil creds")
	}
	if creds.APIKey != "key123" {
		t.Fatalf("first Load() APIKey = %q, want %q", creds.APIKey, "key123")
	}

	creds2, err := Load()
	if err != nil {
		t.Fatalf("second Load() error: %v", err)
	}
	if creds2 == nil {
		t.Fatal("second Load() returned nil creds")
	}
	if creds2.APIKey != "key123" {
		t.Fatalf("second Load() APIKey = %q, want %q", creds2.APIKey, "key123")
	}

	if requestCount != 1 {
		t.Fatalf("expected exactly 1 server request, got %d (second Load should have hit cache)", requestCount)
	}
}
