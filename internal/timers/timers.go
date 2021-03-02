package timers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/boodmo/praefectus/internal/config"
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
		fmt.Printf("Ticker error: command is required\n")
		return
	}

	tc := time.NewTicker(time.Duration(t.config.Timer.Frequency) * time.Second)
	for range tc.C {
		fmt.Printf("Ticker! %s\n", t.config.Timer.Command)

		command := exec.Command(chunks[0], chunks[1:]...)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		if err := command.Run(); err != nil {
			fmt.Printf("Ticker Error: %s\n", err)
		}
	}
}
