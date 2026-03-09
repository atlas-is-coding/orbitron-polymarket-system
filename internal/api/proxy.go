package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"golang.org/x/net/proxy"

	"github.com/atlasdev/polytrade-bot/internal/config"
)

// DialFunc is a function compatible with fasthttp.Client.Dial and
// websocket.Dialer.NetDial. Returns nil when proxy is disabled.
type DialFunc func(addr string) (net.Conn, error)

// BuildDialer returns a DialFunc for the given proxy config.
// Returns nil, nil when proxy is disabled — callers use default behaviour.
func BuildDialer(cfg config.ProxyConfig) (DialFunc, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	switch cfg.Type {
	case "socks5":
		var auth *proxy.Auth
		if cfg.Username != "" {
			auth = &proxy.Auth{User: cfg.Username, Password: cfg.Password}
		}
		dialer, err := proxy.SOCKS5("tcp", cfg.Addr, auth, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("proxy: build SOCKS5 dialer for %s: %w", cfg.Addr, err)
		}
		return func(addr string) (net.Conn, error) {
			return dialer.Dial("tcp", addr)
		}, nil
	case "http":
		proxyURL, err := url.Parse("http://" + cfg.Addr)
		if err != nil {
			return nil, fmt.Errorf("proxy: parse HTTP proxy addr %q: %w", cfg.Addr, err)
		}
		if cfg.Username != "" {
			proxyURL.User = url.UserPassword(cfg.Username, cfg.Password)
		}
		transport := &http.Transport{Proxy: http.ProxyURL(proxyURL)}
		return func(addr string) (net.Conn, error) {
			return transport.DialContext(context.Background(), "tcp", addr)
		}, nil
	default:
		return nil, fmt.Errorf("proxy: unknown type %q (use socks5 or http)", cfg.Type)
	}
}
