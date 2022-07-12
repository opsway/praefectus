package timers

import (
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/opsway/praefectus/internal/config"
)

type Timers struct {
	config     *config.Config
	isStopping chan struct{}
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

	for {
		select {
		case <-t.isStopping:
			return
		case <-tc.C:
			log.WithField("cmd", t.config.Timer.Command).Debug("Ticker: Run command")
			command := exec.Command(chunks[0], chunks[1:]...)
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
			if err := command.Run(); err != nil {
				log.WithError(err).Error("Ticker error: failed to run command")
			}
		}
	}
}
