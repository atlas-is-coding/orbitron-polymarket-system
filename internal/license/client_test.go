package license

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchCredentials(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		json.NewEncoder(w).Encode(BuilderCredentials{
			APIKey:    "test-key",
			Secret:    "test-secret",
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		})
	}))
	defer srv.Close()

	creds, err := fetchCredentials(srv.URL, "any-token", "0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	if creds.APIKey != "test-key" {
		t.Fatalf("unexpected key: %s", creds.APIKey)
	}
}

func TestFetchCredentials_InvalidToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid_token"}`))
	}))
	defer srv.Close()

	_, err := fetchCredentials(srv.URL, "bad-token", "0.0.1")
	if err == nil {
		t.Fatal("expected error for 401")
	}
}
