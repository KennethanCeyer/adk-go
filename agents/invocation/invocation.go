package invocation

import (
	"context"

	"github.com/KennethanCeyer/adk-go/agents/interfaces"
	"github.com/KennethanCeyer/adk-go/common/types"
	"github.com/KennethanCeyer/adk-go/models"
	"github.com/KennethanCeyer/adk-go/sessions"
)

// RunConfig holds configuration for a specific agent execution.
type RunConfig struct {
	// Fields for run-specific configurations if needed.
}

// InvocationContext provides context for a single agent invocation.
type InvocationContext struct {
	Ctx          context.Context
	InvocationID types.InvocationID
	Agent        interfaces.Agent // The agent being invoked
	Session      *sessions.Session
	UserContent  *models.Content // Initial user input for this specific invocation
	RunConfig    *RunConfig
	Branch       types.BranchID
}

// ReadonlyContext offers a read-only view of InvocationContext.
type ReadonlyContext struct {
	InvocationCtx *InvocationContext
}

// NewReadonlyContext creates a ReadonlyContext.
func NewReadonlyContext(invCtx *InvocationContext) *ReadonlyContext {
	return &ReadonlyContext{InvocationCtx: invCtx}
}

// GetSession returns the session from the context.
func (roc *ReadonlyContext) GetSession() *sessions.Session {
	if roc.InvocationCtx == nil {
		return nil
	}
	return roc.InvocationCtx.Session
}
