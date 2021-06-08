package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/opsway/praefectus/internal/config"
	"github.com/opsway/praefectus/internal/metrics"
	"github.com/opsway/praefectus/internal/rpc"
	"github.com/opsway/praefectus/internal/server"
	"github.com/opsway/praefectus/internal/signals"
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
			Timer: config.TimerConfig{
				Command:   flagTimerCmd,
				Frequency: flagTimerInterval,
			},
		}
		cfg.Workers = append(cfg.Workers, flagWorkerPoolCmds...)

		//err := errors.New("Something went wrong")
		////log.WithError(err).Error("Ticker error: failed to run command")
		//newLog := log.WithFields(log.Fields{"cmd": "asdasd", "interval": 123})
		//newLog.
		//	WithFields(log.Fields{"alt": "zxczxcz", "isOK": true}).
		//	WithError(err).
		//	Debug("Ticker: Start command")
		//log.Info("!!!")
		//return

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
