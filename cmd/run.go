package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/boodmo/praefectus/internal/config"
	"github.com/boodmo/praefectus/internal/metrics"
	"github.com/boodmo/praefectus/internal/rpc"
	"github.com/boodmo/praefectus/internal/server"
	"github.com/boodmo/praefectus/internal/signals"
	"github.com/boodmo/praefectus/internal/timers"
	"github.com/boodmo/praefectus/internal/workers"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start workers, timers and API server for metrics",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				Host: flagServerHost,
				Port: flagServerPort,
			},
			Workers: make([]config.WorkersConfig, 0, len(flagWorkerPoolCmds)),
			Timers:  make([]config.TimersConfig, 0, 1),
		}

		for _, cmd := range flagWorkerPoolCmds {
			cfg.Workers = append(cfg.Workers, config.WorkersConfig{
				Command: cmd,
				Number:  flagWorkerNumber,
			})
		}

		cfg.Timers = append(cfg.Timers, config.TimersConfig{
			Command:   flagTimerCmd,
			Frequency: flagTimerInterval,
		})

		qStorage := metrics.NewQueueStorage()
		qmStorage := metrics.NewQueueMessageStorage()
		wsStorage := metrics.NewWorkerStatStorage()

		rpcHandler := rpc.NewRPCHandler(qStorage, qmStorage, wsStorage)
		if err := rpc.Register(rpcHandler); err != nil {
			log.Fatal(err)
		}

		isStopping := make(chan struct{})
		signals.CatchSigterm(isStopping)

		m, err := metrics.NewMetrics(qStorage, qmStorage, wsStorage)
		if err != nil {
			log.Fatal(err)
		}
		go m.Start()

		apiServer := server.New(cfg, m)
		go apiServer.Start()

		t := timers.New(cfg, isStopping)
		go t.Start()

		p := workers.NewPool(cfg, isStopping, wsStorage)
		p.Run()
	},
}
