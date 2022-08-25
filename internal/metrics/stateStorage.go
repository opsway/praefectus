package metrics

import (
	"math"
	"sync"
	"time"
)

type secondsInState struct {
	spend int
}

type StateStorage struct {
	storage map[int]*timestampWorkerState
	mu      sync.Mutex
}

func newStateStorage() *StateStorage {
	return &StateStorage{
		storage: make(map[int]*timestampWorkerState),
		mu:      sync.Mutex{},
	}
}

func (s *StateStorage) add(status WorkerState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()
	if s.has(len(s.storage)) {
		previousState := s.storage[len(s.storage)]
		// milliseconds
		spent := int(now - previousState.timestamp)

		previousState.secondsInState = &secondsInState{spend: spent}
	}

	s.storage[len(s.storage)+1] = &timestampWorkerState{state: status, timestamp: now, secondsInState: nil}
}

func (s *StateStorage) WorkerStatePercentage(state WorkerState, start int64, end int64) uint8 {
	s.mu.Lock()
	defer s.mu.Unlock()
	spentTime := 0
	for _, timeWorkerState := range s.storage {
		if timeWorkerState.state != state {
			continue
		}

		switch true {
		// in case process still running and belong to current period
		case isCurrentStateInterval(timeWorkerState, end):
			if start > timeWorkerState.timestamp {
				spentTime += int(end - start)
				break
			}
			spentTime += int(end - timeWorkerState.timestamp)
			break
		// in case process partially was running in a period
		case isPartialRunningInInterval(timeWorkerState, start, end):
			spentTime += int(timeWorkerState.timestamp-start) + timeWorkerState.secondsInState.spend
			break
		case timeWorkerState.timestamp < start:
		case timeWorkerState.timestamp > end:
		case timeWorkerState.secondsInState == nil:
			break
		default:
			spentTime += timeWorkerState.secondsInState.spend
		}
	}
	result := math.Round(float64(spentTime) / float64(end-start) * 100)

	return uint8(result)
}

// CheckIdleFreeze - started more than timeSpentLimit time ago and is still active (secondsInState == nil)
func (s *StateStorage) CheckIdleFreeze(timeSpentLimit time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UnixMilli()
	for _, timeWorkerState := range s.storage {
		if timeWorkerState.state != WorkerStateIdle {
			continue
		}
		if now-timeWorkerState.timestamp > int64(timeSpentLimit*time.Millisecond) && timeWorkerState.secondsInState == nil {
			return true
		}
	}

	return false
}

func (s *StateStorage) has(pid int) bool {
	_, found := s.storage[pid]

	return found
}

func isPartialRunningInInterval(workerState *timestampWorkerState, start int64, end int64) bool {
	return workerState.secondsInState != nil &&
		workerState.timestamp+int64(workerState.secondsInState.spend) > start &&
		workerState.timestamp+int64(workerState.secondsInState.spend) <= end &&
		workerState.timestamp < start
}

func isCurrentStateInterval(workerState *timestampWorkerState, end int64) bool {
	return workerState.secondsInState == nil &&
		workerState.timestamp < end
}
