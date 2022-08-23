package timers

import (
	"context"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/opsway/praefectus/internal/config"
)

type Timers struct {
	config     *config.Config
	isStopping chan struct{}
	lastRun    *config.LastRunSeconds
}

func New(cfg *config.Config, isStopping chan struct{}) *Timers {
	return &Timers{
		config:     cfg,
		isStopping: isStopping,
	}
}

func (t *Timers) Start() {
	chunks := strings.Split(t.config.Timer.Command, " ")
	if chunks[0] == "" {
		log.Error("Ticker error: command is required")
		return
	}

	tc := time.NewTicker(time.Duration(t.config.Timer.Frequency) * time.Second)
	defer tc.Stop()

	log.WithFields(log.Fields{"cmd": t.config.Timer.Command, "interval": t.config.Timer.Frequency}).
		Debug("Ticker: Start command")

	go t.listenIpcCall()

	for {
		select {
		case <-t.isStopping:
			return
		case <-tc.C:
			log.WithField("cmd", t.config.Timer.Command).Debug("Ticker: Run command")
			go t.runCommand()
			t.lastRun = &config.LastRunSeconds{Timestamp: time.Now().Unix()}
		}
	}
}

func (t *Timers) runCommand() {
	chunks := strings.Split(t.config.Timer.Command, " ")
	ctx := context.Background()

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), time.Duration(t.config.Timer.Frequency)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, chunks[0], chunks[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.WithError(err).Error("Ticker error: failed to run command")
	}
}

func (t *Timers) listenIpcCall() {
	sockAddr := filepath.Join(os.TempDir(), config.TimersIpcSocketPath)
	log.Debug("Timer::ipc listen:", sockAddr)
	if err := os.RemoveAll(sockAddr); err != nil {
		log.WithField("Timer::ipc", "").Fatal(err)
	}
	wg := &sync.WaitGroup{}

	rpcHandler := newLiveness(t)

	rpcServer := rpc.NewServer()
	rpcServer.Register(rpcHandler)

	listener, err := net.Listen("unix", sockAddr)
	if err != nil {
		log.Fatal("Timers::ipc listen error:", err)
	}
	wg.Add(1)
	go func(wGroup *sync.WaitGroup) {
		rpcServer.Accept(listener)
		listener.Close()
		wGroup.Done()
	}(wg)
	wg.Wait()
}
