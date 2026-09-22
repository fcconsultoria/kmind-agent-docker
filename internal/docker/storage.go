package docker

import (
	"sync"
	"time"
)

// ImageMetadata is deliberately kept out of Prometheus labels. Image IDs and
// digests are high-cardinality identity data and belong to OTLP attributes.
type ImageMetadata struct {
	ID, Repository, Tag, Digest, Architecture, OS string
	SizeBytes                                     int64
	CollectedAt                                   time.Time
}
type ImageCache struct {
	mu     sync.Mutex
	values map[string]ImageMetadata
	limit  int
}

func NewImageCache(limit int) *ImageCache {
	return &ImageCache{values: map[string]ImageMetadata{}, limit: limit}
}
func (c *ImageCache) Get(id string, now time.Time, refresh time.Duration) (ImageMetadata, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok := c.values[id]
	return value, ok && now.Sub(value.CollectedAt) < refresh
}
func (c *ImageCache) Put(value ImageMetadata) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if value.ID == "" {
		return
	}
	if len(c.values) >= c.limit {
		var oldest string
		var at time.Time
		for id, item := range c.values {
			if oldest == "" || item.CollectedAt.Before(at) {
				oldest = id
				at = item.CollectedAt
			}
		}
		delete(c.values, oldest)
	}
	c.values[value.ID] = value
}

// ContainerStorage is collected at a slow interval because SizeRw/SizeRootFs
// can make Docker calculate filesystem usage. Negative values mean unavailable.
type ContainerStorage struct {
	SizeRW, SizeRootFS int64
	CollectedAt        time.Time
}

// VolumeInventory never claims capacity from a Docker volume unless the driver
// reports it explicitly. Most drivers expose no reliable per-volume capacity.
type VolumeInventory struct {
	Name, Driver, Scope string
	Labels              map[string]string
	Mountpoint          string
	CapacityStatus      string
	CapacityBytes       int64
}

func UnavailableVolume(name, driver, scope string, labels map[string]string) VolumeInventory {
	return VolumeInventory{Name: name, Driver: driver, Scope: scope, Labels: labels, CapacityStatus: "unavailable", CapacityBytes: -1}
}
