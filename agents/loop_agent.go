package agents

import (
	"context"
	"fmt"

	"github.com/KennethanCeyer/adk-go/agents/interfaces"
	"github.com/KennethanCeyer/adk-go/agents/invocation"
	"github.com/KennethanCeyer/adk-go/llmproviders"
	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
	"github.com/KennethanCeyer/adk-go/tools"
)

// StopCondition is a function that determines if a loop should stop.
type StopCondition func(latestResponse *modelstypes.Message) bool

// LoopAgent executes its sub-agents repeatedly until a condition is met.
type LoopAgent struct {
	AgentName        string
	AgentDescription string
	SubAgents        []interfaces.LlmAgent
	MaxIterations    int
	StopWhen         StopCondition
}

// NewLoopAgent creates a new LoopAgent.
func NewLoopAgent(name, description string, subAgents []interfaces.LlmAgent, maxIterations int, stopWhen StopCondition) *LoopAgent {
	return &LoopAgent{
		AgentName:        name,
		AgentDescription: description,
		SubAgents:        subAgents,
		MaxIterations:    maxIterations,
		StopWhen:         stopWhen,
	}
}

func (a *LoopAgent) GetName() string                            { return a.AgentName }
func (a *LoopAgent) GetDescription() string                     { return a.AgentDescription }
func (a *LoopAgent) GetModelIdentifier() string                 { return "workflow-loop" }
func (a *LoopAgent) GetSystemInstruction() *modelstypes.Message { return nil }
func (a *LoopAgent) GetTools() []tools.Tool                     { return nil }
func (a *LoopAgent) GetLLMProvider() llmproviders.LLMProvider   { return nil }

// Process executes sub-agents in a loop.
func (a *LoopAgent) Process(
	ctx context.Context,
	history []modelstypes.Message,
	latestContent modelstypes.Message,
) (*modelstypes.Message, error) {
	invocation.SendInternalLog(ctx, "Starting loop for agent '%s' (max %d iterations)...", a.GetName(), a.MaxIterations)

	currentHistory := history
	currentMessage := latestContent
	var finalResponse *modelstypes.Message

	for i := 0; i < a.MaxIterations; i++ {
		invocation.SendInternalLog(ctx, "Loop iteration %d/%d", i+1, a.MaxIterations)

		for _, subAgent := range a.SubAgents {
			response, err := subAgent.Process(ctx, currentHistory, currentMessage)
			if err != nil {
				return nil, fmt.Errorf("sub-agent '%s' in loop failed: %w", subAgent.GetName(), err)
			}
			currentHistory = append(currentHistory, currentMessage)
			if response != nil {
				currentHistory = append(currentHistory, *response)
				currentMessage = *response
			}
			finalResponse = response
		}

		if a.StopWhen != nil && a.StopWhen(finalResponse) {
			invocation.SendInternalLog(ctx, "Loop stop condition met on iteration %d.", i+1)
			break
		}
	}
	return finalResponse, nil
}
