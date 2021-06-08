package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/opsway/praefectus/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show information about version, commit and build time",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Version:    ", version.Version)
		fmt.Println("Commit:     ", version.Commit)
		fmt.Println("Build time: ", version.BuildTime)
	},
}
