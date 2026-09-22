package transport

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Client struct {
	endpoint, apiKey string
	http             *http.Client
	mu               sync.Mutex
	failures         int
	blockedUntil     time.Time
}

func New(endpoint, apiKey string, timeout time.Duration) *Client {
	return &Client{endpoint: strings.TrimRight(endpoint, "/"), apiKey: apiKey, http: &http.Client{Timeout: timeout}}
}
func (c *Client) Send(ctx context.Context, signal string, payload []byte) error {
	if !c.allowed() {
		return fmt.Errorf("transport circuit is open")
	}
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	if _, err := writer.Write(payload); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimSuffix(c.endpoint, "/v1")+"/v1/"+signal, &compressed)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+c.apiKey)
	request.Header.Set("Content-Type", "application/x-protobuf")
	request.Header.Set("Content-Encoding", "gzip")
	response, err := c.http.Do(request)
	if err != nil {
		c.failed()
		return err
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		c.failed()
		return fmt.Errorf("OTLP endpoint returned HTTP %d", response.StatusCode)
	}
	c.success()
	return nil
}
func (c *Client) allowed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return !time.Now().Before(c.blockedUntil)
}
func (c *Client) failed() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failures++
	if c.failures >= 3 {
		delay := time.Duration(1<<min(c.failures-3, 5)) * time.Second
		c.blockedUntil = time.Now().Add(delay)
	}
}
func (c *Client) success() { c.mu.Lock(); c.failures = 0; c.blockedUntil = time.Time{}; c.mu.Unlock() }
func min(left, right int) int {
	if left < right {
		return left
	}
	return right
}
