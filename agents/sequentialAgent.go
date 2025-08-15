package agents

import (
	"context"
	"fmt"
	"log"

	"github.com/KennethanCeyer/adk-go/llmproviders"
	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
	"github.com/KennethanCeyer/adk-go/tools"
)

// SequentialAgent executes a list of sub-agents in order.
type SequentialAgent struct {
	AgentName        string
	AgentDescription string
	SubAgents        []LlmAgent
}

// NewSequentialAgent creates a new SequentialAgent.
func NewSequentialAgent(name, description string, subAgents []LlmAgent) *SequentialAgent {
	return &SequentialAgent{
		AgentName:        name,
		AgentDescription: description,
		SubAgents:        subAgents,
	}
}

func (a *SequentialAgent) GetName() string                        { return a.AgentName }
func (a *SequentialAgent) GetDescription() string                 { return a.AgentDescription }
func (a *SequentialAgent) GetModelIdentifier() string             { return "workflow-sequential" }
func (a *SequentialAgent) GetSystemInstruction() *modelstypes.Message { return nil }
func (a *SequentialAgent) GetTools() []tools.Tool                 { return nil }
func (a *SequentialAgent) GetLLMProvider() llmproviders.LLMProvider { return nil }

// Process executes sub-agents sequentially.
func (a *SequentialAgent) Process(
	ctx context.Context,
	history []modelstypes.Message,
	latestContent modelstypes.Message,
) (*modelstypes.Message, error) {
	var finalResponse *modelstypes.Message
	currentHistory := history
	currentInput := latestContent

	for i, subAgent := range a.SubAgents {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		log.Printf("--- Running sub-agent (%d/%d): %s ---\n", i+1, len(a.SubAgents), subAgent.GetName())
		response, err := subAgent.Process(ctx, currentHistory, currentInput)
		if err != nil {
			return nil, fmt.Errorf("sub-agent '%s' failed: %w", subAgent.GetName(), err)
		}

		// The history for the next agent includes the previous agent's interaction.
		currentHistory = append(currentHistory, currentInput)
		if response != nil {
			currentHistory = append(currentHistory, *response)
		}

		// The output of one agent becomes the input for the next.
		// For simplicity, we'll just pass the response content.
		// A more advanced implementation might use state management.
		if response != nil {
			currentInput = *response
		} else {
			// If an agent gives no response, we can't proceed with it as input.
			// We'll just pass an empty message to the next agent.
			currentInput = modelstypes.Message{Role: "user", Parts: []modelstypes.Part{}}
		}
		finalResponse = response
	}

	// The final response of the sequential agent is the response of the last sub-agent.
	return finalResponse, nil
}
