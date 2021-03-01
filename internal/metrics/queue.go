package metrics

import (
	"sync"
	"time"
)

type Queue struct {
	Transport  string
	Bus        string
	Size       uint
	LastUpdate time.Time
}

type QueueStorage struct {
	storage map[string]*Queue
	mu      sync.Mutex
}

func NewQueueStorage() *QueueStorage {
	return &QueueStorage{
		storage: make(map[string]*Queue),
		mu:      sync.Mutex{},
	}
}

func (qStorage *QueueStorage) Add(transport, bus string) *Queue {
	qStorage.mu.Lock()
	defer qStorage.mu.Unlock()

	q := &Queue{
		Transport:  transport,
		Bus:        bus,
		Size:       0,
		LastUpdate: time.Now(),
	}
	id := qStorage.generateId(transport, bus)
	qStorage.storage[id] = q

	return q
}

func (qStorage *QueueStorage) Get(transport, bus string) *Queue {
	for _, q := range qStorage.storage {
		if q.Transport == transport && q.Bus == bus {
			return q
		}
	}

	return nil
}

func (qStorage *QueueStorage) Has(transport, bus string) bool {
	id := qStorage.generateId(transport, bus)
	_, found := qStorage.storage[id]

	return found
}

func (qStorage *QueueStorage) ChangeSize(q *Queue, size uint) {
	qStorage.mu.Lock()
	defer qStorage.mu.Unlock()

	q.Size = size
	q.LastUpdate = time.Now()
}

func (qStorage *QueueStorage) GetList() map[string]*Queue {
	return qStorage.storage
}

func (qStorage *QueueStorage) generateId(transport, bus string) string {
	return transport + "|" + bus
}
