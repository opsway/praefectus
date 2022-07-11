package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	flagWorkerPoolCmds []string
	flagServerHost     string
	flagServerPort     int
	flagTimerCmd       string
	flagTimerInterval  uint16
	flagVerbose        bool
	flagMode           string
)

func init() {
	runCmd.Flags().StringArrayVarP(&flagWorkerPoolCmds, "worker-pool-cmd", "w", []string{}, "Commands for running as worker")
	runCmd.Flags().StringVarP(&flagServerHost, "host", "", "0.0.0.0", "Server listening host")
	runCmd.Flags().IntVarP(&flagServerPort, "port", "", 9000, "Server listening port")
	runCmd.Flags().StringVarP(&flagTimerCmd, "timer-cmd", "", "", "Command for running by timer")
	runCmd.Flags().Uint16VarP(&flagTimerInterval, "timer-interval", "", 60, "Interval of timer")
	runCmd.Flags().BoolVarP(&flagVerbose, "verbose", "v", false, "Show extra debug info")
	runCmd.Flags().StringVarP(&flagMode, "mode", "m", "static", "Pool scale mode")
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	rootCmd := &cobra.Command{
		Use: "praefectus [command]",
	}

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
