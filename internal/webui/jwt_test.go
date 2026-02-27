package webui

import (
	"errors"
	"testing"
	"time"
)

func TestJWTRoundTrip(t *testing.T) {
	token, err := signJWT("user", time.Hour, "secret")
	if err != nil {
		t.Fatal(err)
	}
	sub, err := verifyJWT(token, "secret")
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if sub != "user" {
		t.Fatalf("expected subject 'user', got %q", sub)
	}
}

func TestJWTWrongSecret(t *testing.T) {
	token, _ := signJWT("user", time.Hour, "secret")
	_, err := verifyJWT(token, "other")
	if !errors.Is(err, errTokenInvalid) {
		t.Fatalf("expected errTokenInvalid, got %v", err)
	}
}

func TestJWTExpired(t *testing.T) {
	token, _ := signJWT("user", -time.Second, "secret")
	_, err := verifyJWT(token, "secret")
	if !errors.Is(err, errTokenExpired) {
		t.Fatalf("expected errTokenExpired, got %v", err)
	}
}
