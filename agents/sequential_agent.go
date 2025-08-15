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

type SequentialAgent struct {
	AgentName        string
	AgentDescription string
	SubAgents        []interfaces.LlmAgent
}

func NewSequentialAgent(name, description string, subAgents []interfaces.LlmAgent) *SequentialAgent {
	return &SequentialAgent{
		AgentName:        name,
		AgentDescription: description,
		SubAgents:        subAgents,
	}
}

func (a *SequentialAgent) GetName() string                            { return a.AgentName }
func (a *SequentialAgent) GetDescription() string                     { return a.AgentDescription }
func (a *SequentialAgent) GetModelIdentifier() string                 { return "workflow-sequential" }
func (a *SequentialAgent) GetSystemInstruction() *modelstypes.Message { return nil }
func (a *SequentialAgent) GetTools() []tools.Tool                     { return nil }
func (a *SequentialAgent) GetLLMProvider() llmproviders.LLMProvider   { return nil }

func (a *SequentialAgent) Process(
	ctx context.Context,
	history []modelstypes.Message,
	latestContent modelstypes.Message,
) (*modelstypes.Message, error) {
	invocation.SendInternalLog(ctx, "Starting sequential execution for %d sub-agents...", len(a.SubAgents))

	currentHistory := history
	currentMessage := latestContent
	var finalResponse *modelstypes.Message

	for _, subAgent := range a.SubAgents {
		invocation.SendInternalLog(ctx, "Running sub-agent in sequence: %s", subAgent.GetName())
		response, err := subAgent.Process(ctx, currentHistory, currentMessage)
		if err != nil {
			return nil, fmt.Errorf("sub-agent '%s' failed: %w", subAgent.GetName(), err)
		}

		currentHistory = append(currentHistory, currentMessage)
		if response != nil {
			currentHistory = append(currentHistory, *response)
			currentMessage = *response
		}
		finalResponse = response
	}

	invocation.SendInternalLog(ctx, "Sequential execution finished.")
	return finalResponse, nil
}
