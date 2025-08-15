package processors

import (
	"github.com/KennethanCeyer/adk-go/agents/invocation"
	"github.com/KennethanCeyer/adk-go/events"
	"github.com/KennethanCeyer/adk-go/models"
)

// BaseLlmRequestProcessor defines the interface for processing LLM requests.
type BaseLlmRequestProcessor interface {
	RunAsync(invocationCtx *invocation.InvocationContext, llmReq *models.LlmRequest) (<-chan *events.Event, error)
}

// BaseLlmResponseProcessor defines the interface for processing LLM responses.
type BaseLlmResponseProcessor interface {
	RunAsync(invocationCtx *invocation.InvocationContext, llmResp *models.LlmResponse, modelResponseEvent *events.Event) (<-chan *events.Event, error)
}
