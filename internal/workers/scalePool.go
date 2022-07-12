package workers

import (
	"fmt"
	"github.com/opsway/praefectus/internal/config"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/opsway/praefectus/internal/metrics"
)

const (
	PoolRunning PoolState = iota
	PoolStopped
)

type PoolState uint8

type ScalePool struct {
	command         string
	wsStorage       *metrics.WorkerStatStorage
	isStopping      chan struct{}
	workerRunCh     chan *WorkerCommand
	workersRegistry *WorkerStorage
	state           PoolState
	id              int
	childStopChan   map[int]chan struct{}
	config          *config.ScalePoolConfig
}

func NewScalePoolRange(commandConfig *config.Config, stoppingChan chan struct{}, wsStorage *metrics.WorkerStatStorage) []*ScalePool {
	pool := make([]*ScalePool, 0, len(commandConfig.Workers))
	channelPool := make([]chan struct{}, 0, len(commandConfig.Workers))
	poolId := 1
	for _, command := range commandConfig.Workers {
		childChan := make(chan struct{})
		channelPool = append(channelPool, childChan)
		poolConfig := config.SetupPoolConfig()
		pool = append(pool, NewScalePool(command, childChan, wsStorage, poolConfig, poolId))
		poolId++
	}
	go func() {
		select {
		case <-stoppingChan:
			for _, childChannel := range channelPool {
				childChannel <- struct{}{}
			}
		}
	}()

	return pool
}

func NewScalePool(command string, stoppingChan chan struct{}, wsStorage *metrics.WorkerStatStorage, poolConfig *config.ScalePoolConfig, id int) *ScalePool {
	return &ScalePool{
		command:         command,
		wsStorage:       wsStorage,
		isStopping:      stoppingChan,
		workerRunCh:     make(chan *WorkerCommand, poolConfig.MaxProcessPullSize),
		workersRegistry: NewWorkerStorage(),
		state:           PoolRunning,
		id:              id,
		childStopChan:   map[int]chan struct{}{},
		config:          poolConfig,
	}
}

func RunScalePoolRange(poolRange []*ScalePool) {
	wg := &sync.WaitGroup{}
	for _, p := range poolRange {
		wg.Add(1)
		pool := p
		go func() {
			defer wg.Done()
			pool.Run()
		}()
	}
	wg.Wait()
}

func (p *ScalePool) Run() {
	log.WithField("id", p.id).Info(fmt.Sprintf("ScalePool: Start workers of %s", p.command))
	wg := &sync.WaitGroup{}
	p.addCommand(1)
	p.scale(wg)
	p.catchStopSignal(wg)
	p.workerLoop(wg)
	wg.Wait()
	for _, childChan := range p.childStopChan {
		close(childChan)
	}
	close(p.workerRunCh)
	log.WithField("id", p.id).Info(fmt.Sprintf("ScalePool: All workers of %s stopped", p.command))
}

func (p *ScalePool) workerLoop(wg *sync.WaitGroup) {
	loopStopChan := make(chan struct{})
	p.childStopChan[len(p.childStopChan)+1] = loopStopChan
	var workerIdCounter uint32
	for {
		select {
		case <-loopStopChan:
			log.WithField("id", p.id).Debug(fmt.Sprintf("ScalePool: loop sigterm %s", p.command))
			for _, command := range p.workersRegistry.storage {
				command.stop <- struct{}{}
				command.state = Remove
			}
			return
		case cmd := <-p.workerRunCh:
			workerIdCounter++
			if p.state == PoolStopped || cmd.state == Remove {
				return
			}
			wg.Add(1)
			log.WithFields(log.Fields{"worker_id": workerIdCounter, "cmd": cmd}).Debug("ScalePool: Got command for new worker")
			go func(workerId uint32) {
				defer wg.Done()
				workerLog := log.WithField("worker_id", workerId)
				w, err := NewWorker(workerId, cmd.command, p.wsStorage)
				if err != nil {
					workerLog.WithError(err).Error("ScalePool: Worker init error")
					return
				}

				workerLog.WithField("id", p.id).Info("ScalePool: Starting new worker process")

				if err := w.StartScaleWorker(cmd.stop, cmd); err != nil {
					workerLog.WithError(err).Error("ScalePool: Process starting error")
					time.Sleep(3 * time.Second)
				}
				workerLog.WithField("id", p.id).Info("ScalePool: Finished worker process")

				if cmd.state == Remove || p.state == PoolStopped {
					workerLog.WithField("id", p.id).Debug("Pool: remove worker")
					return
				}

				workerLog.WithField("id", p.id).Debug("ScalePool: Restart new worker after stopping")

				p.wsStorage.Remove(cmd.processId.id)

				cmd.processId = nil

				p.workerRunCh <- cmd

			}(workerIdCounter)
		}
	}
}

func (p *ScalePool) addCommand(num uint8) {
	if p.state == PoolStopped {
		return
	}

	if num+uint8(len(p.workersRegistry.storage)) > p.config.MaxProcessPullSize {
		num = p.config.MaxProcessPullSize - uint8(len(p.workersRegistry.storage))
	}
	if num == 0 {
		return
	}
	log.WithField("id", p.id).Debug(fmt.Sprintf("ScalePool: Got signal to rise %s workers", string(num)))

	p.workersRegistry.mu.Lock()
	defer p.workersRegistry.mu.Unlock()

	for i := uint8(0); i < num; i++ {
		command := NewWorkerCommand(p.command)
		p.workersRegistry.Add(command)
		p.workerRunCh <- command
	}
}

func (p *ScalePool) catchStopSignal(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-p.isStopping:
				log.WithField("id", p.id).Debug("Scale pool: got SIGTERM")
				p.state = PoolStopped
				for _, childChan := range p.childStopChan {
					childChan <- struct{}{}
				}
				return
			}
		}
	}()
}
