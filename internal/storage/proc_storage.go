package storage

import (
	"os"
	"sync"
	"time"
)

type ProcStorage struct {
	storage map[int]*WorkerStat
	mu      sync.Mutex
}

type WorkerStat struct {
	Process   *os.Process
	State     int
	UpdatedAt time.Time
}

func NewProcStorage() *ProcStorage {
	return &ProcStorage{
		storage: make(map[int]*WorkerStat),
		mu:      sync.Mutex{},
	}
}

func (ps *ProcStorage) Add(proc *os.Process) *WorkerStat {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if !ps.Has(proc.Pid) {
		ws := &WorkerStat{
			Process: proc,
			State:   StateUnknown,
		}
		ps.storage[proc.Pid] = ws
		return ws
	}

	return nil
}

func (ps *ProcStorage) Get(pid int) *WorkerStat {
	if ps.Has(pid) {
		return ps.storage[pid]
	}

	return nil
}

func (ps *ProcStorage) Has(pid int) bool {
	_, found := ps.storage[pid]

	return found
}

func (ps *ProcStorage) Remove(pid int) {
	if ps.Has(pid) {
		delete(ps.storage, pid)
	}
}

func (ps *ProcStorage) GetList() map[int]*WorkerStat {
	return ps.storage
}

func (ps *ProcStorage) ChangeState(worker *WorkerStat, state int) error {
	if w := ps.Get(worker.Process.Pid); w != nil {
		// State validation?
		w.State = state
		w.UpdatedAt = time.Now()
	}

	return nil
}
