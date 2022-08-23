package workers

import (
	"fmt"
	"github.com/opsway/praefectus/internal/config"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"path/filepath"
	"time"
)

type Liveness struct {
	pool *ScalePool
}

func newLiveness(pool *ScalePool) *Liveness {
	return &Liveness{
		pool: pool,
	}
}

func (s *Liveness) Check(request string, reply *bool) error {
	runningProcesses := RunningProcesses(s.pool.command)

	runningWorkersNumber := len(runningProcesses)
	workersIpcCheck := CheckWorkersIpcListener(runningProcesses...)
	poolActiveWorkers := s.pool.workersRegistry.activeWorkers()
	lastUpScale := s.pool.lastUpScale
	if lastUpScale == nil {
		s.pool.tryRiseWorkers()
		lastUpScale = s.pool.lastUpScale
	}
	lastDownScale := s.pool.lastDownScale
	if lastDownScale == nil {
		s.pool.tryDownscaleWorkers()
		lastDownScale = s.pool.lastDownScale
	}

	now := time.Now().Unix()

	upScaleFreeze := lastUpScale == nil || (now-lastUpScale.Timestamp) > int64(s.pool.config.ScaleTick.Seconds())
	downScaleFreeze := lastDownScale == nil || (now-lastDownScale.Timestamp) > int64(s.pool.config.DownscaleTick.Seconds())

	workersHasIdleFreeze := s.pool.checkWorkersIdleFreeze(s.pool.config.ProcessIdleSpentLimit)

	logrus.WithFields(logrus.Fields{
		"poolActiveWorkersPositive": poolActiveWorkers > 0,
		"workers":                   poolActiveWorkers <= runningWorkersNumber,
		"poolActiveWorkers":         poolActiveWorkers,
		"runningProcesses":          runningProcesses,
		"poolState":                 s.pool.state != PoolStopped,
		"wsStorage":                 s.pool.wsStorage.Length() > 0,
		"upScaleFreeze":             upScaleFreeze,
		"DownScaleFreeze":           downScaleFreeze,
		"workerHasFreeze":           workersHasIdleFreeze,
		"workerIpcCheck":            workersIpcCheck,
	}).Debug("Liveness result")

	*reply = poolActiveWorkers > 0 &&
		poolActiveWorkers <= runningWorkersNumber &&
		s.pool.state != PoolStopped &&
		s.pool.wsStorage.Length() > -1 &&
		!upScaleFreeze &&
		!downScaleFreeze &&
		workersHasIdleFreeze == false &&
		workersIpcCheck

	return nil
}

func RunningProcesses(command string) []*process.Process {
	res := make([]*process.Process, 0)
	processes, _ := process.Processes()
	logrus.WithField("number", len(processes)).Debug("Processes number")
	for _, proc := range processes {
		cmd, _ := proc.Cmdline()
		name, _ := proc.Name()
		logrus.WithField("command", cmd).Debug("Running processes")
		logrus.WithField("sprintf", fmt.Sprintf("%s %s", name, command)).Debug("Running processes")
		if cmd == command || cmd == fmt.Sprintf("%s %s", name, command) {
			res = append(res, proc)
		}
	}
	logrus.WithField("command", command).Debug("Actual process")

	return res
}

func CheckWorkersIpcListener(processes ...*process.Process) bool {
	for _, proc := range processes {
		socketPath := filepath.Join(os.TempDir(), fmt.Sprintf(config.WorkerSocketPath, proc.Pid))
		_, err := net.Dial("unix", socketPath)
		if err != nil {
			return false
		}
	}
	return true
}
