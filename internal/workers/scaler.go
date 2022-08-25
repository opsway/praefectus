package workers

import (
	"github.com/opsway/praefectus/internal/config"
	"github.com/opsway/praefectus/internal/metrics"
	log "github.com/sirupsen/logrus"
	"math"
	"sync"
	"time"
)

type WorkerBusynessRange uint8

func (p *ScalePool) scale(wg *sync.WaitGroup) {
	scaleTickerStopChan := make(chan struct{})
	downscaleTickerStopChan := make(chan struct{})
	p.childStopChan[len(p.childStopChan)+1] = scaleTickerStopChan
	p.childStopChan[len(p.childStopChan)+1] = downscaleTickerStopChan

	// scale
	wg.Add(1)
	go func() {
		upScale := time.NewTicker(p.config.ScaleTick * time.Second)
		defer func() {
			upScale.Stop()
			p.lastUpScale = nil
			wg.Done()
		}()
		for {
			select {
			case <-scaleTickerStopChan:
				return
			case <-upScale.C:
				if p.state == PoolStopped {
					return
				}
				p.tryRiseWorkers()
			}
		}
	}()

	//downscale
	wg.Add(1)
	go func() {
		downScale := time.NewTicker(p.config.DownscaleTick * time.Second)
		defer func() {
			downScale.Stop()
			p.lastDownScale = nil
			wg.Done()
		}()
		for {
			select {
			case <-downscaleTickerStopChan:
				return
			case <-downScale.C:
				if p.state == PoolStopped {
					return
				}
				p.tryDownscaleWorkers()
			}
		}
	}()
}

func (p *ScalePool) tryDownscaleWorkers() {
	if p.state == PoolStopped {
		return
	}
	p.lastDownScale = &config.LastRunSeconds{Timestamp: time.Now().Unix()}
	workerPercentage, activeWorkers, idleWorkers := p.processIdleWorkers()
	if workerPercentage == 0 && idleWorkers == nil {
		log.Warning("processIdleWorkers inappropriate result on active pool")
		return
	}

	removedWorkers := p.removeIdleWorkers(idleWorkers...)
	log.WithFields(log.Fields{"pool": p.id, "idle": workerPercentage, "remove": removedWorkers, "activeWorkers": activeWorkers}).
		Debug("PScale pool: pool total idle %")
}

func (p *ScalePool) tryRiseWorkers() {
	if p.state == PoolStopped {
		return
	}

	p.lastUpScale = &config.LastRunSeconds{Timestamp: time.Now().Unix()}
	workersBusinessPercentage := p.calculateWorkersBusiness()
	workersIncrease := uint8(0)
	switch true {
	case workersBusinessPercentage <= p.config.WorkerBusynessLow:
		return
	case workersBusinessPercentage <= p.config.WorkerBusynessAverage:
		workersIncrease = p.config.WorkerNumberLowIncrease
		break
	case workersBusinessPercentage < p.config.WorkerBusynessHigh:
		workersIncrease = p.config.WorkerNumberAverageIncrease
		break
	default:
		workersIncrease = p.config.WorkerNumberHighIncrease
	}
	p.addCommand(workersIncrease)
}

func (p *ScalePool) calculateWorkersBusiness() uint8 {
	if p.state == PoolStopped {
		return 0
	}
	p.workersRegistry.mu.Lock()
	defer p.workersRegistry.mu.Unlock()

	activeWorkers := uint8(len(p.workersRegistry.storage))
	workersBusyness := 0
	end := time.Now().UnixMilli()
	start := end - int64(p.config.ScaleTick*1000)
	for _, command := range p.workersRegistry.storage {
		if command.processId == nil {
			activeWorkers--
			continue
		}
		workerStats := p.wsStorage.Get(command.processId.id)
		if workerStats == nil {
			activeWorkers--
			continue
		}
		res := workerStats.StateStorage.WorkerStatePercentage(metrics.WorkerStateBusy, start, end)
		workersBusyness += int(res)
		log.WithFields(log.Fields{"pool": p.id, "business": res}).Debug("Scale pool: Worker business %")
	}

	workerPercentage := math.Round(float64(workersBusyness) / float64(activeWorkers))
	log.WithFields(log.Fields{"pool": p.id, "business": workerPercentage, "activeWorkers": activeWorkers}).Debug("Scale pool: pool total business %")

	return uint8(workerPercentage)
}

func (p *ScalePool) processIdleWorkers() (int, int, []*WorkerCommand) {
	if p.state == PoolStopped {
		return 0, 0, nil
	}
	p.workersRegistry.mu.Lock()
	defer p.workersRegistry.mu.Unlock()

	totalIdlePercentage := 0
	idleWorkers := make([]*WorkerCommand, 0)
	activeWorkers := len(p.workersRegistry.storage)
	// append idleWorkers, mark fresh if command is busy
	for _, command := range p.workersRegistry.storage {
		if command.processId == nil {
			activeWorkers--
			continue
		}
		workerStats := p.wsStorage.Get(command.processId.id)
		if workerStats == nil {
			activeWorkers--
			continue
		}
		now := time.Now().UnixMilli()

		workerIdleStatePercentage := workerStats.StateStorage.WorkerStatePercentage(metrics.WorkerStateIdle, now-int64(p.config.ScaleTick*1000), now)
		totalIdlePercentage += int(workerIdleStatePercentage)

		log.WithFields(log.Fields{"pool": p.id, "idle": workerIdleStatePercentage}).Debug("Scale pool: Worker idle %")
		if workerIdleStatePercentage >= p.config.WorkerIdlePercentageLimit {
			idleWorkers = append(idleWorkers, command)
		}
		if workerIdleStatePercentage < p.config.WorkerIdlePercentageLimit {
			command.state = Fresh
		}
	}
	workerPercentage := totalIdlePercentage / activeWorkers

	return workerPercentage, activeWorkers, idleWorkers
}

func (p *ScalePool) removeIdleWorkers(commands ...*WorkerCommand) int {
	if p.state == PoolStopped {
		return 0
	}
	p.workersRegistry.mu.Lock()
	defer p.workersRegistry.mu.Unlock()

	removedWorkers := 0
	idleWorkersNum := len(commands)
	for _, command := range commands {
		switch true {
		case command.state == Fresh:
			command.state = MarkRemove
			break
		case command.state == MarkRemove:
			removedWorkers++
			if idleWorkersNum == 1 {
				break
			}
			command.state = Remove
			command.stop <- struct{}{}
			p.workersRegistry.Remove(command)
			break
		}
		idleWorkersNum--
	}

	return removedWorkers
}
