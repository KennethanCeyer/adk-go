package invocation

import (
	"context"
	"fmt"

	"github.com/KennethanCeyer/adk-go/agents/interfaces"
)

type contextKey string

const (
	invocationContextKey = contextKey("invocationContext")
	uiSenderKey = contextKey("uiSender")
)

type InvocationContext struct {
	ID    string
	Agent interfaces.LlmAgent
}

func WithInvocationContext(ctx context.Context, invCtx *InvocationContext) context.Context {
	return context.WithValue(ctx, invocationContextKey, invCtx)
}

func FromContext(ctx context.Context) *InvocationContext {
	invCtx, ok := ctx.Value(invocationContextKey).(*InvocationContext)
	if !ok || invCtx == nil {
		return &InvocationContext{}
	}
	return invCtx
}

type UISender func(messageType string, payload any)

func WithUISender(ctx context.Context, sender UISender) context.Context {
	return context.WithValue(ctx, uiSenderKey, sender)
}

func GetUISender(ctx context.Context) (UISender, bool) {
	sender, ok := ctx.Value(uiSenderKey).(UISender)
	return sender, ok
}

func SendInternalLog(ctx context.Context, format string, a ...any) {
	if sender, ok := GetUISender(ctx); ok {
		text := fmt.Sprintf(format, a...)
		sender("internal_log", map[string]string{"text": text})
	}
}
