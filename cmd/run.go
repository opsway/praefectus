package main

import (
	"github.com/opsway/praefectus/internal/signals"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/opsway/praefectus/internal/config"
	"github.com/opsway/praefectus/internal/metrics"
	"github.com/opsway/praefectus/internal/rpc"
	"github.com/opsway/praefectus/internal/server"
	"github.com/opsway/praefectus/internal/timers"
	"github.com/opsway/praefectus/internal/workers"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start workers, timers and API server for metrics",
	Run: func(cmd *cobra.Command, args []string) {
		if flagVerbose {
			log.SetLevel(log.DebugLevel) // ToDo: Move to global flag
		}

		cfg := &config.Config{
			Server: config.ServerConfig{
				Host: flagServerHost,
				Port: flagServerPort,
			},
			Workers: make([]string, 0, len(flagWorkerPoolCmds)),
			Timer:   config.SetupTimersConfig(flagTimerCmd, flagTimerInterval),
		}
		cfg.Workers = append(cfg.Workers, flagWorkerPoolCmds...)

		qStorage := metrics.NewQueueStorage()
		qmStorage := metrics.NewQueueMessageStorage()
		wsStorage := metrics.NewWorkerStatStorage()

		rpcHandler := rpc.NewRPCHandler(qStorage, qmStorage, wsStorage)
		if err := rpc.Register(rpcHandler); err != nil {
			log.Fatal(err)
		}

		channelPool := make([]chan struct{}, 0, 2)

		tickerIsStopping := make(chan struct{})
		poolIsStopping := make(chan struct{})

		channelPool = append(channelPool, tickerIsStopping, poolIsStopping)

		m, err := metrics.NewMetrics(qStorage, qmStorage, wsStorage)
		if err != nil {
			log.Fatal(err)
		}
		go m.Start()

		apiServer := server.New(cfg, m)
		go apiServer.Start()

		t := timers.New(cfg, tickerIsStopping)
		go t.Start()

		switch flagMode {
		case "static":
			p := workers.NewPool(cfg, poolIsStopping, wsStorage)
			p.Run()
			break
		case "scale":
			poolRange := workers.NewScalePoolRange(cfg, poolIsStopping, wsStorage)
			workers.RunScalePoolRange(poolRange)
			break
		default:
			log.Fatal("Unknown scale-mode")
		}

		signals.CatchSigterm(channelPool)
	},
}
