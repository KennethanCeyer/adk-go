package agents

import (
	"context"
	"fmt"
	"sync"

	"github.com/KennethanCeyer/adk-go/agents/invocation"

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

// GetModelIdentifier returns the model identifier (e.g., "gemini-2.5-flash").
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

// Process handles the main interaction loop with the LLM, including tool execution.
// It will continue to call the LLM with tool results until a final text response is generated.
func (a *BaseLlmAgent) Process(
	ctx context.Context,
	history []modelstypes.Message,
	latestMessage modelstypes.Message,
) (*modelstypes.Message, error) {
	if a.llmProvider == nil {
		return nil, fmt.Errorf("agent '%s' has no LLM provider configured", a.name)
	}

	turnHistory := make([]modelstypes.Message, len(history))
	copy(turnHistory, history)

	currentMessage := latestMessage

	// Loop for potential tool calls. Set a max number of tool calls per turn to avoid infinite loops.
	const maxToolCalls = 10 // Increased to support the looping_guesser example
	for i := 0; i < maxToolCalls; i++ {
		llmResponse, err := a.llmProvider.GenerateContent(
			ctx,
			a.modelIdentifier,
			a.systemInstruction,
			a.GetTools(),
			turnHistory,
			currentMessage,
		)
		if err != nil {
			return nil, fmt.Errorf("LLM interaction failed: %w", err)
		}

		turnHistory = append(turnHistory, currentMessage)
		if llmResponse != nil {
			turnHistory = append(turnHistory, *llmResponse)
		} else {
			// Should not happen, but handle defensively
			return nil, fmt.Errorf("LLM returned a nil response")
		}

		// A single model response can request multiple tool calls.
		var functionCalls []*modelstypes.FunctionCall
		for _, part := range llmResponse.Parts {
			if part.FunctionCall != nil {
				functionCalls = append(functionCalls, part.FunctionCall)
			}
		}

		// If there are no function calls, the agent's turn is over. Return the text response.
		if len(functionCalls) == 0 {
			return llmResponse, nil
		}

		var wg sync.WaitGroup
		toolResponseParts := make(chan modelstypes.Part, len(functionCalls))

		invocation.SendInternalLog(ctx, "Agent '%s' is calling %d tools in parallel...", a.name, len(functionCalls))

		for _, fc := range functionCalls {
			wg.Add(1)
			go func(call *modelstypes.FunctionCall) {
				defer wg.Done()
				invocation.SendInternalLog(ctx, "  - Calling tool '%s'", call.Name)
				toolToExecute, found := a.tools[call.Name]
				var responsePart modelstypes.Part

				if !found {
					errText := fmt.Sprintf("tool '%s' not found", call.Name)
					invocation.SendInternalLog(ctx, "  - Error: %s", errText)
					responsePart = modelstypes.Part{FunctionResponse: &modelstypes.FunctionResponse{Name: call.Name, Response: map[string]any{"error": errText}}}
				} else {
					toolResult, err := toolToExecute.Execute(ctx, call.Args)
					if err != nil {
						errText := fmt.Sprintf("tool '%s' execution failed: %v", toolToExecute.Name(), err)
						invocation.SendInternalLog(ctx, "  - Error: %s", errText)
						responsePart = modelstypes.Part{FunctionResponse: &modelstypes.FunctionResponse{Name: call.Name, Response: map[string]any{"error": errText}}}
					} else {
						invocation.SendInternalLog(ctx, "  - Tool '%s' executed successfully", toolToExecute.Name())
						responsePart = modelstypes.Part{FunctionResponse: &modelstypes.FunctionResponse{Name: call.Name, Response: toolResult}}
					}
				}
				toolResponseParts <- responsePart
			}(fc)
		}

		wg.Wait()
		close(toolResponseParts)

		var collectedParts []modelstypes.Part
		for part := range toolResponseParts {
			collectedParts = append(collectedParts, part)
		}

		toolResponseMessage := modelstypes.Message{Role: "function", Parts: collectedParts}

		currentMessage = toolResponseMessage
	}

	return nil, fmt.Errorf("exceeded maximum tool calls (%d) in a single turn", maxToolCalls)
}
