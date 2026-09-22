package buffer

import "sync"

// Queue is bounded in memory; telemetry is intentionally dropped instead of persisted.
type Queue struct {
	mu          sync.Mutex
	limit, used int
	values      [][]byte
}

func New(limit int) *Queue { return &Queue{limit: limit} }
func (q *Queue) Push(value []byte) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(value) > q.limit-q.used {
		return false
	}
	copyValue := append([]byte(nil), value...)
	q.values = append(q.values, copyValue)
	q.used += len(copyValue)
	return true
}
func (q *Queue) Pop() []byte {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.values) == 0 {
		return nil
	}
	value := q.values[0]
	q.values = q.values[1:]
	q.used -= len(value)
	return value
}
func (q *Queue) Clear() { q.mu.Lock(); q.values = nil; q.used = 0; q.mu.Unlock() }
