// Copyright 2024 Pegasus Heavy Industries LLC
// Contact: pegasusheavyindustries@gmail.com

package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/PegasusHeavyIndustries/pulumi-rollback/pkg/history"
	"github.com/spf13/cobra"
)

var (
	listLimit int
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List deployment history for a stack",
	Long: `List the deployment history for a Pulumi stack, showing version numbers,
timestamps, and deployment results.

Examples:
  # List all deployment history
  pulumi-rollback list --stack mystack

  # List last 10 deployments
  pulumi-rollback list --stack mystack --limit 10`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().IntVarP(&listLimit, "limit", "n", 0, "Limit the number of entries to show (0 = all)")
}

func runList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	stack, err := getStackName()
	if err != nil {
		return err
	}

	projectPath := getProjectPath()

	if isVerbose() {
		fmt.Printf("Fetching history for stack %s in %s...\n", stack, projectPath)
	}

	updates, err := history.GetStackHistory(ctx, projectPath, stack)
	if err != nil {
		return fmt.Errorf("failed to get stack history: %w", err)
	}

	if len(updates) == 0 {
		fmt.Println("No deployment history found for this stack.")
		return nil
	}

	// Apply limit if specified
	if listLimit > 0 && listLimit < len(updates) {
		updates = updates[:listLimit]
	}

	// Create a tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "VERSION\tKIND\tRESULT\tTIME\tCHANGES\tMESSAGE")
	fmt.Fprintln(w, "-------\t----\t------\t----\t-------\t-------")

	for _, update := range updates {
		timeStr := formatTime(update.StartTime)
		changesStr := formatChanges(update.ResourceChanges)
		message := truncateString(update.Message, 40)

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
			update.Version,
			update.Kind,
			formatResult(update.Result),
			timeStr,
			changesStr,
			message,
		)
	}

	w.Flush()

	fmt.Printf("\nTotal: %d deployment(s)\n", len(updates))
	fmt.Println("\nUse 'pulumi-rollback preview --stack <stack> --version <n>' to preview a rollback")

	return nil
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}
	return t.Format("2006-01-02 15:04")
}

func formatResult(result string) string {
	switch result {
	case "succeeded":
		return "✓ success"
	case "failed":
		return "✗ failed"
	case "in-progress":
		return "⟳ running"
	default:
		return result
	}
}

func formatChanges(changes map[string]int) string {
	if len(changes) == 0 {
		return "-"
	}

	create := changes["create"]
	update := changes["update"]
	delete := changes["delete"]
	same := changes["same"]

	parts := []string{}
	if create > 0 {
		parts = append(parts, fmt.Sprintf("+%d", create))
	}
	if update > 0 {
		parts = append(parts, fmt.Sprintf("~%d", update))
	}
	if delete > 0 {
		parts = append(parts, fmt.Sprintf("-%d", delete))
	}
	if same > 0 && len(parts) == 0 {
		return fmt.Sprintf("=%d", same)
	}

	if len(parts) == 0 {
		return "-"
	}

	result := ""
	for i, p := range parts {
		if i > 0 {
			result += " "
		}
		result += p
	}
	return result
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
