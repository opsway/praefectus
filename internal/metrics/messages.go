package metrics

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
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
	Name       string
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
	ttl     time.Duration
	mu      sync.Mutex
}

func NewQueueMessageStorage() *QueueMessageStorage {
	return &QueueMessageStorage{
		storage: make(map[string]*QueueMessage),
		ttl:     15 * time.Minute,
		mu:      sync.Mutex{},
	}
}

func (qmStorage *QueueMessageStorage) Add(id, name, transport, bus string) *QueueMessage {
	qmStorage.mu.Lock()
	defer qmStorage.mu.Unlock()
	qm := &QueueMessage{
		ID:        id,
		Name:      name,
		Transport: transport,
		Bus:       bus,
		State:     MessageStateProcessing,
		StartedAt: time.Now(),
	}
	qmStorage.storage[id] = qm

	// Remove outdated items
	for idx, item := range qmStorage.storage {
		if !item.FinishedAt.IsZero() && time.Now().After(item.FinishedAt.Add(qmStorage.ttl)) {
			log.Debugf("Remove outdated item: %s", idx)
			delete(qmStorage.storage, idx)
		}
	}

	return qm
}

func (qmStorage *QueueMessageStorage) Get(id string) *QueueMessage {
	if qmStorage.Has(id) {
		qmStorage.mu.Lock()
		defer qmStorage.mu.Unlock()

		return qmStorage.storage[id]
	}

	return nil
}

func (qmStorage *QueueMessageStorage) Has(id string) bool {
	qmStorage.mu.Lock()
	defer qmStorage.mu.Unlock()
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
	qmStorage.mu.Lock()
	defer qmStorage.mu.Unlock()
	for _, qm := range qmStorage.storage {
		if qm.State == state && qm.FinishedAt.After(afterTime) {
			count++
		}
	}

	return count
}

func (qmStorage *QueueMessageStorage) GetProcessedAfter(afterTime time.Time) []*QueueMessage {
	result := make([]*QueueMessage, 0, 10)
	qmStorage.mu.Lock()
	defer qmStorage.mu.Unlock()
	for _, qm := range qmStorage.storage {
		if (qm.State == MessageStateSucceed || qm.State == MessageStateFailed) && qm.FinishedAt.After(afterTime) {
			result = append(result, qm)
		}
	}

	return result
}
