package workers

import (
	"fmt"
	"sync"
	"time"

	"github.com/boodmo/praefectus/internal/config"
	"github.com/boodmo/praefectus/internal/metrics"
)

type Pool struct {
	config     *config.Config
	wsStorage  *metrics.WorkerStatStorage
	isStopping chan struct{}
}

func NewPool(cfg *config.Config, stoppingChan chan struct{}, wsStorage *metrics.WorkerStatStorage) *Pool {
	return &Pool{
		config:     cfg,
		wsStorage:  wsStorage,
		isStopping: stoppingChan,
	}
}

func (p *Pool) Run() {
	wg := &sync.WaitGroup{}
	commands := p.config.Workers
	workerChan := make(chan string, len(commands))
	for _, cmd := range commands {
		workerChan <- cmd
	}

	p.workerLoop(workerChan, wg)
	wg.Wait()
	fmt.Printf("Main: Done!\n")
}

func (p *Pool) workerLoop(workerChan chan string, wg *sync.WaitGroup) {
	var idx int
	for {
		select {
		case <-p.isStopping:
			fmt.Printf("Got stop signal\n")
			return
		case cmd := <-workerChan:
			idx++
			fmt.Printf("[WRK#%d] Got command for new worker: %s\n", idx, cmd)
			wg.Add(1)
			go func(idx int) {
				w, err := NewWorker(cmd, p.wsStorage)
				if err != nil {
					fmt.Printf("[WRK#%d] Worker init error: %s\n", idx, err)
					return
				}
				fmt.Printf("[WRK#%d] Starting new worker process\n", idx)
				if err := w.Start(w.isStopping); err != nil {
					fmt.Printf("[WRK#%d] Process error: %s\n", idx, err)
					time.Sleep(3 * time.Second)
				}
				fmt.Printf("[WRK#%d] Done!\n", idx)
				wg.Done()

				fmt.Printf("[WRK#%d] Start new worker after stopping\n", idx)
				workerChan <- cmd
			}(idx)
		}
	}
}
