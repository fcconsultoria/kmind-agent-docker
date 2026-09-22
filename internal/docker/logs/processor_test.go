package logs

import (
	"testing"
	"time"
)

func TestProcessorRedactsAndLimits(t *testing.T) {
	p, err := New(100, nil)
	if err != nil {
		t.Fatal(err)
	}
	line, ok, redacted := p.Process("api", "Authorization: Bearer secret-token\n", 100, 100, time.Now())
	if !ok || !redacted || line != "[REDACTED]" {
		t.Fatalf("unsafe output: %q", line)
	}
	_, ok, _ = p.Process("api", "0123456789", 15, 100, time.Now())
	if ok {
		t.Fatal("global limit must discard logs")
	}
}
