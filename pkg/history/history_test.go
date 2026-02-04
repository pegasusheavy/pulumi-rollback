// Copyright 2024 Pegasus Heavy Industries LLC
// Contact: pegasusheavyindustries@gmail.com

package history

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

// MockStack implements the Stack interface for testing
type MockStack struct {
	HistoryFunc func(ctx context.Context, pageSize int, page int) ([]auto.UpdateSummary, error)
}

func (m *MockStack) History(ctx context.Context, pageSize int, page int) ([]auto.UpdateSummary, error) {
	if m.HistoryFunc != nil {
		return m.HistoryFunc(ctx, pageSize, page)
	}
	return nil, nil
}

// MockStackSelector implements the StackSelector interface for testing
type MockStackSelector struct {
	SelectStackFunc func(ctx context.Context, stackName, projectPath string) (Stack, error)
}

func (m *MockStackSelector) SelectStack(ctx context.Context, stackName, projectPath string) (Stack, error) {
	if m.SelectStackFunc != nil {
		return m.SelectStackFunc(ctx, stackName, projectPath)
	}
	return nil, nil
}

func TestGetStackHistoryWithSelector_Success(t *testing.T) {
	endTime := "2024-01-15T10:05:00Z"
	resourceChanges := map[string]int{"create": 3, "update": 2}

	mockStack := &MockStack{
		HistoryFunc: func(ctx context.Context, pageSize int, page int) ([]auto.UpdateSummary, error) {
			return []auto.UpdateSummary{
				{
					Version:         5,
					Kind:            "update",
					StartTime:       "2024-01-15T10:00:00Z",
					EndTime:         &endTime,
					Result:          "succeeded",
					Message:         "Test deployment",
					ResourceChanges: &resourceChanges,
				},
			}, nil
		},
	}

	mockSelector := &MockStackSelector{
		SelectStackFunc: func(ctx context.Context, stackName, projectPath string) (Stack, error) {
			return mockStack, nil
		},
	}

	ctx := context.Background()
	history, err := GetStackHistoryWithSelector(ctx, "/path/to/project", "test-stack", mockSelector)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf("Expected 1 update, got %d", len(history))
	}

	update := history[0]
	if update.Version != 5 {
		t.Errorf("Expected Version 5, got %d", update.Version)
	}
	if update.Kind != "update" {
		t.Errorf("Expected Kind 'update', got %q", update.Kind)
	}
	if update.Result != "succeeded" {
		t.Errorf("Expected Result 'succeeded', got %q", update.Result)
	}
	if update.Message != "Test deployment" {
		t.Errorf("Expected Message 'Test deployment', got %q", update.Message)
	}
	if update.ResourceChanges["create"] != 3 {
		t.Errorf("Expected ResourceChanges['create'] = 3, got %d", update.ResourceChanges["create"])
	}
}

func TestGetStackHistoryWithSelector_SelectStackError(t *testing.T) {
	mockSelector := &MockStackSelector{
		SelectStackFunc: func(ctx context.Context, stackName, projectPath string) (Stack, error) {
			return nil, errors.New("stack not found")
		},
	}

	ctx := context.Background()
	_, err := GetStackHistoryWithSelector(ctx, "/path/to/project", "test-stack", mockSelector)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !errors.Is(err, err) {
		t.Errorf("Expected error to wrap original error")
	}
}

func TestGetStackHistoryWithSelector_HistoryError(t *testing.T) {
	mockStack := &MockStack{
		HistoryFunc: func(ctx context.Context, pageSize int, page int) ([]auto.UpdateSummary, error) {
			return nil, errors.New("history unavailable")
		},
	}

	mockSelector := &MockStackSelector{
		SelectStackFunc: func(ctx context.Context, stackName, projectPath string) (Stack, error) {
			return mockStack, nil
		},
	}

	ctx := context.Background()
	_, err := GetStackHistoryWithSelector(ctx, "/path/to/project", "test-stack", mockSelector)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestConvertUpdates(t *testing.T) {
	endTime := "2024-01-15T10:05:00Z"
	resourceChanges := map[string]int{"create": 1}

	tests := []struct {
		name     string
		input    []auto.UpdateSummary
		expected []UpdateInfo
	}{
		{
			name:     "empty slice",
			input:    []auto.UpdateSummary{},
			expected: nil,
		},
		{
			name:     "nil slice",
			input:    nil,
			expected: nil,
		},
		{
			name: "single update with all fields",
			input: []auto.UpdateSummary{
				{
					Version:         1,
					Kind:            "update",
					StartTime:       "2024-01-15T10:00:00Z",
					EndTime:         &endTime,
					Result:          "succeeded",
					Message:         "test",
					ResourceChanges: &resourceChanges,
				},
			},
			expected: []UpdateInfo{
				{
					Version:         1,
					Kind:            "update",
					StartTime:       time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
					EndTime:         time.Date(2024, 1, 15, 10, 5, 0, 0, time.UTC),
					Result:          "succeeded",
					Message:         "test",
					ResourceChanges: map[string]int{"create": 1},
				},
			},
		},
		{
			name: "update with nil end time",
			input: []auto.UpdateSummary{
				{
					Version:   2,
					EndTime:   nil,
					StartTime: "2024-01-15T10:00:00Z",
				},
			},
			expected: []UpdateInfo{
				{
					Version:         2,
					StartTime:       time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
					EndTime:         time.Time{},
					ResourceChanges: map[string]int{},
				},
			},
		},
		{
			name: "update with nil resource changes",
			input: []auto.UpdateSummary{
				{
					Version:         3,
					ResourceChanges: nil,
				},
			},
			expected: []UpdateInfo{
				{
					Version:         3,
					ResourceChanges: map[string]int{},
				},
			},
		},
		{
			name: "update with invalid start time format",
			input: []auto.UpdateSummary{
				{
					Version:   4,
					StartTime: "invalid-date",
				},
			},
			expected: []UpdateInfo{
				{
					Version:         4,
					StartTime:       time.Time{},
					ResourceChanges: map[string]int{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertUpdates(tt.input)

			if len(result) != len(tt.expected) {
				t.Fatalf("Expected %d updates, got %d", len(tt.expected), len(result))
			}

			for i, exp := range tt.expected {
				if result[i].Version != exp.Version {
					t.Errorf("Update %d: expected Version %d, got %d", i, exp.Version, result[i].Version)
				}
			}
		})
	}
}

func TestConvertUpdates_EmptyEndTime(t *testing.T) {
	emptyEndTime := ""
	input := []auto.UpdateSummary{
		{
			Version: 1,
			EndTime: &emptyEndTime,
		},
	}

	result := ConvertUpdates(input)

	if len(result) != 1 {
		t.Fatalf("Expected 1 update, got %d", len(result))
	}

	if !result[0].EndTime.IsZero() {
		t.Error("Expected EndTime to be zero time for empty end time string")
	}
}

func TestFindUpdateByVersion(t *testing.T) {
	history := []UpdateInfo{
		{Version: 1, Kind: "create"},
		{Version: 2, Kind: "update"},
		{Version: 3, Kind: "update"},
	}

	tests := []struct {
		name        string
		version     int
		expectError bool
		expectKind  string
	}{
		{
			name:        "find first version",
			version:     1,
			expectError: false,
			expectKind:  "create",
		},
		{
			name:        "find middle version",
			version:     2,
			expectError: false,
			expectKind:  "update",
		},
		{
			name:        "find last version",
			version:     3,
			expectError: false,
			expectKind:  "update",
		},
		{
			name:        "version not found",
			version:     99,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FindUpdateByVersion(history, tt.version)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.Kind != tt.expectKind {
					t.Errorf("Expected Kind %q, got %q", tt.expectKind, result.Kind)
				}
			}
		})
	}
}

func TestFindUpdateByVersion_EmptyHistory(t *testing.T) {
	_, err := FindUpdateByVersion([]UpdateInfo{}, 1)
	if err == nil {
		t.Error("Expected error for empty history, got nil")
	}
}

func TestGetLatestVersionFromHistory(t *testing.T) {
	tests := []struct {
		name        string
		history     []UpdateInfo
		stackName   string
		expected    int
		expectError bool
	}{
		{
			name: "single version",
			history: []UpdateInfo{
				{Version: 5},
			},
			stackName:   "test",
			expected:    5,
			expectError: false,
		},
		{
			name: "multiple versions - latest first",
			history: []UpdateInfo{
				{Version: 10},
				{Version: 9},
				{Version: 8},
			},
			stackName:   "test",
			expected:    10,
			expectError: false,
		},
		{
			name:        "empty history",
			history:     []UpdateInfo{},
			stackName:   "test",
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetLatestVersionFromHistory(tt.history, tt.stackName)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected version %d, got %d", tt.expected, result)
				}
			}
		})
	}
}

func TestGetUpdateByVersionWithSelector(t *testing.T) {
	mockStack := &MockStack{
		HistoryFunc: func(ctx context.Context, pageSize int, page int) ([]auto.UpdateSummary, error) {
			return []auto.UpdateSummary{
				{Version: 1, Kind: "create"},
				{Version: 2, Kind: "update"},
			}, nil
		},
	}

	mockSelector := &MockStackSelector{
		SelectStackFunc: func(ctx context.Context, stackName, projectPath string) (Stack, error) {
			return mockStack, nil
		},
	}

	ctx := context.Background()

	// Test finding existing version
	result, err := GetUpdateByVersionWithSelector(ctx, "/path", "stack", 2, mockSelector)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Kind != "update" {
		t.Errorf("Expected Kind 'update', got %q", result.Kind)
	}

	// Test version not found
	_, err = GetUpdateByVersionWithSelector(ctx, "/path", "stack", 99, mockSelector)
	if err == nil {
		t.Error("Expected error for non-existent version")
	}
}

func TestGetLatestVersionWithSelector(t *testing.T) {
	mockStack := &MockStack{
		HistoryFunc: func(ctx context.Context, pageSize int, page int) ([]auto.UpdateSummary, error) {
			return []auto.UpdateSummary{
				{Version: 5},
				{Version: 4},
			}, nil
		},
	}

	mockSelector := &MockStackSelector{
		SelectStackFunc: func(ctx context.Context, stackName, projectPath string) (Stack, error) {
			return mockStack, nil
		},
	}

	ctx := context.Background()
	result, err := GetLatestVersionWithSelector(ctx, "/path", "stack", mockSelector)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != 5 {
		t.Errorf("Expected version 5, got %d", result)
	}
}

func TestGetLatestVersionWithSelector_Error(t *testing.T) {
	mockSelector := &MockStackSelector{
		SelectStackFunc: func(ctx context.Context, stackName, projectPath string) (Stack, error) {
			return nil, errors.New("connection failed")
		},
	}

	ctx := context.Background()
	_, err := GetLatestVersionWithSelector(ctx, "/path", "stack", mockSelector)

	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestUpdateInfoStruct(t *testing.T) {
	now := time.Now()
	info := UpdateInfo{
		Version:   5,
		Kind:      "update",
		StartTime: now,
		EndTime:   now.Add(time.Minute),
		Result:    "succeeded",
		Message:   "test deployment",
		ResourceChanges: map[string]int{
			"create": 3,
			"update": 2,
			"delete": 1,
		},
	}

	if info.Version != 5 {
		t.Errorf("Expected Version to be 5, got %d", info.Version)
	}
	if info.Kind != "update" {
		t.Errorf("Expected Kind to be 'update', got %q", info.Kind)
	}
	if info.Result != "succeeded" {
		t.Errorf("Expected Result to be 'succeeded', got %q", info.Result)
	}
	if info.Message != "test deployment" {
		t.Errorf("Expected Message to be 'test deployment', got %q", info.Message)
	}
}
