package clob

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/atlasdev/orbitron/internal/api"
	"github.com/atlasdev/orbitron/internal/auth"
)

// Client — CLOB API клиент.
type Client struct {
	http  *api.Client
	creds *auth.L2Credentials
}

// NewClient создаёт CLOB-клиент.
// creds может быть nil для публичных эндпоинтов.
func NewClient(httpClient *api.Client, creds *auth.L2Credentials) *Client {
	return &Client{http: httpClient, creds: creds}
}

// --- Вспомогательные методы ---

func (c *Client) publicGet(path string) (*api.Response, error) {
	return c.http.Get(path, nil)
}

func (c *Client) privateGet(path string) (*api.Response, error) {
	if c.creds == nil {
		return nil, fmt.Errorf("clob: L2 credentials required")
	}
	headers, err := c.creds.L2Headers(http.MethodGet, path, "")
	if err != nil {
		return nil, err
	}
	return c.http.Get(path, headers)
}

func (c *Client) privatePost(path string, body []byte) (*api.Response, error) {
	if c.creds == nil {
		return nil, fmt.Errorf("clob: L2 credentials required")
	}
	headers, err := c.creds.L2Headers(http.MethodPost, path, string(body))
	if err != nil {
		return nil, err
	}
	return c.http.Post(path, body, headers)
}

func (c *Client) privateDelete(path string, body []byte) (*api.Response, error) {
	if c.creds == nil {
		return nil, fmt.Errorf("clob: L2 credentials required")
	}
	bodyStr := ""
	if len(body) > 0 {
		bodyStr = string(body)
	}
	headers, err := c.creds.L2Headers(http.MethodDelete, path, bodyStr)
	if err != nil {
		return nil, err
	}
	return c.http.Delete(path, body, headers)
}

func decode[T any](resp *api.Response) (*T, error) {
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("clob: HTTP %d: %s", resp.StatusCode, string(resp.Body))
	}
	var result T
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("clob: decode response: %w", err)
	}
	return &result, nil
}
