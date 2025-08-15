package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/KennethanCeyer/adk-go/agents/callbacks"
	"github.com/KennethanCeyer/adk-go/agents/invocation"

	"github.com/KennethanCeyer/adk-go/agents/interfaces"
	"github.com/KennethanCeyer/adk-go/llmproviders"
	"github.com/KennethanCeyer/adk-go/models"
	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
	"github.com/KennethanCeyer/adk-go/tools"
)

// BaseLlmAgent provides a foundational implementation of the LlmAgent interface,
// handling the core logic of interacting with an LLM, including tool use.
type BaseLlmAgent struct {
	name              string
	description       string
	modelIdentifier   string
	systemInstruction *modelstypes.Message
	llmProvider       llmproviders.LLMProvider
	tools             map[string]tools.Tool

	// Callbacks
	BeforeAgentCallback  callbacks.BeforeAgentCallback
	AfterAgentCallback   callbacks.AfterAgentCallback
	BeforeModelCallback  callbacks.BeforeModelCallback
	AfterModelCallback   callbacks.AfterModelCallback
	BeforeToolCallback   callbacks.BeforeToolCallback
	AfterToolCallback    callbacks.AfterToolCallback
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

// Process handles the main interaction loop with the LLM, including tool execution,
// until a final text response is generated.
func (a *BaseLlmAgent) Process(
	ctx context.Context,
	history []modelstypes.Message,
	latestMessage modelstypes.Message,
) (*modelstypes.Message, error) {
	if a.llmProvider == nil {
		return nil, fmt.Errorf("agent '%s' has no LLM provider configured", a.name)
	}

	callbackCtx := &callbacks.CallbackContext{
		AgentName: a.GetName(),
		InvocationID: invocation.FromContext(ctx).ID,
		SessionState: nil, // TODO: Implement session state access
		UserContent:  &latestMessage,
	}

	// Execute BeforeAgentCallback
	if a.BeforeAgentCallback != nil {
		if overrideResponse := a.BeforeAgentCallback(callbackCtx); overrideResponse != nil {
			return overrideResponse, nil
		}
	}

	turnHistory := make([]modelstypes.Message, len(history))
	copy(turnHistory, history)

	currentMessage := latestMessage

	// Loop for potential tool calls. Set a max number of calls per turn to avoid infinite loops.
	const maxToolCalls = 10
	for i := 0; i < maxToolCalls; i++ {
		llmReq := &models.LlmRequest{
			ModelIdentifier:   a.modelIdentifier,
			SystemInstruction: a.systemInstruction,
			Tools:             a.GetTools(),
			History:           turnHistory,
			LatestMessage:     currentMessage,
		}

		// Execute BeforeModelCallback
		var llmResponse *models.LlmResponse
		if a.BeforeModelCallback != nil {
			llmResponse = a.BeforeModelCallback(callbackCtx, llmReq)
		}

		if llmResponse != nil {
			// Callback provided a response, skip actual LLM call
			invocation.SendInternalLog(ctx, "Agent '%s' model call was overridden by a callback.", a.name)
		} else {
			// No override, call the LLM provider
			llmResponseMsg, err := a.llmProvider.GenerateContent(
				ctx,
				llmReq.ModelIdentifier,
				llmReq.SystemInstruction,
				llmReq.Tools,
				llmReq.History,
				llmReq.LatestMessage,
			)
			if err != nil {
				return nil, fmt.Errorf("LLM interaction failed: %w", err)
			}
			llmResponse = &models.LlmResponse{Content: llmResponseMsg}
		}

		// Execute AfterModelCallback
		if a.AfterModelCallback != nil {
			if modifiedResponse := a.AfterModelCallback(callbackCtx, llmResponse); modifiedResponse != nil {
				llmResponse = modifiedResponse
			}
		}

		if llmResponse.Content == nil {
			return nil, fmt.Errorf("LLM returned a nil response content")
		}

		turnHistory = append(turnHistory, currentMessage)
		turnHistory = append(turnHistory, *llmResponse.Content)

		// A single model response can request multiple tool calls.
		var functionCalls []*modelstypes.FunctionCall
		for _, part := range llmResponse.Content.Parts {
			if part.FunctionCall != nil {
				functionCalls = append(functionCalls, part.FunctionCall)
			}
		}

		// If there are no function calls, the agent's turn is over. Return the text response.
		if len(functionCalls) == 0 {
			// Execute AfterAgentCallback
			if a.AfterAgentCallback != nil {
				if finalResponse := a.AfterAgentCallback(callbackCtx, llmResponse.Content); finalResponse != nil {
					return finalResponse, nil
				}
			}
			return llmResponse.Content, nil
		}

		var wg sync.WaitGroup
		toolResponseParts := make(chan modelstypes.Part, len(functionCalls))

		invocation.SendInternalLog(ctx, "Agent '%s' is calling %d tools in parallel...", a.name, len(functionCalls))

		for _, fc := range functionCalls {
			wg.Add(1)
			go func(call *modelstypes.FunctionCall) {
				defer wg.Done()
				argsStr := ""
				if call.Args != nil && len(call.Args) > 0 {
					argsBytes, err := json.Marshal(call.Args)
					if err == nil {
						argsStr = fmt.Sprintf(" with args: %s", string(argsBytes))
					} else {
						argsStr = fmt.Sprintf(" with args: %v", call.Args)
					}
				}
				invocation.SendInternalLog(ctx, "  - Calling tool '%s'%s", call.Name, argsStr)
				toolToExecute, found := a.tools[call.Name]
				var responsePart modelstypes.Part

				if !found {
					errText := fmt.Sprintf("tool '%s' not found", call.Name)
					invocation.SendInternalLog(ctx, "  - Error: %s", errText)
					responsePart = modelstypes.Part{FunctionResponse: &modelstypes.FunctionResponse{Name: call.Name, Response: map[string]any{"error": errText}}}
				} else {
					// Execute BeforeToolCallback
					if a.BeforeToolCallback != nil {
						if modifiedArgs := a.BeforeToolCallback(callbackCtx, toolToExecute, call.Args); modifiedArgs != nil {
							call.Args = modifiedArgs
						}
					}

					toolResult, err := toolToExecute.Execute(ctx, call.Args)
					if err != nil {
						errText := fmt.Sprintf("tool '%s' execution failed: %v", toolToExecute.Name(), err)
						invocation.SendInternalLog(ctx, "  - Error: %s", errText)
						responsePart = modelstypes.Part{FunctionResponse: &modelstypes.FunctionResponse{Name: call.Name, Response: map[string]any{"error": errText}}}
					} else {
						// The result from a tool must be a map[string]any to be used in the FunctionResponse.
						toolResultMap, ok := toolResult.(map[string]any)
						if !ok {
							errText := fmt.Sprintf("tool '%s' result is not a map[string]any, but %T", toolToExecute.Name(), toolResult)
							invocation.SendInternalLog(ctx, "  - Error: %s", errText)
							responsePart = modelstypes.Part{FunctionResponse: &modelstypes.FunctionResponse{Name: call.Name, Response: map[string]any{"error": errText}}}
						} else {
							// Execute AfterToolCallback
							if a.AfterToolCallback != nil {
								if modifiedResult := a.AfterToolCallback(callbackCtx, toolToExecute, call.Args, toolResultMap); modifiedResult != nil {
									toolResultMap = modifiedResult
								}
							}
							invocation.SendInternalLog(ctx, "  - Tool '%s' executed successfully", toolToExecute.Name())
							responsePart = modelstypes.Part{FunctionResponse: &modelstypes.FunctionResponse{Name: call.Name, Response: toolResultMap}}
						}
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

	// If loop finishes, it means max tool calls exceeded. Before returning error, call AfterAgentCallback.
	if a.AfterAgentCallback != nil {
		// Pass nil as there's no final successful response
		if finalResponse := a.AfterAgentCallback(callbackCtx, nil); finalResponse != nil {
			return finalResponse, nil
		}
	}
	return nil, fmt.Errorf("exceeded maximum tool calls (%d) in a single turn", maxToolCalls)
}
