package workers

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/opsway/praefectus/internal/config"
	"github.com/opsway/praefectus/internal/metrics"
)

type Pool struct {
	config       *config.Config
	wsStorage    *metrics.WorkerStatStorage
	isStopping   chan struct{}
	workerStopCh map[uint32]chan struct{}
}

func NewPool(cfg *config.Config, stoppingChan chan struct{}, wsStorage *metrics.WorkerStatStorage) *Pool {
	return &Pool{
		config:       cfg,
		wsStorage:    wsStorage,
		isStopping:   stoppingChan,
		workerStopCh: map[uint32]chan struct{}{},
	}
}

func (p *Pool) Run() {
	wg := &sync.WaitGroup{}
	commands := p.config.Workers
	workerCmdCh := make(chan string, len(commands))
	for _, cmd := range commands {
		workerCmdCh <- cmd
	}

	p.workerLoop(workerCmdCh, wg)
	wg.Wait()
	log.Info("Pool: All workers stopped")
}

func (p *Pool) workerLoop(workerCmdCh chan string, wg *sync.WaitGroup) {
	var workerIdCounter uint32
	for {
		select {
		case <-p.isStopping:
			log.Debug("Pool: Got stop signal")
			for _, ch := range p.workerStopCh {
				ch <- struct{}{}
			}
			return
		case cmd := <-workerCmdCh:
			workerIdCounter++
			wg.Add(1)
			log.WithFields(log.Fields{"worker_id": workerIdCounter, "cmd": cmd}).
				Debug("Pool: Got command for new worker")
			go func(workerId uint32) {
				workerLog := log.WithField("worker_id", workerId)
				w, err := NewWorker(workerId, cmd, p.wsStorage)
				if err != nil {
					workerLog.WithError(err).
						Error("Pool: Worker init error")
					return
				}

				workerLog.Info("Pool: Starting new worker process")
				p.workerStopCh[workerId] = w.isStopping
				if err := w.Start(w.isStopping); err != nil {
					workerLog.WithError(err).
						Error("Pool: Process starting error")
					time.Sleep(3 * time.Second)
				}
				delete(p.workerStopCh, workerId)
				workerLog.Info("Pool: Finished worker process")
				wg.Done()

				workerLog.WithField("cmd", cmd).
					Debug("Pool: Restart new worker after stopping")
				workerCmdCh <- cmd
			}(workerIdCounter)
		}
	}
}
