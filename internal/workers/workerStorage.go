package workers

import (
	"github.com/rs/xid"
	"sync"
)

func NewWorkerStorage() *WorkerStorage {
	return &WorkerStorage{
		storage: make(map[xid.ID]*WorkerCommand),
		mu:      sync.Mutex{},
	}
}

func (s *WorkerStorage) Add(command *WorkerCommand) {
	if !s.has(command) {
		s.storage[command.id] = command
	}
}

func (s *WorkerStorage) Remove(command *WorkerCommand) {
	if s.has(command) {
		delete(s.storage, command.id)
	}
}

func (s *WorkerStorage) has(command *WorkerCommand) bool {
	_, found := s.storage[command.id]

	return found
}
