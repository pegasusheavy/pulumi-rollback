// Copyright 2026 Pegasus Heavy Industries LLC
// Contact: pegasusheavyindustries@gmail.com

package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/PegasusHeavyIndustries/pulumi-rollback/pkg/history"
	"github.com/PegasusHeavyIndustries/pulumi-rollback/pkg/rollback"
	"github.com/spf13/cobra"
)

var (
	rollbackVersion int
	skipConfirm     bool
)

var toCmd = &cobra.Command{
	Use:   "to",
	Short: "Roll back to a specific version",
	Long: `Roll back the stack to a specific version from the deployment history.

This will:
1. Restore the stack state to the target version
2. Refresh to reconcile with actual infrastructure
3. Run 'up' to apply any necessary changes

Examples:
  # Roll back to version 5
  pulumi-rollback to --stack mystack --version 5

  # Roll back without confirmation prompt
  pulumi-rollback to --stack mystack --version 5 --yes`,
	RunE: runRollback,
}

func init() {
	rootCmd.AddCommand(toCmd)
	toCmd.Flags().IntVarP(&rollbackVersion, "version", "V", 0, "Target version to roll back to (required)")
	toCmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation prompt")
	toCmd.MarkFlagRequired("version")
}

func runRollback(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	stack, err := getStackName()
	if err != nil {
		return err
	}

	projectPath := getProjectPath()

	// Validate the version exists
	update, err := history.GetUpdateByVersion(ctx, projectPath, stack, rollbackVersion)
	if err != nil {
		return fmt.Errorf("failed to find version %d: %w", rollbackVersion, err)
	}

	// Check if this is the latest version
	latest, err := history.GetLatestVersion(ctx, projectPath, stack)
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}

	if rollbackVersion == latest {
		fmt.Println("Version", rollbackVersion, "is the current version. No rollback needed.")
		return nil
	}

	// Show target version info
	fmt.Printf("Rolling back stack '%s' to version %d\n", stack, rollbackVersion)
	fmt.Printf("  Kind: %s\n", update.Kind)
	fmt.Printf("  Result: %s\n", update.Result)
	fmt.Printf("  Time: %s\n", formatTime(update.StartTime))
	if update.Message != "" {
		fmt.Printf("  Message: %s\n", update.Message)
	}
	fmt.Println()

	// Warn about rollback
	fmt.Println("⚠️  WARNING: This will modify your infrastructure!")
	fmt.Printf("   Current version: %d\n", latest)
	fmt.Printf("   Target version:  %d\n", rollbackVersion)
	fmt.Println()

	// Confirmation prompt
	if !skipConfirm {
		fmt.Print("Do you want to proceed? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Rollback cancelled.")
			return nil
		}
	}

	fmt.Println("\nStarting rollback...")

	opts := rollback.RollbackOptions{
		ProjectPath:   projectPath,
		StackName:     stack,
		TargetVersion: rollbackVersion,
		DryRun:        false,
		Verbose:       isVerbose(),
		Output:        os.Stdout,
	}

	result, err := rollback.ExecuteRollback(ctx, opts)
	if err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	fmt.Println("\n✓", result.Message)

	if len(result.ResourceChanges) > 0 {
		fmt.Println("\nResource changes applied:")
		for change, count := range result.ResourceChanges {
			fmt.Printf("  %s: %d\n", change, count)
		}
	}

	return nil
}
