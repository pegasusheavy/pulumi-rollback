// Copyright 2026 Pegasus Heavy Industries LLC
// Contact: pegasusheavyindustries@gmail.com

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	stackName   string
	projectPath string
	verbose     bool
)

var rootCmd = &cobra.Command{
	Use:   "pulumi-rollback",
	Short: "Roll back Pulumi deployments to previous states",
	Long: `pulumi-rollback is a CLI tool that allows you to roll back Pulumi stack
deployments to previous states from the deployment history.

It works with all Pulumi backends including Pulumi Cloud, S3, Azure Blob,
Google Cloud Storage, and local filesystem.

Examples:
  # List deployment history for a stack
  pulumi-rollback list --stack mystack

  # Preview what would change when rolling back
  pulumi-rollback preview --stack mystack --version 5

  # Roll back to a specific version
  pulumi-rollback to --stack mystack --version 5`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&stackName, "stack", "s", "", "Name of the Pulumi stack")
	rootCmd.PersistentFlags().StringVarP(&projectPath, "cwd", "C", ".", "Path to the Pulumi project directory")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}

func getStackName() (string, error) {
	if stackName != "" {
		return stackName, nil
	}

	// Try to detect from environment or Pulumi.yaml
	if envStack := os.Getenv("PULUMI_STACK"); envStack != "" {
		return envStack, nil
	}

	return "", fmt.Errorf("stack name is required: use --stack flag or set PULUMI_STACK environment variable")
}

func getProjectPath() string {
	return projectPath
}

func isVerbose() bool {
	return verbose
}
