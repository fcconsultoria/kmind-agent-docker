package buffer

import "testing"

func TestQueueBoundsMemoryAndCopiesPayload(t *testing.T) {
	queue := New(4)
	payload := []byte("abc")
	if !queue.Push(payload) || queue.Push([]byte("de")) {
		t.Fatal("queue did not enforce its byte limit")
	}
	payload[0] = 'x'
	if got := string(queue.Pop()); got != "abc" {
		t.Fatalf("queued payload = %q, want copied value", got)
	}
}
