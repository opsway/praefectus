package workers

import (
	"fmt"
	"sync"

	"github.com/boodmo/praefectus/internal/storage"
)

type Pool struct {
	workers      []*Worker
	workerNum    int
	isStopping   chan struct{}
	stateStorage *storage.ProcStorage
}

func NewPool(num int, stoppingChan chan struct{}, ps *storage.ProcStorage) *Pool {
	return &Pool{
		workers:      make([]*Worker, 0, num),
		workerNum:    num,
		isStopping:   stoppingChan,
		stateStorage: ps,
	}
}

func (p *Pool) Run() {
	wg := &sync.WaitGroup{}
	workerChan := make(chan int, p.workerNum)
	for i := 0; i < p.workerNum; i++ {
		workerChan <- i + 1
	}

	p.workerLoop(workerChan, wg)
	wg.Wait()
	fmt.Printf("Main: Done!\n")
}

func (p *Pool) workerLoop(workerChan chan int, wg *sync.WaitGroup) {
	for {
		select {
		case <-p.isStopping:
			fmt.Printf("Got stop signal\n")
			return
		case idx := <-workerChan:
			fmt.Printf("[WRK#%d] Got worker queue\n", idx)
			wg.Add(1)
			go func(idx int) {
				w := NewWorker(p.stateStorage)
				fmt.Printf("[WRK#%d] Starting new worker process\n", idx)
				if err := w.Start(w.isStopping); err != nil {
					// ToDo: Start retry limit
					fmt.Printf("[WRK#%d] Process error: %s\n", idx, err)
				}
				fmt.Printf("[WRK#%d] Done!\n", idx)
				wg.Done()

				fmt.Printf("[WRK#%d] Start new worker after stopping\n", idx)
				workerChan <- idx
			}(idx)
		}
	}
}
