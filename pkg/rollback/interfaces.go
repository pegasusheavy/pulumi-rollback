// Copyright 2026 Pegasus Heavy Industries LLC
// Contact: pegasusheavyindustries@gmail.com

package rollback

import (
	"context"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optpreview"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optrefresh"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

// StackOperator is an interface for stack operations needed for rollback
type StackOperator interface {
	SelectStack(ctx context.Context, stackName, projectPath string) (RollbackStack, error)
}

// RollbackStack is an interface for stack operations needed for rollback
type RollbackStack interface {
	Export(ctx context.Context) (apitype.UntypedDeployment, error)
	Import(ctx context.Context, state apitype.UntypedDeployment) error
	History(ctx context.Context, pageSize int, page int) ([]auto.UpdateSummary, error)
	Preview(ctx context.Context, opts ...optpreview.Option) (auto.PreviewResult, error)
	Refresh(ctx context.Context, opts ...optrefresh.Option) (auto.RefreshResult, error)
	Up(ctx context.Context, opts ...optup.Option) (auto.UpResult, error)
}

// DefaultStackOperator uses the real Pulumi SDK
type DefaultStackOperator struct{}

// SelectStack selects a stack using the Pulumi SDK
func (d *DefaultStackOperator) SelectStack(ctx context.Context, stackName, projectPath string) (RollbackStack, error) {
	stack, err := auto.SelectStackLocalSource(ctx, stackName, projectPath)
	if err != nil {
		return nil, err
	}
	return &RealRollbackStack{stack: stack}, nil
}

// RealRollbackStack wraps a real Pulumi stack
type RealRollbackStack struct {
	stack auto.Stack
}

// Export exports the stack state
func (r *RealRollbackStack) Export(ctx context.Context) (apitype.UntypedDeployment, error) {
	return r.stack.Export(ctx)
}

// Import imports a stack state
func (r *RealRollbackStack) Import(ctx context.Context, state apitype.UntypedDeployment) error {
	return r.stack.Import(ctx, state)
}

// History returns the stack history
func (r *RealRollbackStack) History(ctx context.Context, pageSize int, page int) ([]auto.UpdateSummary, error) {
	return r.stack.History(ctx, pageSize, page)
}

// Preview runs a preview
func (r *RealRollbackStack) Preview(ctx context.Context, opts ...optpreview.Option) (auto.PreviewResult, error) {
	return r.stack.Preview(ctx, opts...)
}

// Refresh runs a refresh
func (r *RealRollbackStack) Refresh(ctx context.Context, opts ...optrefresh.Option) (auto.RefreshResult, error) {
	return r.stack.Refresh(ctx, opts...)
}

// Up runs an update
func (r *RealRollbackStack) Up(ctx context.Context, opts ...optup.Option) (auto.UpResult, error) {
	return r.stack.Up(ctx, opts...)
}

// DefaultOperator is the default stack operator using real Pulumi SDK
var DefaultOperator StackOperator = &DefaultStackOperator{}
