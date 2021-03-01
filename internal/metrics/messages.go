package metrics

import (
	"sync"
	"time"
)

// QueueMessageState enum
const (
	MessageStateUnknown QueueMessageState = iota
	MessageStateProcessing
	MessageStateSucceed
	MessageStateFailed
)

type QueueMessageState uint8

func (s QueueMessageState) String() string {
	return [...]string{"Unknown", "Processing", "Succeed", "Failed"}[s]
}

// QueueMessage
type QueueMessage struct {
	ID         string
	Transport  string
	Bus        string
	State      QueueMessageState
	StartedAt  time.Time
	FinishedAt time.Time
}

func (qm *QueueMessage) GetProcessedTime() time.Duration {
	// ToDo: Check state
	return qm.FinishedAt.Sub(qm.StartedAt)
}

// QueueMessageStorage
// ToDo: Flush map by TTL
type QueueMessageStorage struct {
	storage map[string]*QueueMessage
	mu      sync.Mutex
}

func NewQueueMessageStorage() *QueueMessageStorage {
	return &QueueMessageStorage{
		storage: make(map[string]*QueueMessage),
		mu:      sync.Mutex{},
	}
}

func (qmStorage *QueueMessageStorage) Add(id, transport, bus string) *QueueMessage {
	qmStorage.mu.Lock()
	defer qmStorage.mu.Unlock()
	qm := &QueueMessage{
		ID:        id,
		Transport: transport,
		Bus:       bus,
		State:     MessageStateProcessing,
		StartedAt: time.Now(),
	}
	qmStorage.storage[id] = qm

	return qm
}

func (qmStorage *QueueMessageStorage) Get(id string) *QueueMessage {
	if qmStorage.Has(id) {
		return qmStorage.storage[id]
	}

	return nil
}

func (qmStorage *QueueMessageStorage) Has(id string) bool {
	_, found := qmStorage.storage[id]

	return found
}

func (qmStorage *QueueMessageStorage) ChangeState(qm *QueueMessage, state QueueMessageState) error {
	if qmStorage.Has(qm.ID) {
		qmStorage.mu.Lock()
		defer qmStorage.mu.Unlock()

		// State validation?
		qm.State = state
		qm.FinishedAt = time.Now()
	}

	return nil
}

func (qmStorage *QueueMessageStorage) CountByState(state QueueMessageState, afterTime time.Time) float64 {
	var count float64
	for _, qm := range qmStorage.storage {
		if qm.State == state && qm.FinishedAt.After(afterTime) {
			count++
		}
	}

	return count
}

func (qmStorage *QueueMessageStorage) GetProcessedAfter(afterTime time.Time) []*QueueMessage {
	result := make([]*QueueMessage, 0, 10)
	for _, qm := range qmStorage.storage {
		if (qm.State == MessageStateSucceed || qm.State == MessageStateFailed) && qm.FinishedAt.After(afterTime) {
			result = append(result, qm)
		}
	}

	return result
}
