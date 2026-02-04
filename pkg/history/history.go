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
	Version     int
	Kind        string
	StartTime   time.Time
	EndTime     time.Time
	Result      string
	Message     string
	ResourceChanges map[string]int
}

// GetStackHistory retrieves the deployment history for a stack
func GetStackHistory(ctx context.Context, projectPath, stackName string) ([]UpdateInfo, error) {
	// Create or select the stack using the local workspace
	stack, err := auto.SelectStackLocalSource(ctx, stackName, projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to select stack %s: %w", stackName, err)
	}

	// Get the stack history
	history, err := stack.History(ctx, 0, 0) // pageSize=0, page=0 means get all
	if err != nil {
		return nil, fmt.Errorf("failed to get stack history: %w", err)
	}

	var updates []UpdateInfo
	for _, update := range history {
		info := UpdateInfo{
			Version:   update.Version,
			Kind:      update.Kind,
			Result:    update.Result,
			Message:   update.Message,
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

	return updates, nil
}

// GetUpdateByVersion retrieves a specific update by version number
func GetUpdateByVersion(ctx context.Context, projectPath, stackName string, version int) (*UpdateInfo, error) {
	history, err := GetStackHistory(ctx, projectPath, stackName)
	if err != nil {
		return nil, err
	}

	for _, update := range history {
		if update.Version == version {
			return &update, nil
		}
	}

	return nil, fmt.Errorf("version %d not found in stack history", version)
}

// GetLatestVersion returns the latest version number
func GetLatestVersion(ctx context.Context, projectPath, stackName string) (int, error) {
	history, err := GetStackHistory(ctx, projectPath, stackName)
	if err != nil {
		return 0, err
	}

	if len(history) == 0 {
		return 0, fmt.Errorf("no deployment history found for stack %s", stackName)
	}

	// History is returned in reverse chronological order
	return history[0].Version, nil
}
