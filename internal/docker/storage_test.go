package docker

import (
	"testing"
	"time"
)

func TestImageCacheExpiresAndBoundsEntries(t *testing.T) {
	now := time.Now()
	cache := NewImageCache(1)
	cache.Put(ImageMetadata{ID: "a", CollectedAt: now})
	if _, ok := cache.Get("a", now, time.Minute); !ok {
		t.Fatal("expected cache hit")
	}
	if _, ok := cache.Get("a", now.Add(2*time.Minute), time.Minute); ok {
		t.Fatal("expected expired entry")
	}
	cache.Put(ImageMetadata{ID: "b", CollectedAt: now})
	if _, ok := cache.Get("a", now, time.Hour); ok {
		t.Fatal("cache must be bounded")
	}
}
func TestUnavailableVolumeNeverInventsCapacity(t *testing.T) {
	volume := UnavailableVolume("data", "local", "local", nil)
	if volume.CapacityStatus != "unavailable" || volume.CapacityBytes >= 0 {
		t.Fatal("unavailable volume must not report capacity")
	}
}
