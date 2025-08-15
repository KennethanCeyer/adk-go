package agents

import (
	"context"
	"fmt"

	"github.com/KennethanCeyer/adk-go/agents/interfaces"
	"github.com/KennethanCeyer/adk-go/agents/invocation"
	"github.com/KennethanCeyer/adk-go/common"
	"github.com/KennethanCeyer/adk-go/llmproviders"
	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
	"github.com/KennethanCeyer/adk-go/tools"
)

// LoopAgent executes a sub-agent repeatedly until a stop condition is met or max loops are reached.
type LoopAgent struct {
	AgentName        string
	AgentDescription string
	SubAgent         interfaces.LlmAgent
	MaxLoops         int
	// StopCondition is a function that returns true if the loop should terminate.
	StopCondition func(latestResponse *modelstypes.Message) bool
}

// NewLoopAgent creates a new LoopAgent.
func NewLoopAgent(name, description string, subAgent interfaces.LlmAgent, maxLoops int, stopCondition func(*modelstypes.Message) bool) *LoopAgent {
	if stopCondition == nil {
		// Default stop condition is to never stop early, only by MaxLoops.
		stopCondition = func(latestResponse *modelstypes.Message) bool { return false }
	}
	return &LoopAgent{
		AgentName:        name,
		AgentDescription: description,
		SubAgent:         subAgent,
		MaxLoops:         maxLoops,
		StopCondition:    stopCondition,
	}
}

func (a *LoopAgent) GetName() string                            { return a.AgentName }
func (a *LoopAgent) GetDescription() string                     { return a.AgentDescription }
func (a *LoopAgent) GetModelIdentifier() string                 { return "workflow-loop" }
func (a *LoopAgent) GetSystemInstruction() *modelstypes.Message { return nil }
func (a *LoopAgent) GetTools() []tools.Tool                     { return nil }
func (a *LoopAgent) GetLLMProvider() llmproviders.LLMProvider   { return nil }

// Process executes the sub-agent in a loop.
func (a *LoopAgent) Process(
	ctx context.Context,
	history []modelstypes.Message,
	latestContent modelstypes.Message,
) (*modelstypes.Message, error) {
	loopHistory := make([]modelstypes.Message, len(history))
	copy(loopHistory, history)

	currentMessage := latestContent
	var finalResponse *modelstypes.Message

	invocation.SendInternalLog(ctx, "Starting loop for agent '%s' (max %d iterations)...", a.SubAgent.GetName(), a.MaxLoops)
	for i := 0; i < a.MaxLoops; i++ {
		invocation.SendInternalLog(ctx, "Loop iteration %d/%d", i+1, a.MaxLoops)

		// The sub-agent's Process method will handle its own tool calls internally and return a final response for the turn.
		response, err := a.SubAgent.Process(ctx, loopHistory, currentMessage)
		if err != nil {
			return nil, fmt.Errorf("sub-agent '%s' in loop failed on iteration %d: %w", a.SubAgent.GetName(), i+1, err)
		}

		if !common.IsMessageEmpty(currentMessage) {
			loopHistory = append(loopHistory, currentMessage)
		}
		if response != nil {
			loopHistory = append(loopHistory, *response)
		}

		finalResponse = response

		if a.StopCondition(finalResponse) {
			invocation.SendInternalLog(ctx, "Loop stop condition met on iteration %d.", i+1)
			break
		}

		if finalResponse != nil {
			currentMessage = *finalResponse
		} else {
			// If sub-agent returns nothing, we can't continue.
			return nil, fmt.Errorf("loop agent '%s' sub-agent returned a nil response on iteration %d", a.AgentName, i+1)
		}
	}

	if finalResponse == nil {
		return nil, fmt.Errorf("loop agent '%s' finished after %d iterations without producing a final response", a.AgentName, a.MaxLoops)
	}

	return finalResponse, nil
}
