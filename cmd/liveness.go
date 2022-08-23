package main

import (
	"fmt"
	"github.com/opsway/praefectus/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/rpc"
	"os"
	"path/filepath"
	"time"
)

var livenessCmd = &cobra.Command{
	Use:   "liveness",
	Short: "Check scalePool liveness",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.SetupLivenessProbeConfig(flagPoolNumber)
		sockets := make([]string, 0)
		for i := 1; i <= cfg.PoolNumber; i++ {
			sockAddr := filepath.Join(os.TempDir(), fmt.Sprintf(cfg.IpcSocketPath, i))
			sockets = append(sockets, sockAddr)
		}
		sockets = append(sockets, filepath.Join(os.TempDir(), config.TimersIpcSocketPath))

		for _, socketPath := range sockets {
			reply := false
			client, err := rpc.Dial("unix", socketPath)
			if err != nil {
				log.WithField("process:", socketPath).Fatal("Liveness.Check dialing:", err)
			}

			channel := make(chan error, 1)
			go func() {
				channel <- client.Call("Liveness.Check", "", &reply)
			}()
			select {
			case err := <-channel:
				if err != nil {
					log.WithField("process:", socketPath).Fatal("Liveness.Check:", err)
				}
				log.WithField(fmt.Sprintf("Ipc call %s", socketPath), reply).Debug("Liveness probe result")
			case <-time.After(time.Second * cfg.ProcessTimeout):
				log.WithField("Liveness.Check:", socketPath).Fatal("Got timeout:", err)
			}
			if reply == false {
				os.Exit(1)
			}
		}

		os.Exit(0)
	},
}
