package metrics

import (
	"sync"
	"time"
)

// WorkerState enum
const (
	WorkerStateUnknown WorkerState = iota
	WorkerStateStarting
	WorkerStateStarted
	WorkerStateIdle
	WorkerStateBusy
	WorkerStateStopping
	WorkerStateStopped
)

type WorkerState uint8

func (s WorkerState) String() string {
	return [...]string{"Unknown", "Starting", "Started", "Idle", "Busy", "Stopping", "Stopped"}[s]
}

// WorkerStat
type WorkerStat struct {
	PID        int
	Transport  string
	Bus        string
	State      WorkerState
	StartedAt  time.Time
	FinishedAt time.Time
}

// WorkerStatStorage
type WorkerStatStorage struct {
	storage map[int]*WorkerStat
	mu      sync.Mutex
}

func NewWorkerStatStorage() *WorkerStatStorage {
	return &WorkerStatStorage{
		storage: make(map[int]*WorkerStat),
		mu:      sync.Mutex{},
	}
}

func (wsStorage *WorkerStatStorage) Add(pid int) *WorkerStat {
	wsStorage.mu.Lock()
	defer wsStorage.mu.Unlock()

	if !wsStorage.Has(pid) {
		wStat := &WorkerStat{
			PID:       pid,
			State:     WorkerStateStarting,
			StartedAt: time.Now(),
		}
		wsStorage.storage[pid] = wStat

		return wStat
	}

	return nil
}

func (wsStorage *WorkerStatStorage) Get(pid int) *WorkerStat {
	if wsStorage.Has(pid) {
		return wsStorage.storage[pid]
	}

	return nil
}

func (wsStorage *WorkerStatStorage) Has(pid int) bool {
	_, found := wsStorage.storage[pid]

	return found
}

func (wsStorage *WorkerStatStorage) ChangeState(worker *WorkerStat, state WorkerState) error {
	if w := wsStorage.Get(worker.PID); w != nil {
		wsStorage.mu.Lock()
		defer wsStorage.mu.Unlock()

		// State validation?
		w.State = state
	}

	return nil
}

func (wsStorage *WorkerStatStorage) CountByState(state WorkerState) float64 {
	var count float64
	for _, ws := range wsStorage.storage {
		if ws.State == state {
			count++
		}
	}

	return count
}
