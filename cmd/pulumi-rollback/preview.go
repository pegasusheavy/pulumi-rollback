// Copyright 2026 Pegasus Heavy Industries LLC
// Contact: pegasusheavyindustries@gmail.com

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/PegasusHeavyIndustries/pulumi-rollback/pkg/history"
	"github.com/PegasusHeavyIndustries/pulumi-rollback/pkg/rollback"
	"github.com/spf13/cobra"
)

var (
	previewVersion int
)

var previewCmd = &cobra.Command{
	Use:   "preview",
	Short: "Preview changes that would be made by rolling back",
	Long: `Preview what infrastructure changes would be made by rolling back
to a specific version without actually making any changes.

This is equivalent to 'pulumi preview' but targeting a historical state.

Examples:
  # Preview rolling back to version 5
  pulumi-rollback preview --stack mystack --version 5`,
	RunE: runPreview,
}

func init() {
	rootCmd.AddCommand(previewCmd)
	previewCmd.Flags().IntVarP(&previewVersion, "version", "V", 0, "Target version to roll back to (required)")
	previewCmd.MarkFlagRequired("version")
}

func runPreview(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	stack, err := getStackName()
	if err != nil {
		return err
	}

	projectPath := getProjectPath()

	// Validate the version exists
	update, err := history.GetUpdateByVersion(ctx, projectPath, stack, previewVersion)
	if err != nil {
		return fmt.Errorf("failed to find version %d: %w", previewVersion, err)
	}

	// Check if this is the latest version
	latest, err := history.GetLatestVersion(ctx, projectPath, stack)
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}

	if previewVersion == latest {
		fmt.Println("Warning: Version", previewVersion, "is the current version. No rollback needed.")
		return nil
	}

	fmt.Printf("Previewing rollback to version %d...\n", previewVersion)
	fmt.Printf("  Kind: %s\n", update.Kind)
	fmt.Printf("  Result: %s\n", update.Result)
	fmt.Printf("  Time: %s\n", formatTime(update.StartTime))
	if update.Message != "" {
		fmt.Printf("  Message: %s\n", update.Message)
	}
	fmt.Println()

	opts := rollback.RollbackOptions{
		ProjectPath:   projectPath,
		StackName:     stack,
		TargetVersion: previewVersion,
		DryRun:        true,
		Verbose:       isVerbose(),
		Output:        os.Stdout,
	}

	result, err := rollback.PreviewRollback(ctx, opts)
	if err != nil {
		return fmt.Errorf("preview failed: %w", err)
	}

	fmt.Println("\n" + result.Message)

	if len(result.ResourceChanges) > 0 {
		fmt.Println("\nResource changes:")
		for change, count := range result.ResourceChanges {
			fmt.Printf("  %s: %d\n", change, count)
		}
	}

	fmt.Println("\nTo execute this rollback, run:")
	fmt.Printf("  pulumi-rollback to --stack %s --version %d\n", stack, previewVersion)

	return nil
}
