package main

import (
	"log"

	"github.com/spf13/cobra"
)

var (
	flagWorkerPoolCmds []string
	flagWorkerNumber   uint8
	flagServerHost     string
	flagServerPort     int
	flagTimerCmd       string
	flagTimerInterval  uint16
)

func init() {
	runCmd.Flags().StringArrayVarP(&flagWorkerPoolCmds, "worker-pool-cmd", "w", []string{}, "")
	runCmd.Flags().Uint8VarP(&flagWorkerNumber, "worker-number", "n", 1, "")
	runCmd.Flags().StringVarP(&flagServerHost, "host", "", "0.0.0.0", "")
	runCmd.Flags().IntVarP(&flagServerPort, "port", "", 9000, "")
	runCmd.Flags().StringVarP(&flagTimerCmd, "timer-cmd", "", "", "")
	runCmd.Flags().Uint16VarP(&flagTimerInterval, "timer-interval", "", 60, "")
}

func main() {
	rootCmd := &cobra.Command{
		Use: "praefectus [command]",
	}

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
