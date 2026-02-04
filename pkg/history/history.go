// Copyright 2024 Pegasus Heavy Industries LLC
// Contact: pegasusheavyindustries@gmail.com

package history

import (
	"context"
	"fmt"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

// UpdateInfo represents information about a stack update
type UpdateInfo struct {
	Version         int
	Kind            string
	StartTime       time.Time
	EndTime         time.Time
	Result          string
	Message         string
	ResourceChanges map[string]int
}

// GetStackHistory retrieves the deployment history for a stack
func GetStackHistory(ctx context.Context, projectPath, stackName string) ([]UpdateInfo, error) {
	return GetStackHistoryWithSelector(ctx, projectPath, stackName, DefaultSelector)
}

// GetStackHistoryWithSelector retrieves the deployment history using a custom selector
func GetStackHistoryWithSelector(ctx context.Context, projectPath, stackName string, selector StackSelector) ([]UpdateInfo, error) {
	// Create or select the stack using the provided selector
	stack, err := selector.SelectStack(ctx, stackName, projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to select stack %s: %w", stackName, err)
	}

	// Get the stack history
	history, err := stack.History(ctx, 0, 0) // pageSize=0, page=0 means get all
	if err != nil {
		return nil, fmt.Errorf("failed to get stack history: %w", err)
	}

	return ConvertUpdates(history), nil
}

// ConvertUpdates converts auto.UpdateSummary slice to UpdateInfo slice
func ConvertUpdates(history []auto.UpdateSummary) []UpdateInfo {
	var updates []UpdateInfo
	for _, update := range history {
		info := UpdateInfo{
			Version:         update.Version,
			Kind:            update.Kind,
			Result:          update.Result,
			Message:         update.Message,
			ResourceChanges: make(map[string]int),
		}

		// Parse timestamps
		if update.StartTime != "" {
			if t, err := time.Parse(time.RFC3339, update.StartTime); err == nil {
				info.StartTime = t
			}
		}
		if update.EndTime != nil && *update.EndTime != "" {
			if t, err := time.Parse(time.RFC3339, *update.EndTime); err == nil {
				info.EndTime = t
			}
		}

		// Copy resource changes
		if update.ResourceChanges != nil {
			for k, v := range *update.ResourceChanges {
				info.ResourceChanges[k] = v
			}
		}

		updates = append(updates, info)
	}

	return updates
}

// GetUpdateByVersion retrieves a specific update by version number
func GetUpdateByVersion(ctx context.Context, projectPath, stackName string, version int) (*UpdateInfo, error) {
	return GetUpdateByVersionWithSelector(ctx, projectPath, stackName, version, DefaultSelector)
}

// GetUpdateByVersionWithSelector retrieves a specific update by version number using a custom selector
func GetUpdateByVersionWithSelector(ctx context.Context, projectPath, stackName string, version int, selector StackSelector) (*UpdateInfo, error) {
	history, err := GetStackHistoryWithSelector(ctx, projectPath, stackName, selector)
	if err != nil {
		return nil, err
	}

	return FindUpdateByVersion(history, version)
}

// FindUpdateByVersion finds an update by version in a slice of updates
func FindUpdateByVersion(history []UpdateInfo, version int) (*UpdateInfo, error) {
	for _, update := range history {
		if update.Version == version {
			return &update, nil
		}
	}

	return nil, fmt.Errorf("version %d not found in stack history", version)
}

// GetLatestVersion returns the latest version number
func GetLatestVersion(ctx context.Context, projectPath, stackName string) (int, error) {
	return GetLatestVersionWithSelector(ctx, projectPath, stackName, DefaultSelector)
}

// GetLatestVersionWithSelector returns the latest version number using a custom selector
func GetLatestVersionWithSelector(ctx context.Context, projectPath, stackName string, selector StackSelector) (int, error) {
	history, err := GetStackHistoryWithSelector(ctx, projectPath, stackName, selector)
	if err != nil {
		return 0, err
	}

	return GetLatestVersionFromHistory(history, stackName)
}

// GetLatestVersionFromHistory returns the latest version from a history slice
func GetLatestVersionFromHistory(history []UpdateInfo, stackName string) (int, error) {
	if len(history) == 0 {
		return 0, fmt.Errorf("no deployment history found for stack %s", stackName)
	}

	// History is returned in reverse chronological order
	return history[0].Version, nil
}
