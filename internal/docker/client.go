package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// Reader deliberately exposes only Docker Engine read operations. Future code
// must not receive a mutable Docker SDK client.
type Reader interface {
	Ping(context.Context) error
	Containers(context.Context) ([]container.Summary, error)
}
type Client struct{ api *client.Client }

func New(socket string) (*Client, error) {
	api, err := client.NewClientWithOpts(client.WithHost(socket), client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("create Docker client: %w", err)
	}
	return &Client{api: api}, nil
}
func (c *Client) Ping(ctx context.Context) error { _, err := c.api.Ping(ctx); return err }
func (c *Client) Containers(ctx context.Context) ([]container.Summary, error) {
	return c.api.ContainerList(ctx, container.ListOptions{All: true})
}
