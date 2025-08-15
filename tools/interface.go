package tools

import "context"

type Tool interface {
	Name() string
	Description() string
	Parameters() any
	Execute(ctx context.Context, args any) (any, error)
}
