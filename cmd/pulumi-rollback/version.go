// Copyright 2024 Pegasus Heavy Industries LLC
// Contact: pegasusheavyindustries@gmail.com

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is set during build via ldflags
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("pulumi-rollback %s\n", Version)
		fmt.Printf("  Git commit: %s\n", GitCommit)
		fmt.Printf("  Build date: %s\n", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
