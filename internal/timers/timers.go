package timers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
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
	wg := &sync.WaitGroup{}
	for _, timerCfg := range t.config.Timers {
		wg.Add(1)
		go func(cfg config.TimersConfig, isStopping chan struct{}) {
			defer wg.Done()
			chunks := strings.Split(cfg.Command, " ")
			if chunks[0] == "" {
				fmt.Printf("Ticker error: command is required\n")
				return
			}

			tc := time.NewTicker(time.Duration(cfg.Frequency) * time.Second)
			for _ = range tc.C {
				fmt.Printf("Ticker! %s\n", cfg.Command)

				command := exec.Command(chunks[0], chunks[1:]...)
				command.Stdout = os.Stdout
				command.Stderr = os.Stderr
				if err := command.Run(); err != nil {
					fmt.Printf("Ticker Error: %s\n", err)
				}
			}
		}(timerCfg, t.isStopping)
	}
	wg.Wait()
}
