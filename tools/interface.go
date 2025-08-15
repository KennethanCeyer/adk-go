package tools

import "context"

// Tool is the interface that all tools must implement.
type Tool interface {
	Name() string
	Description() string
	Parameters() any
	Execute(ctx context.Context, args any) (any, error)
}
