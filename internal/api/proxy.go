package api

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"

	"golang.org/x/net/proxy"

	"github.com/atlasdev/orbitron/internal/config"
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
	if cfg.Addr == "" {
		return nil, fmt.Errorf("proxy: addr must not be empty when proxy is enabled")
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
		proxyAddr := cfg.Addr
		var proxyAuth string
		if cfg.Username != "" {
			creds := base64.StdEncoding.EncodeToString([]byte(cfg.Username + ":" + cfg.Password))
			proxyAuth = "Proxy-Authorization: Basic " + creds + "\r\n"
		}
		return func(addr string) (net.Conn, error) {
			conn, err := net.Dial("tcp", proxyAddr)
			if err != nil {
				return nil, fmt.Errorf("proxy: connect to HTTP proxy %s: %w", proxyAddr, err)
			}
			_, err = fmt.Fprintf(conn, "CONNECT %s HTTP/1.1\r\nHost: %s\r\n%s\r\n", addr, addr, proxyAuth)
			if err != nil {
				conn.Close()
				return nil, fmt.Errorf("proxy: send CONNECT to %s: %w", proxyAddr, err)
			}
			br := bufio.NewReader(conn)
			resp, err := http.ReadResponse(br, nil)
			if err != nil {
				conn.Close()
				return nil, fmt.Errorf("proxy: read CONNECT response from %s: %w", proxyAddr, err)
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				conn.Close()
				return nil, fmt.Errorf("proxy: CONNECT to %s via %s failed: %s", addr, proxyAddr, resp.Status)
			}
			return conn, nil
		}, nil
	default:
		return nil, fmt.Errorf("proxy: unknown type %q (use socks5 or http)", cfg.Type)
	}
}
