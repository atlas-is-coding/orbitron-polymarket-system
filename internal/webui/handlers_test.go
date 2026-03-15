package webui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/atlasdev/orbitron/internal/config"
)

func makeTestServer(t *testing.T) *Server {
	t.Helper()
	cfg := &config.Config{}
	cfg.WebUI.JWTSecret = "test-secret"
	cfg.WebUI.Listen = "127.0.0.1:0"
	return &Server{
		cfg:      cfg,
		cfgPath:  "",
		password: "correct",
		state:    newWebState(),
		hub:      newHub(),
	}
}

func loginToken(t *testing.T, s *Server) string {
	t.Helper()
	body := `{"password":"correct"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleLogin(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login failed: %d %s", w.Code, w.Body)
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	return resp["token"]
}

func TestLoginWrongPassword(t *testing.T) {
	s := makeTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/login",
		strings.NewReader(`{"password":"wrong"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleLogin(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestLoginSuccess(t *testing.T) {
	s := makeTestServer(t)
	token := loginToken(t, s)
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestGetOverviewRequiresAuth(t *testing.T) {
	s := makeTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/overview", nil)
	w := httptest.NewRecorder()
	s.jwtMiddleware(s.handleOverview)(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestGetOverviewWithAuth(t *testing.T) {
	s := makeTestServer(t)
	token, _ := signJWT("admin", time.Hour, s.cfg.WebUI.JWTSecret)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/overview", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	s.jwtMiddleware(s.handleOverview)(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
