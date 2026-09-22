package events

import (
	"container/list"
	"sync"
	"time"
)

type Event struct {
	ID, Action, Name, Image, Host string
	Labels                        map[string]string
	Timestamp                     time.Time
}

var allowed = map[string]bool{"create": true, "start": true, "stop": true, "die": true, "kill": true, "pause": true, "unpause": true, "restart": true, "destroy": true, "oom": true, "health_status": true, "rename": true, "update": true, "attach": true, "detach": true}

func Accepted(event Event) bool { return allowed[event.Action] }
func (e Event) Key() string {
	return e.ID + "|" + e.Action + "|" + e.Timestamp.UTC().Format(time.RFC3339Nano)
}

type Deduplicator struct {
	mu    sync.Mutex
	limit int
	items map[string]*list.Element
	order *list.List
}

func NewDeduplicator(limit int) *Deduplicator {
	return &Deduplicator{limit: limit, items: map[string]*list.Element{}, order: list.New()}
}
func (d *Deduplicator) Seen(event Event) bool {
	key := event.Key()
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.items[key]; ok {
		return true
	}
	d.items[key] = d.order.PushBack(key)
	if d.order.Len() > d.limit {
		old := d.order.Front()
		delete(d.items, old.Value.(string))
		d.order.Remove(old)
	}
	return false
}

type Cursor struct {
	mu    sync.Mutex
	since time.Time
}

func (c *Cursor) Since() time.Time { c.mu.Lock(); defer c.mu.Unlock(); return c.since }
func (c *Cursor) Advance(value time.Time) {
	c.mu.Lock()
	if value.After(c.since) {
		c.since = value
	}
	c.mu.Unlock()
}
