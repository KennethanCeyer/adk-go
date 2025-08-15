package agents

import (
	"context"
	"fmt"
	"log"

	"github.com/KennethanCeyer/adk-go/agents/interfaces"
	"github.com/KennethanCeyer/adk-go/llmproviders"
	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
	"github.com/KennethanCeyer/adk-go/tools"
)

// SequentialAgent executes a list of sub-agents in order.
type SequentialAgent struct {
	AgentName        string
	AgentDescription string
	SubAgents        []interfaces.LlmAgent
}

// NewSequentialAgent creates a new SequentialAgent.
func NewSequentialAgent(name, description string, subAgents []interfaces.LlmAgent) *SequentialAgent {
	return &SequentialAgent{
		AgentName:        name,
		AgentDescription: description,
		SubAgents:        subAgents,
	}
}

func (a *SequentialAgent) GetName() string                            { return a.AgentName }
func (a *SequentialAgent) GetDescription() string                     { return a.AgentDescription }
func (a *SequentialAgent) GetModelIdentifier() string                 { return "workflow-sequential" } // This agent orchestrates, it doesn't have its own model.
func (a *SequentialAgent) GetSystemInstruction() *modelstypes.Message { return nil }
func (a *SequentialAgent) GetTools() []tools.Tool                     { return nil }
func (a *SequentialAgent) GetLLMProvider() llmproviders.LLMProvider   { return nil }

// Process executes sub-agents sequentially, passing the output of one as the input to the next.
func (a *SequentialAgent) Process(
	ctx context.Context,
	history []modelstypes.Message,
	latestContent modelstypes.Message,
) (*modelstypes.Message, error) {
	currentContent := latestContent
	var finalResponse *modelstypes.Message

	for i, subAgent := range a.SubAgents {
		log.Printf("--- Running sub-agent (%d/%d): %s ---\n", i+1, len(a.SubAgents), subAgent.GetName())
		response, err := subAgent.Process(ctx, history, currentContent)
		if err != nil { return nil, fmt.Errorf("sub-agent '%s' failed: %w", subAgent.GetName(), err) }
		if response != nil { currentContent = *response; finalResponse = response }
	}
	return finalResponse, nil
}
