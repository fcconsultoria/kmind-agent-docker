package logs

import (
	"regexp"
	"strings"
	"sync"
	"time"
)

type Processor struct {
	patterns    []*regexp.Regexp
	max         int
	mu          sync.Mutex
	global, per map[string]*bucket
}
type bucket struct {
	at   time.Time
	used int
}

func New(max int, extras []string) (*Processor, error) {
	raw := append([]string{`(?i)(password|passwd|pwd)\s*[=:]\s*[^\s,;]+`, `(?i)authorization:\s*bearer\s+[^\s]+`, `(?i)(access_token|refresh_token|token)\s*[=:]\s*[^\s,;]+`, `(?i)[\w.+-]+@[\w.-]+\.[A-Za-z]{2,}`}, extras...)
	p := &Processor{max: max, global: map[string]*bucket{}, per: map[string]*bucket{}}
	for _, value := range raw {
		rule, err := regexp.Compile(value)
		if err != nil {
			return nil, err
		}
		p.patterns = append(p.patterns, rule)
	}
	return p, nil
}
func (p *Processor) Process(container, line string, limit, perLimit int, now time.Time) (string, bool, bool) {
	if len(line) > p.max {
		line = line[:p.max]
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if !allow(p.global, "all", limit, len(line), now) || !allow(p.per, container, perLimit, len(line), now) {
		return "", false, false
	}
	redacted := false
	for _, rule := range p.patterns {
		next := rule.ReplaceAllString(line, "[REDACTED]")
		redacted = redacted || next != line
		line = next
	}
	return strings.TrimRight(line, "\r\n"), true, redacted
}
func allow(items map[string]*bucket, key string, limit, size int, now time.Time) bool {
	// Bound per-container accounting even when workloads constantly recreate IDs.
	if key != "all" && len(items) >= 2048 {
		for candidate, candidateBucket := range items {
			if now.Sub(candidateBucket.at) >= time.Minute {
				delete(items, candidate)
			}
		}
		if len(items) >= 2048 {
			return false
		}
	}
	b := items[key]
	if b == nil || now.Sub(b.at) >= time.Second {
		items[key] = &bucket{at: now, used: size}
		return size <= limit
	}
	if b.used+size > limit {
		return false
	}
	b.used += size
	return true
}
