package webui

import (
	"testing"
	"time"
)

func TestJWTRoundTrip(t *testing.T) {
	token, err := signJWT("user", time.Hour, "secret")
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("empty token")
	}
	if err := verifyJWT(token, "secret"); err != nil {
		t.Fatalf("verify: %v", err)
	}
}

func TestJWTWrongSecret(t *testing.T) {
	token, _ := signJWT("user", time.Hour, "secret")
	if err := verifyJWT(token, "other"); err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestJWTExpired(t *testing.T) {
	token, _ := signJWT("user", -time.Second, "secret")
	if err := verifyJWT(token, "secret"); err == nil {
		t.Fatal("expected error for expired token")
	}
}
