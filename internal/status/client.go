package status

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type State struct {
	Active     bool
	CheckAfter time.Duration
}
type Client struct {
	endpoint, apiKey string
	http             *http.Client
}

func New(endpoint, apiKey string, timeout time.Duration) *Client {
	return &Client{endpoint: endpoint, apiKey: apiKey, http: &http.Client{Timeout: timeout}}
}
func (c *Client) Check(ctx context.Context) (State, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint, nil)
	if err != nil {
		return State{}, err
	}
	request.Header.Set("Authorization", "Bearer "+c.apiKey)
	response, err := c.http.Do(request)
	if err != nil {
		return State{}, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return State{}, fmt.Errorf("status endpoint returned HTTP %d", response.StatusCode)
	}
	var payload struct {
		Active            bool `json:"active"`
		CheckAfterSeconds int  `json:"checkAfterSeconds"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(nil, response.Body, 8192)).Decode(&payload); err != nil {
		return State{}, err
	}
	after := time.Duration(payload.CheckAfterSeconds) * time.Second
	if after < 15*time.Second {
		after = 15 * time.Second
	}
	if after > 10*time.Minute {
		after = 10 * time.Minute
	}
	return State{Active: payload.Active, CheckAfter: after}, nil
}
