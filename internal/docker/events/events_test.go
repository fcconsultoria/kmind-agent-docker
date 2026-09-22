package events

import (
	"testing"
	"time"
)

func TestDeduplicatesAndBounds(t *testing.T) {
	d := NewDeduplicator(1)
	e := Event{ID: "a", Action: "start", Timestamp: time.Now()}
	if d.Seen(e) || !d.Seen(e) {
		t.Fatal("event dedup failed")
	}
	d.Seen(Event{ID: "b", Action: "stop", Timestamp: time.Now()})
	if d.Seen(e) {
		t.Fatal("old event must expire from bounded cache")
	}
}
func TestAcceptedActions(t *testing.T) {
	if !Accepted(Event{Action: "oom"}) || Accepted(Event{Action: "exec_start"}) {
		t.Fatal("unexpected event allowlist")
	}
}
