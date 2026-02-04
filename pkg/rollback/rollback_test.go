// Copyright 2024 Pegasus Heavy Industries LLC
// Contact: pegasusheavyindustries@gmail.com

package rollback

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optpreview"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optrefresh"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

// MockRollbackStack implements RollbackStack for testing
type MockRollbackStack struct {
	ExportFunc  func(ctx context.Context) (apitype.UntypedDeployment, error)
	ImportFunc  func(ctx context.Context, state apitype.UntypedDeployment) error
	HistoryFunc func(ctx context.Context, pageSize int, page int) ([]auto.UpdateSummary, error)
	PreviewFunc func(ctx context.Context, opts ...optpreview.Option) (auto.PreviewResult, error)
	RefreshFunc func(ctx context.Context, opts ...optrefresh.Option) (auto.RefreshResult, error)
	UpFunc      func(ctx context.Context, opts ...optup.Option) (auto.UpResult, error)
}

func (m *MockRollbackStack) Export(ctx context.Context) (apitype.UntypedDeployment, error) {
	if m.ExportFunc != nil {
		return m.ExportFunc(ctx)
	}
	return apitype.UntypedDeployment{Deployment: json.RawMessage(`{}`)}, nil
}

func (m *MockRollbackStack) Import(ctx context.Context, state apitype.UntypedDeployment) error {
	if m.ImportFunc != nil {
		return m.ImportFunc(ctx, state)
	}
	return nil
}

func (m *MockRollbackStack) History(ctx context.Context, pageSize int, page int) ([]auto.UpdateSummary, error) {
	if m.HistoryFunc != nil {
		return m.HistoryFunc(ctx, pageSize, page)
	}
	return []auto.UpdateSummary{{Version: 1}}, nil
}

func (m *MockRollbackStack) Preview(ctx context.Context, opts ...optpreview.Option) (auto.PreviewResult, error) {
	if m.PreviewFunc != nil {
		return m.PreviewFunc(ctx, opts...)
	}
	return auto.PreviewResult{}, nil
}

func (m *MockRollbackStack) Refresh(ctx context.Context, opts ...optrefresh.Option) (auto.RefreshResult, error) {
	if m.RefreshFunc != nil {
		return m.RefreshFunc(ctx, opts...)
	}
	return auto.RefreshResult{}, nil
}

func (m *MockRollbackStack) Up(ctx context.Context, opts ...optup.Option) (auto.UpResult, error) {
	if m.UpFunc != nil {
		return m.UpFunc(ctx, opts...)
	}
	return auto.UpResult{}, nil
}

// MockStackOperator implements StackOperator for testing
type MockStackOperator struct {
	SelectStackFunc func(ctx context.Context, stackName, projectPath string) (RollbackStack, error)
}

func (m *MockStackOperator) SelectStack(ctx context.Context, stackName, projectPath string) (RollbackStack, error) {
	if m.SelectStackFunc != nil {
		return m.SelectStackFunc(ctx, stackName, projectPath)
	}
	return &MockRollbackStack{}, nil
}

func TestConvertOpTypeChangeSummary(t *testing.T) {
	tests := []struct {
		name     string
		input    map[apitype.OpType]int
		expected map[string]int
	}{
		{
			name:     "nil map",
			input:    nil,
			expected: map[string]int{},
		},
		{
			name:     "empty map",
			input:    map[apitype.OpType]int{},
			expected: map[string]int{},
		},
		{
			name: "single entry",
			input: map[apitype.OpType]int{
				apitype.OpCreate: 5,
			},
			expected: map[string]int{
				"create": 5,
			},
		},
		{
			name: "multiple entries",
			input: map[apitype.OpType]int{
				apitype.OpCreate: 3,
				apitype.OpUpdate: 2,
				apitype.OpDelete: 1,
			},
			expected: map[string]int{
				"create": 3,
				"update": 2,
				"delete": 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertOpTypeChangeSummary(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("convertOpTypeChangeSummary() returned map with %d entries, want %d", len(result), len(tt.expected))
			}

			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("convertOpTypeChangeSummary()[%q] = %d, want %d", k, result[k], v)
				}
			}
		})
	}
}

func TestVersionExistsInHistory(t *testing.T) {
	history := []auto.UpdateSummary{
		{Version: 1},
		{Version: 2},
		{Version: 5},
	}

	tests := []struct {
		name     string
		version  int
		expected bool
	}{
		{"version exists - first", 1, true},
		{"version exists - middle", 2, true},
		{"version exists - last", 5, true},
		{"version not exists", 3, false},
		{"version not exists - zero", 0, false},
		{"version not exists - negative", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VersionExistsInHistory(history, tt.version)
			if result != tt.expected {
				t.Errorf("VersionExistsInHistory() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestVersionExistsInHistory_EmptyHistory(t *testing.T) {
	result := VersionExistsInHistory([]auto.UpdateSummary{}, 1)
	if result {
		t.Error("Expected false for empty history")
	}
}

func TestValidateDeployment(t *testing.T) {
	tests := []struct {
		name        string
		deployment  apitype.UntypedDeployment
		expectError bool
	}{
		{
			name:        "valid empty object",
			deployment:  apitype.UntypedDeployment{Deployment: json.RawMessage(`{}`)},
			expectError: false,
		},
		{
			name:        "valid with data",
			deployment:  apitype.UntypedDeployment{Deployment: json.RawMessage(`{"key": "value"}`)},
			expectError: false,
		},
		{
			name:        "invalid json",
			deployment:  apitype.UntypedDeployment{Deployment: json.RawMessage(`{invalid}`)},
			expectError: true,
		},
		{
			name:        "empty deployment",
			deployment:  apitype.UntypedDeployment{Deployment: json.RawMessage(``)},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDeployment(tt.deployment)
			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestPreviewRollback_Success(t *testing.T) {
	mockStack := &MockRollbackStack{
		HistoryFunc: func(ctx context.Context, pageSize int, page int) ([]auto.UpdateSummary, error) {
			return []auto.UpdateSummary{{Version: 1}, {Version: 2}}, nil
		},
		ExportFunc: func(ctx context.Context) (apitype.UntypedDeployment, error) {
			return apitype.UntypedDeployment{Deployment: json.RawMessage(`{}`)}, nil
		},
		PreviewFunc: func(ctx context.Context, opts ...optpreview.Option) (auto.PreviewResult, error) {
			return auto.PreviewResult{
				StdOut: "preview output",
				ChangeSummary: map[apitype.OpType]int{
					apitype.OpCreate: 1,
				},
			}, nil
		},
	}

	mockOperator := &MockStackOperator{
		SelectStackFunc: func(ctx context.Context, stackName, projectPath string) (RollbackStack, error) {
			return mockStack, nil
		},
	}

	var output bytes.Buffer
	opts := RollbackOptions{
		ProjectPath:   "/path/to/project",
		StackName:     "test-stack",
		TargetVersion: 1,
		Output:        &output,
		Operator:      mockOperator,
	}

	result, err := PreviewRollback(context.Background(), opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Success {
		t.Error("Expected Success to be true")
	}
	if result.ResourceChanges["create"] != 1 {
		t.Errorf("Expected ResourceChanges['create'] = 1, got %d", result.ResourceChanges["create"])
	}
}

func TestPreviewRollback_SelectStackError(t *testing.T) {
	mockOperator := &MockStackOperator{
		SelectStackFunc: func(ctx context.Context, stackName, projectPath string) (RollbackStack, error) {
			return nil, errors.New("stack not found")
		},
	}

	opts := RollbackOptions{
		StackName: "test",
		Operator:  mockOperator,
	}

	_, err := PreviewRollback(context.Background(), opts)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestPreviewRollback_ExportError(t *testing.T) {
	mockStack := &MockRollbackStack{
		ExportFunc: func(ctx context.Context) (apitype.UntypedDeployment, error) {
			return apitype.UntypedDeployment{}, errors.New("export failed")
		},
	}

	mockOperator := &MockStackOperator{
		SelectStackFunc: func(ctx context.Context, stackName, projectPath string) (RollbackStack, error) {
			return mockStack, nil
		},
	}

	opts := RollbackOptions{
		StackName: "test",
		Operator:  mockOperator,
	}

	_, err := PreviewRollback(context.Background(), opts)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestPreviewRollback_VersionNotFound(t *testing.T) {
	mockStack := &MockRollbackStack{
		ExportFunc: func(ctx context.Context) (apitype.UntypedDeployment, error) {
			return apitype.UntypedDeployment{Deployment: json.RawMessage(`{}`)}, nil
		},
		HistoryFunc: func(ctx context.Context, pageSize int, page int) ([]auto.UpdateSummary, error) {
			return []auto.UpdateSummary{{Version: 1}}, nil
		},
	}

	mockOperator := &MockStackOperator{
		SelectStackFunc: func(ctx context.Context, stackName, projectPath string) (RollbackStack, error) {
			return mockStack, nil
		},
	}

	opts := RollbackOptions{
		StackName:     "test",
		TargetVersion: 99,
		Operator:      mockOperator,
	}

	_, err := PreviewRollback(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for non-existent version")
	}
}

func TestPreviewRollback_ImportError(t *testing.T) {
	mockStack := &MockRollbackStack{
		ExportFunc: func(ctx context.Context) (apitype.UntypedDeployment, error) {
			return apitype.UntypedDeployment{Deployment: json.RawMessage(`{}`)}, nil
		},
		HistoryFunc: func(ctx context.Context, pageSize int, page int) ([]auto.UpdateSummary, error) {
			return []auto.UpdateSummary{{Version: 1}}, nil
		},
		ImportFunc: func(ctx context.Context, state apitype.UntypedDeployment) error {
			return errors.New("import failed")
		},
	}

	mockOperator := &MockStackOperator{
		SelectStackFunc: func(ctx context.Context, stackName, projectPath string) (RollbackStack, error) {
			return mockStack, nil
		},
	}

	opts := RollbackOptions{
		StackName:     "test",
		TargetVersion: 1,
		Operator:      mockOperator,
	}

	_, err := PreviewRollback(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for import failure")
	}
}

func TestPreviewRollback_PreviewError(t *testing.T) {
	importCount := 0
	mockStack := &MockRollbackStack{
		ExportFunc: func(ctx context.Context) (apitype.UntypedDeployment, error) {
			return apitype.UntypedDeployment{Deployment: json.RawMessage(`{}`)}, nil
		},
		HistoryFunc: func(ctx context.Context, pageSize int, page int) ([]auto.UpdateSummary, error) {
			return []auto.UpdateSummary{{Version: 1}}, nil
		},
		ImportFunc: func(ctx context.Context, state apitype.UntypedDeployment) error {
			importCount++
			return nil
		},
		PreviewFunc: func(ctx context.Context, opts ...optpreview.Option) (auto.PreviewResult, error) {
			return auto.PreviewResult{}, errors.New("preview failed")
		},
	}

	mockOperator := &MockStackOperator{
		SelectStackFunc: func(ctx context.Context, stackName, projectPath string) (RollbackStack, error) {
			return mockStack, nil
		},
	}

	var output bytes.Buffer
	opts := RollbackOptions{
		StackName:     "test",
		TargetVersion: 1,
		Operator:      mockOperator,
		Output:        &output,
	}

	_, err := PreviewRollback(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for preview failure")
	}

	// Verify state was restored (import called twice)
	if importCount != 2 {
		t.Errorf("Expected import to be called twice (once for target, once for restore), got %d", importCount)
	}
}

func TestPreviewRollback_RestoreError(t *testing.T) {
	importCount := 0
	mockStack := &MockRollbackStack{
		ExportFunc: func(ctx context.Context) (apitype.UntypedDeployment, error) {
			return apitype.UntypedDeployment{Deployment: json.RawMessage(`{}`)}, nil
		},
		HistoryFunc: func(ctx context.Context, pageSize int, page int) ([]auto.UpdateSummary, error) {
			return []auto.UpdateSummary{{Version: 1}}, nil
		},
		ImportFunc: func(ctx context.Context, state apitype.UntypedDeployment) error {
			importCount++
			if importCount == 2 {
				return errors.New("restore failed")
			}
			return nil
		},
		PreviewFunc: func(ctx context.Context, opts ...optpreview.Option) (auto.PreviewResult, error) {
			return auto.PreviewResult{}, nil
		},
	}

	mockOperator := &MockStackOperator{
		SelectStackFunc: func(ctx context.Context, stackName, projectPath string) (RollbackStack, error) {
			return mockStack, nil
		},
	}

	var output bytes.Buffer
	opts := RollbackOptions{
		StackName:     "test",
		TargetVersion: 1,
		Operator:      mockOperator,
		Output:        &output,
	}

	result, err := PreviewRollback(context.Background(), opts)
	// Should still succeed even with restore error
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("Expected Success to be true")
	}

	// Verify warning was written
	if !bytes.Contains(output.Bytes(), []byte("Warning")) {
		t.Error("Expected warning message in output")
	}
}

func TestExecuteRollback_Success(t *testing.T) {
	resourceChanges := map[string]int{"create": 2}
	mockStack := &MockRollbackStack{
		HistoryFunc: func(ctx context.Context, pageSize int, page int) ([]auto.UpdateSummary, error) {
			return []auto.UpdateSummary{{Version: 1}}, nil
		},
		ExportFunc: func(ctx context.Context) (apitype.UntypedDeployment, error) {
			return apitype.UntypedDeployment{Deployment: json.RawMessage(`{}`)}, nil
		},
		UpFunc: func(ctx context.Context, opts ...optup.Option) (auto.UpResult, error) {
			return auto.UpResult{
				StdOut: "up output",
				Summary: auto.UpdateSummary{
					ResourceChanges: &resourceChanges,
				},
			}, nil
		},
	}

	mockOperator := &MockStackOperator{
		SelectStackFunc: func(ctx context.Context, stackName, projectPath string) (RollbackStack, error) {
			return mockStack, nil
		},
	}

	var output bytes.Buffer
	opts := RollbackOptions{
		StackName:     "test",
		TargetVersion: 1,
		Operator:      mockOperator,
		Output:        &output,
	}

	result, err := ExecuteRollback(context.Background(), opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Success {
		t.Error("Expected Success to be true")
	}
	if result.ResourceChanges["create"] != 2 {
		t.Errorf("Expected ResourceChanges['create'] = 2, got %d", result.ResourceChanges["create"])
	}
}

func TestExecuteRollback_SelectStackError(t *testing.T) {
	mockOperator := &MockStackOperator{
		SelectStackFunc: func(ctx context.Context, stackName, projectPath string) (RollbackStack, error) {
			return nil, errors.New("stack not found")
		},
	}

	opts := RollbackOptions{
		StackName: "test",
		Operator:  mockOperator,
	}

	_, err := ExecuteRollback(context.Background(), opts)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestExecuteRollback_RefreshError(t *testing.T) {
	mockStack := &MockRollbackStack{
		HistoryFunc: func(ctx context.Context, pageSize int, page int) ([]auto.UpdateSummary, error) {
			return []auto.UpdateSummary{{Version: 1}}, nil
		},
		ExportFunc: func(ctx context.Context) (apitype.UntypedDeployment, error) {
			return apitype.UntypedDeployment{Deployment: json.RawMessage(`{}`)}, nil
		},
		RefreshFunc: func(ctx context.Context, opts ...optrefresh.Option) (auto.RefreshResult, error) {
			return auto.RefreshResult{}, errors.New("refresh failed")
		},
	}

	mockOperator := &MockStackOperator{
		SelectStackFunc: func(ctx context.Context, stackName, projectPath string) (RollbackStack, error) {
			return mockStack, nil
		},
	}

	var output bytes.Buffer
	opts := RollbackOptions{
		StackName:     "test",
		TargetVersion: 1,
		Operator:      mockOperator,
		Output:        &output,
	}

	_, err := ExecuteRollback(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for refresh failure")
	}
}

func TestExecuteRollback_UpError(t *testing.T) {
	mockStack := &MockRollbackStack{
		HistoryFunc: func(ctx context.Context, pageSize int, page int) ([]auto.UpdateSummary, error) {
			return []auto.UpdateSummary{{Version: 1}}, nil
		},
		ExportFunc: func(ctx context.Context) (apitype.UntypedDeployment, error) {
			return apitype.UntypedDeployment{Deployment: json.RawMessage(`{}`)}, nil
		},
		UpFunc: func(ctx context.Context, opts ...optup.Option) (auto.UpResult, error) {
			return auto.UpResult{}, errors.New("up failed")
		},
	}

	mockOperator := &MockStackOperator{
		SelectStackFunc: func(ctx context.Context, stackName, projectPath string) (RollbackStack, error) {
			return mockStack, nil
		},
	}

	var output bytes.Buffer
	opts := RollbackOptions{
		StackName:     "test",
		TargetVersion: 1,
		Operator:      mockOperator,
		Output:        &output,
	}

	_, err := ExecuteRollback(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for up failure")
	}
}

func TestExecuteRollback_NilResourceChanges(t *testing.T) {
	mockStack := &MockRollbackStack{
		HistoryFunc: func(ctx context.Context, pageSize int, page int) ([]auto.UpdateSummary, error) {
			return []auto.UpdateSummary{{Version: 1}}, nil
		},
		ExportFunc: func(ctx context.Context) (apitype.UntypedDeployment, error) {
			return apitype.UntypedDeployment{Deployment: json.RawMessage(`{}`)}, nil
		},
		UpFunc: func(ctx context.Context, opts ...optup.Option) (auto.UpResult, error) {
			return auto.UpResult{
				Summary: auto.UpdateSummary{
					ResourceChanges: nil,
				},
			}, nil
		},
	}

	mockOperator := &MockStackOperator{
		SelectStackFunc: func(ctx context.Context, stackName, projectPath string) (RollbackStack, error) {
			return mockStack, nil
		},
	}

	var output bytes.Buffer
	opts := RollbackOptions{
		StackName:     "test",
		TargetVersion: 1,
		Operator:      mockOperator,
		Output:        &output,
	}

	result, err := ExecuteRollback(context.Background(), opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result.ResourceChanges) != 0 {
		t.Errorf("Expected empty ResourceChanges, got %d entries", len(result.ResourceChanges))
	}
}

func TestGetCheckpointForVersion_HistoryError(t *testing.T) {
	mockStack := &MockRollbackStack{
		HistoryFunc: func(ctx context.Context, pageSize int, page int) ([]auto.UpdateSummary, error) {
			return nil, errors.New("history failed")
		},
	}

	_, err := GetCheckpointForVersion(context.Background(), mockStack, 1)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestGetCheckpointForVersion_ExportError(t *testing.T) {
	mockStack := &MockRollbackStack{
		HistoryFunc: func(ctx context.Context, pageSize int, page int) ([]auto.UpdateSummary, error) {
			return []auto.UpdateSummary{{Version: 1}}, nil
		},
		ExportFunc: func(ctx context.Context) (apitype.UntypedDeployment, error) {
			return apitype.UntypedDeployment{}, errors.New("export failed")
		},
	}

	_, err := GetCheckpointForVersion(context.Background(), mockStack, 1)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestGetCheckpointForVersion_InvalidDeployment(t *testing.T) {
	mockStack := &MockRollbackStack{
		HistoryFunc: func(ctx context.Context, pageSize int, page int) ([]auto.UpdateSummary, error) {
			return []auto.UpdateSummary{{Version: 1}}, nil
		},
		ExportFunc: func(ctx context.Context) (apitype.UntypedDeployment, error) {
			return apitype.UntypedDeployment{Deployment: json.RawMessage(`{invalid}`)}, nil
		},
	}

	_, err := GetCheckpointForVersion(context.Background(), mockStack, 1)
	if err == nil {
		t.Error("Expected error for invalid deployment")
	}
}

func TestRollbackOptions(t *testing.T) {
	opts := RollbackOptions{
		ProjectPath:   "/path/to/project",
		StackName:     "test-stack",
		TargetVersion: 5,
		DryRun:        true,
		Verbose:       true,
		Output:        nil,
	}

	if opts.ProjectPath != "/path/to/project" {
		t.Errorf("Expected ProjectPath to be '/path/to/project', got %q", opts.ProjectPath)
	}
	if opts.StackName != "test-stack" {
		t.Errorf("Expected StackName to be 'test-stack', got %q", opts.StackName)
	}
	if opts.TargetVersion != 5 {
		t.Errorf("Expected TargetVersion to be 5, got %d", opts.TargetVersion)
	}
}

func TestRollbackResult(t *testing.T) {
	result := RollbackResult{
		Success: true,
		Message: "test message",
		ResourceChanges: map[string]int{
			"create": 1,
		},
		Stdout: "stdout",
		Stderr: "stderr",
	}

	if !result.Success {
		t.Error("Expected Success to be true")
	}
	if result.Message != "test message" {
		t.Errorf("Expected Message to be 'test message', got %q", result.Message)
	}
}
