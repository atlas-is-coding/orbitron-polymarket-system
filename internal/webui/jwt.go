package webui

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	errTokenInvalid = errors.New("invalid token")
	errTokenExpired = errors.New("token expired")
)

func b64enc(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func b64dec(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

func signJWT(subject string, ttl time.Duration, secret string) (string, error) {
	header := b64enc([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload, err := json.Marshal(map[string]any{
		"sub": subject,
		"exp": time.Now().Add(ttl).Unix(),
	})
	if err != nil {
		return "", err
	}
	body := header + "." + b64enc(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	return body + "." + b64enc(mac.Sum(nil)), nil
}

func verifyJWT(token, secret string) error {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return errTokenInvalid
	}
	body := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	expected := b64enc(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return errTokenInvalid
	}
	raw, err := b64dec(parts[1])
	if err != nil {
		return fmt.Errorf("%w: payload decode", errTokenInvalid)
	}
	var claims map[string]any
	if err := json.Unmarshal(raw, &claims); err != nil {
		return fmt.Errorf("%w: payload json", errTokenInvalid)
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		return fmt.Errorf("%w: missing exp", errTokenInvalid)
	}
	if time.Now().Unix() > int64(exp) {
		return errTokenExpired
	}
	return nil
}
