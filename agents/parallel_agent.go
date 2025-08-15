package agents

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/KennethanCeyer/adk-go/agents/interfaces"
	"github.com/KennethanCeyer/adk-go/llmproviders"
	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
	"github.com/KennethanCeyer/adk-go/tools"
)

// ParallelAgent executes a list of sub-agents concurrently and synthesizes their results.
type ParallelAgent struct {
	AgentName         string
	AgentDescription  string
	SubAgents         []interfaces.LlmAgent
	Provider          llmproviders.LLMProvider
	ModelID           string
	SysInstruction    *modelstypes.Message
}

// NewParallelAgent creates a new ParallelAgent.
func NewParallelAgent(name, description, modelID string, systemInstruction *modelstypes.Message, provider llmproviders.LLMProvider, subAgents []interfaces.LlmAgent) *ParallelAgent {
	return &ParallelAgent{
		AgentName:         name,
		AgentDescription:  description,
		SubAgents:         subAgents,
		Provider:          provider,
		ModelID:           modelID,
		SysInstruction:    systemInstruction,
	}
}

func (a *ParallelAgent) GetName() string                            { return a.AgentName }
func (a *ParallelAgent) GetDescription() string                     { return a.AgentDescription }
func (a *ParallelAgent) GetModelIdentifier() string                 { return a.ModelID }
func (a *ParallelAgent) GetSystemInstruction() *modelstypes.Message { return a.SysInstruction }
func (a *ParallelAgent) GetTools() []tools.Tool                     { return nil } // The parallel agent itself doesn't have tools, its sub-agents do.
func (a *ParallelAgent) GetLLMProvider() llmproviders.LLMProvider   { return a.Provider }

// Process executes sub-agents concurrently and then uses an LLM to synthesize the results.
func (a *ParallelAgent) Process(
	ctx context.Context,
	history []modelstypes.Message,
	latestContent modelstypes.Message,
) (*modelstypes.Message, error) {
	var wg sync.WaitGroup
	resultsChan := make(chan *modelstypes.Message, len(a.SubAgents))
	errChan := make(chan error, len(a.SubAgents))

	for _, subAgent := range a.SubAgents {
		wg.Add(1)
		go func(sa interfaces.LlmAgent) {
			defer wg.Done()
			log.Printf("--- Running sub-agent in parallel: %s ---\n", sa.GetName())
			// Each sub-agent gets the same initial history and input.
			response, err := sa.Process(ctx, history, latestContent)
			if err != nil {
				errChan <- fmt.Errorf("sub-agent '%s' failed: %w", sa.GetName(), err)
				return
			}
			resultsChan <- response
		}(subAgent)
	}

	wg.Wait()
	close(resultsChan)
	close(errChan)

	if len(errChan) > 0 {
		return nil, <-errChan
	}

	var subAgentResults []string
	for result := range resultsChan {
		if result != nil {
			for _, part := range result.Parts {
				if part.Text != nil {
					subAgentResults = append(subAgentResults, *part.Text)
				}
			}
		}
	}

	log.Println("--- Synthesizing results from parallel sub-agents ---")
	synthesisPromptText := fmt.Sprintf("The following information was gathered concurrently:\n\n---\n%s\n---\n\nBased on this information, provide a comprehensive summary to the user.", strings.Join(subAgentResults, "\n---\n"))
	synthesisMessage := modelstypes.Message{Role: "user", Parts: []modelstypes.Part{{Text: &synthesisPromptText}}}

	return a.Provider.GenerateContent(ctx, a.ModelID, a.SysInstruction, nil, nil, synthesisMessage)
}
