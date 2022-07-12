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

type timestampWorkerState struct {
	state          WorkerState
	timestamp      int64
	secondsInState *secondsInState
}

// WorkerStat
type WorkerStat struct {
	PID          int
	Transport    string
	Bus          string
	State        WorkerState
	StartedAt    time.Time
	FinishedAt   time.Time
	StateStorage *StateStorage
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

	if !wsStorage.find(pid) {
		wsStorage.mu.Lock()
		defer wsStorage.mu.Unlock()
		wStat := &WorkerStat{
			PID:          pid,
			State:        WorkerStateStarting,
			StartedAt:    time.Now(),
			StateStorage: newStateStorage(),
		}
		wStat.StateStorage.add(WorkerStateStarting)
		wsStorage.storage[pid] = wStat

		return wStat
	}

	return nil
}

func (wsStorage *WorkerStatStorage) Get(pid int) *WorkerStat {
	if wsStorage.find(pid) {
		wsStorage.mu.Lock()
		defer wsStorage.mu.Unlock()

		return wsStorage.storage[pid]
	}

	return nil
}

func (wsStorage *WorkerStatStorage) Remove(pid int) {
	if !wsStorage.find(pid) {
		return
	}
	wsStorage.mu.Lock()
	defer wsStorage.mu.Unlock()

	delete(wsStorage.storage, pid)
}

func (wsStorage *WorkerStatStorage) Has(pid int) bool {
	wsStorage.mu.Lock()
	defer wsStorage.mu.Unlock()

	return wsStorage.find(pid)
}

func (wsStorage *WorkerStatStorage) find(pid int) bool {
	_, found := wsStorage.storage[pid]

	return found
}

func (wsStorage *WorkerStatStorage) ChangeState(worker *WorkerStat, state WorkerState) error {
	if w := wsStorage.Get(worker.PID); w != nil {
		wsStorage.mu.Lock()
		defer wsStorage.mu.Unlock()

		// State validation?
		w.State = state
		w.StateStorage.add(state)
	}

	return nil
}

func (wsStorage *WorkerStatStorage) CountByState(state WorkerState) float64 {
	wsStorage.mu.Lock()
	defer wsStorage.mu.Unlock()

	var count float64
	for _, ws := range wsStorage.storage {
		if ws.State == state {
			count++
		}
	}

	return count
}
