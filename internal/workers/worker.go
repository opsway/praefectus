package workers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/boodmo/praefectus/internal/storage"
)

type Worker struct {
	command      *exec.Cmd
	stateStorage *storage.ProcStorage
	isStopping   chan struct{}
	socketPath   string
}

func NewWorker(ps *storage.ProcStorage) *Worker {
	cmd := exec.Command("bin/app", "messenger:consume", "messenger.transport.amqp") // ToDo: pass via args
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return &Worker{
		command:      cmd,
		stateStorage: ps,
		isStopping:   make(chan struct{}),
	}
}

func (w *Worker) Start(stopping chan struct{}) error {
	if err := w.command.Start(); err != nil {
		return err
	}
	fmt.Printf("Worker started [PID: %d]\n", w.command.Process.Pid)
	wStat := w.stateStorage.Add(w.command.Process)
	if err := w.stateStorage.ChangeState(wStat, storage.StateStarting); err != nil {
		return err
	}

	go func() {
		<-stopping
		if err := w.stateStorage.ChangeState(wStat, storage.StateStopping); err != nil {
			fmt.Printf("Error Change State [PID: %d] %s\n", w.command.Process.Pid, err)
		}
		fmt.Printf("Send SIGTERM to worker [PID: %d]\n", w.command.Process.Pid)
		if err := w.command.Process.Signal(syscall.SIGTERM); err != nil { // ToDo: Handle error?
			fmt.Printf("Error SIGTERM [PID: %d] %s\n", w.command.Process.Pid, err)
		}
	}()

	go func() {
		w.socketPath = filepath.Join(os.TempDir(), fmt.Sprintf("praefectus_%d.sock", w.PID()))
		fmt.Printf("  Unix socket: %s\n", w.socketPath)
		if err := listenUnixSocket(w.socketPath, w.isStopping); err != nil {
			fmt.Printf("Err: %+s\n", err)
		}
	}()

	if err := w.command.Wait(); err != nil {
		return err
	}

	fmt.Printf("Worker stopped [PID: %d]\n", w.command.Process.Pid)
	if err := w.stateStorage.ChangeState(wStat, storage.StateStopped); err != nil {
		return err
	}

	return nil
}

func (w *Worker) PID() int {
	return w.command.Process.Pid
}
