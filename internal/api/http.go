package api

import (
	"fmt"
	"time"

	"github.com/valyala/fasthttp"
)

// Client — базовый HTTP клиент на основе fasthttp с пулом соединений и retry.
type Client struct {
	hc         *fasthttp.Client
	baseURL    string
	maxRetries int
}

// NewClient создаёт новый HTTP клиент с заданными параметрами.
func NewClient(baseURL string, timeoutSec, maxRetries int) *Client {
	timeout := time.Duration(timeoutSec) * time.Second

	hc := &fasthttp.Client{
		ReadTimeout:              timeout,
		WriteTimeout:             timeout,
		MaxIdleConnDuration:      30 * time.Second,
		MaxConnDuration:          60 * time.Second,
		MaxConnsPerHost:          256,
		DisablePathNormalizing:   true,
		DisableHeaderNamesNormalizing: false,
		Name:                     "polytrade-bot/1.0",
	}

	return &Client{
		hc:         hc,
		baseURL:    baseURL,
		maxRetries: maxRetries,
	}
}

// Request описывает HTTP запрос.
type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    []byte
}

// Response описывает HTTP ответ.
type Response struct {
	StatusCode int
	Body       []byte
}

// Do выполняет запрос с retry-логикой (экспоненциальный backoff).
func (c *Client) Do(req *Request) (*Response, error) {
	freqReq := fasthttp.AcquireRequest()
	freqResp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(freqReq)
	defer fasthttp.ReleaseResponse(freqResp)

	freqReq.Header.SetMethod(req.Method)
	freqReq.SetRequestURI(c.baseURL + req.Path)
	freqReq.Header.Set("Content-Type", "application/json")
	freqReq.Header.Set("Accept", "application/json")

	for k, v := range req.Headers {
		freqReq.Header.Set(k, v)
	}

	if len(req.Body) > 0 {
		freqReq.SetBody(req.Body)
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// Экспоненциальный backoff: 100ms, 200ms, 400ms
			time.Sleep(time.Duration(100*(1<<attempt)) * time.Millisecond)
			freqResp.Reset()
		}

		if err := c.hc.Do(freqReq, freqResp); err != nil {
			lastErr = err
			continue
		}

		statusCode := freqResp.StatusCode()
		// Retry только на 5xx и таймауты
		if statusCode >= 500 && attempt < c.maxRetries {
			lastErr = fmt.Errorf("server error: status %d", statusCode)
			continue
		}

		body := make([]byte, len(freqResp.Body()))
		copy(body, freqResp.Body())

		return &Response{
			StatusCode: statusCode,
			Body:       body,
		}, nil
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", c.maxRetries, lastErr)
}

// Get выполняет GET-запрос.
func (c *Client) Get(path string, headers map[string]string) (*Response, error) {
	return c.Do(&Request{
		Method:  fasthttp.MethodGet,
		Path:    path,
		Headers: headers,
	})
}

// Post выполняет POST-запрос с JSON-телом.
func (c *Client) Post(path string, body []byte, headers map[string]string) (*Response, error) {
	return c.Do(&Request{
		Method:  fasthttp.MethodPost,
		Path:    path,
		Body:    body,
		Headers: headers,
	})
}

// Delete выполняет DELETE-запрос.
func (c *Client) Delete(path string, body []byte, headers map[string]string) (*Response, error) {
	return c.Do(&Request{
		Method:  fasthttp.MethodDelete,
		Path:    path,
		Body:    body,
		Headers: headers,
	})
}
