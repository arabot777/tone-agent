package function

import (
	"context"

	"tone/agent/pkg/common/app/internal/container"
)

type functionBundle struct {
	name string
	fn   func(ctx context.Context) error
}

func (f *functionBundle) Run(ctx context.Context) error {
	return f.fn(ctx)
}

func (f *functionBundle) Type() string {
	return "Function"
}

func (f *functionBundle) Name() string {
	return f.name
}

func (f *functionBundle) Stop() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func NewFunctionBundle(name string, fn func(ctx context.Context) error) container.Bundle {
	return &functionBundle{
		name: name,
		fn:   fn,
	}
}
