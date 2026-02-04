// Copyright 2026 Pegasus Heavy Industries LLC
// Contact: pegasusheavyindustries@gmail.com

package history

import (
	"context"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

// StackSelector is an interface for selecting stacks
type StackSelector interface {
	SelectStack(ctx context.Context, stackName, projectPath string) (Stack, error)
}

// Stack is an interface for stack operations
type Stack interface {
	History(ctx context.Context, pageSize int, page int) ([]auto.UpdateSummary, error)
}

// DefaultStackSelector uses the real Pulumi SDK
type DefaultStackSelector struct{}

// SelectStack selects a stack using the Pulumi SDK
func (d *DefaultStackSelector) SelectStack(ctx context.Context, stackName, projectPath string) (Stack, error) {
	stack, err := auto.SelectStackLocalSource(ctx, stackName, projectPath)
	if err != nil {
		return nil, err
	}
	return &RealStack{stack: stack}, nil
}

// RealStack wraps a real Pulumi stack
type RealStack struct {
	stack auto.Stack
}

// History returns the stack history
func (r *RealStack) History(ctx context.Context, pageSize int, page int) ([]auto.UpdateSummary, error) {
	return r.stack.History(ctx, pageSize, page)
}

// DefaultSelector is the default stack selector using real Pulumi SDK
var DefaultSelector StackSelector = &DefaultStackSelector{}
