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

// BaseLlmAgent provides a foundational implementation of the LlmAgent interface.
// It handles the core logic of interacting with an LLM, including tool use.
type BaseLlmAgent struct {
	name              string
	description       string
	modelIdentifier   string
	systemInstruction *modelstypes.Message
	llmProvider       llmproviders.LLMProvider
	tools             map[string]tools.Tool
}

// NewBaseLlmAgent creates and initializes a new BaseLlmAgent.
func NewBaseLlmAgent(
	name string,
	description string,
	modelIdentifier string,
	systemInstruction *modelstypes.Message,
	provider llmproviders.LLMProvider,
	agentTools []tools.Tool,
) interfaces.LlmAgent {
	toolMap := make(map[string]tools.Tool)
	for _, t := range agentTools {
		if t != nil {
			toolMap[t.Name()] = t
		}
	}
	return &BaseLlmAgent{
		name:              name,
		description:       description,
		modelIdentifier:   modelIdentifier,
		systemInstruction: systemInstruction,
		llmProvider:       provider,
		tools:             toolMap,
	}
}

// GetName returns the agent's name.
func (a *BaseLlmAgent) GetName() string { return a.name }

// GetDescription returns the agent's description.
func (a *BaseLlmAgent) GetDescription() string { return a.description }

// GetModelIdentifier returns the model identifier (e.g., "gemini-1.5-flash-latest").
func (a *BaseLlmAgent) GetModelIdentifier() string { return a.modelIdentifier }

// GetSystemInstruction returns the system instruction message for the agent.
func (a *BaseLlmAgent) GetSystemInstruction() *modelstypes.Message { return a.systemInstruction }

// GetTools returns a slice of tools available to the agent.
func (a *BaseLlmAgent) GetTools() []tools.Tool {
	toolSlice := make([]tools.Tool, 0, len(a.tools))
	for _, t := range a.tools {
		toolSlice = append(toolSlice, t)
	}
	return toolSlice
}

// GetLLMProvider returns the LLM provider used by the agent.
func (a *BaseLlmAgent) GetLLMProvider() llmproviders.LLMProvider { return a.llmProvider }

// Process handles the main interaction loop with the LLM.
func (a *BaseLlmAgent) Process(
	ctx context.Context,
	history []modelstypes.Message,
	latestMessage modelstypes.Message,
) (*modelstypes.Message, error) {
	if a.llmProvider == nil {
		return nil, fmt.Errorf("agent '%s' has no LLM provider configured", a.name)
	}

	currentHistory := history
	currentMessage := latestMessage

	// Loop to handle potential tool calls. A limit is set to prevent infinite loops.
	const maxToolCalls = 5
	for i := 0; i < maxToolCalls; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		llmResponse, err := a.llmProvider.GenerateContent(
			ctx,
			a.modelIdentifier,
			a.systemInstruction,
			a.GetTools(),
			currentHistory,
			currentMessage,
		)
		if err != nil {
			return nil, fmt.Errorf("LLM interaction failed: %w", err)
		}

		if llmResponse == nil || len(llmResponse.Parts) == 0 {
			return nil, fmt.Errorf("LLM returned an empty response")
		}

		// Check if the response contains a function call.
		// We assume the first part will be the function call if one exists.
		if llmResponse.Parts[0].FunctionCall == nil {
			// No tool call, return the text response directly.
			return llmResponse, nil
		}

		// Append the LLM's request to use a tool and the current user message to the history.
		currentHistory = append(currentHistory, currentMessage, *llmResponse)

		// Handle the function call
		functionCall := llmResponse.Parts[0].FunctionCall
		tool, exists := a.tools[functionCall.Name]
		if !exists {
			errText := fmt.Sprintf("tool '%s' not found", functionCall.Name)
			log.Println(errText)
			// Prepare a message to send back to the LLM about the error
			currentMessage = modelstypes.Message{
				Role: "function",
				Parts: []modelstypes.Part{{
					FunctionResponse: &modelstypes.FunctionResponse{Name: functionCall.Name, Response: map[string]any{"error": errText}},
				}},
			}
			continue // Go back to the LLM with the error message
		}

		log.Printf("Agent '%s' executing tool '%s' with args: %v", a.name, tool.Name(), functionCall.Args)
		toolResult, err := tool.Execute(ctx, functionCall.Args)
		if err != nil {
			errText := fmt.Sprintf("tool '%s' execution failed: %v", tool.Name(), err)
			log.Println(errText)
			currentMessage = modelstypes.Message{
				Role: "function",
				Parts: []modelstypes.Part{{
					FunctionResponse: &modelstypes.FunctionResponse{Name: functionCall.Name, Response: map[string]any{"error": errText}},
				}},
			}
			continue // Go back to the LLM with the error message
		}

		// Prepare the tool's result to send back to the LLM.
		currentMessage = modelstypes.Message{
			Role:  "function",
			Parts: []modelstypes.Part{{FunctionResponse: &modelstypes.FunctionResponse{Name: functionCall.Name, Response: toolResult}}},
		}
		// Loop again to get the final text response from the LLM based on the tool result.
	}

	return nil, fmt.Errorf("agent exceeded maximum tool call iterations (%d)", maxToolCalls)
}
