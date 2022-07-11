package workers

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/opsway/praefectus/internal/metrics"
)

type Worker struct {
	id         uint32
	command    *exec.Cmd
	wsStorage  *metrics.WorkerStatStorage
	isStopping chan struct{}
	socketPath string
}

func NewWorker(id uint32, cmd string, wsStorage *metrics.WorkerStatStorage) (*Worker, error) {
	chunks := strings.Split(cmd, " ")
	if chunks[0] == "" {
		return nil, errors.New("command is required")
	}

	command := exec.Command(chunks[0], chunks[1:]...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	return &Worker{
		id:         id,
		command:    command,
		wsStorage:  wsStorage,
		isStopping: make(chan struct{}),
	}, nil
}

func (w *Worker) Start(stopping chan struct{}) error {
	if err := w.command.Start(); err != nil {
		return err
	}

	workerLog := log.WithFields(log.Fields{"worker_id": w.id, "pid": w.command.Process.Pid})
	workerLog.Debug("Worker: Successfully started")
	wStat := w.wsStorage.Add(w.command.Process.Pid)

	go func() {
		<-stopping
		if err := w.wsStorage.ChangeState(wStat, metrics.WorkerStateStopping); err != nil {
			workerLog.WithError(err).
				Error("Worker: Changing state error")
		}

		workerLog.Debug("Worker: Sending SIGTERM")
		if err := w.command.Process.Signal(syscall.SIGTERM); err != nil {
			workerLog.WithError(err).
				Error("Worker: Sending SIGTERM error")
		}
	}()

	go func() {
		w.socketPath = filepath.Join(os.TempDir(), fmt.Sprintf("praefectus_%d.sock", w.command.Process.Pid))
		workerLog.WithField("socket", w.socketPath).
			Debug("Worker: IPC socket path")
		if err := listenUnixSocket(w.socketPath, w.isStopping); err != nil {
			workerLog.WithError(err).
				Error("Worker: IPC communication error")
		}
	}()

	if err := w.command.Wait(); err != nil {
		return err
	}

	workerLog.Debug("Worker: Successfully stopped")
	if err := w.wsStorage.ChangeState(wStat, metrics.WorkerStateStopped); err != nil {
		return err
	}

	return nil
}

func (w *Worker) StartScaleWorker(stopping chan struct{}, command *WorkerCommand) error {
	if err := w.command.Start(); err != nil {
		return err
	}

	workerLog := log.WithFields(log.Fields{"worker_id": w.id, "pid": w.command.Process.Pid})
	workerLog.Debug("Worker: Successfully started")
	wStat := w.wsStorage.Add(w.command.Process.Pid)

	command.processId = &ProcessId{id: w.command.Process.Pid}

	go func() {
		<-stopping
		workerLog.Debug("ScaleWorker: Sending SIGTERM")
		if err := w.wsStorage.ChangeState(wStat, metrics.WorkerStateStopping); err != nil {
			workerLog.WithError(err).
				Error("Worker: Changing state error")
		}

		workerLog.Debug("Worker: Sending SIGTERM")
		if err := w.command.Process.Signal(syscall.SIGTERM); err != nil {
			workerLog.WithError(err).
				Error("Worker: Sending SIGTERM error")
		}
	}()

	go func() {
		w.socketPath = filepath.Join(os.TempDir(), fmt.Sprintf("praefectus_%d.sock", w.command.Process.Pid))
		workerLog.WithField("socket", w.socketPath).
			Debug("Worker: IPC socket path")
		if err := listenUnixSocket(w.socketPath, w.isStopping); err != nil {
			workerLog.WithError(err).
				Error("Worker: IPC communication error")
		}
	}()

	if err := w.command.Wait(); err != nil {
		return err
	}

	workerLog.Debug("Worker: Successfully stopped")
	if err := w.wsStorage.ChangeState(wStat, metrics.WorkerStateStopped); err != nil {
		return err
	}

	return nil
}
