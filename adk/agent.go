package adk

import (
	"context"
	"fmt"

	adktypes "github.com/KennethanCeyer/adk-go/adk/types"
)

type Tool interface {
	Name() string
	Description() string
	Parameters() any
	Execute(ctx context.Context, args any) (any, error)
}

type LLMProvider interface {
	GenerateContent(
		ctx context.Context,
		modelName string,
		systemInstruction string,
		tools []Tool,
		history []adktypes.Message,
		latestMessage adktypes.Message,
	) (adktypes.Message, error)
}

type Agent interface {
	Name() string
	SystemInstruction() string
	Tools() []Tool
	ModelName() string
	LLMProvider() LLMProvider
	Process(ctx context.Context, history []adktypes.Message, latestMessage adktypes.Message) (adktypes.Message, error)
}

type BaseAgent struct {
	AgentNameField              string
	AgentSystemInstructionField string
	AgentToolsField             []Tool
	AgentModelNameField         string
	AgentLLMProviderField       LLMProvider
}

func NewBaseAgent(name, systemInstruction, modelName string, llmProvider LLMProvider, tools []Tool) *BaseAgent {
	if tools == nil {
		tools = []Tool{}
	}
	return &BaseAgent{
		AgentNameField:              name,
		AgentSystemInstructionField: systemInstruction,
		AgentModelNameField:         modelName,
		AgentLLMProviderField:       llmProvider,
		AgentToolsField:             tools,
	}
}

func (a *BaseAgent) Name() string              { return a.AgentNameField }
func (a *BaseAgent) SystemInstruction() string { return a.AgentSystemInstructionField }
func (a *BaseAgent) Tools() []Tool             { return a.AgentToolsField }
func (a *BaseAgent) ModelName() string         { return a.AgentModelNameField }
func (a *BaseAgent) LLMProvider() LLMProvider  { return a.AgentLLMProviderField }

func (a *BaseAgent) Process(ctx context.Context, history []adktypes.Message, latestMessage adktypes.Message) (adktypes.Message, error) {
	if a.AgentLLMProviderField == nil {
		return adktypes.Message{}, fmt.Errorf("agent %s has no LLMProvider configured", a.AgentNameField)
	}

	llmResponseMsg, err := a.AgentLLMProviderField.GenerateContent(
		ctx,
		a.AgentModelNameField,
		a.AgentSystemInstructionField,
		a.AgentToolsField,
		history,
		latestMessage,
	)
	if err != nil {
		return adktypes.Message{}, fmt.Errorf("LLMProvider.GenerateContent failed for agent %s: %w", a.AgentNameField, err)
	}

	var pendingFunctionCall *adktypes.FunctionCall
	for _, part := range llmResponseMsg.Parts {
		if part.FunctionCall != nil {
			pendingFunctionCall = part.FunctionCall
			break
		}
	}

	if pendingFunctionCall != nil {
		var selectedTool Tool
		for _, t := range a.AgentToolsField {
			if t.Name() == pendingFunctionCall.Name {
				selectedTool = t
				break
			}
		}

		var toolResponseMessage adktypes.Message
		if selectedTool == nil {
			errText := fmt.Sprintf("LLM requested unknown tool '%s'", pendingFunctionCall.Name)
			toolResponseMessage = adktypes.Message{
				Role: "user",
				Parts: []adktypes.Part{{
					FunctionResponse: &adktypes.FunctionResponse{
						Name:     pendingFunctionCall.Name,
						Response: map[string]any{"error": errText},
					},
				}},
			}
		} else {
			toolResultData, toolErr := selectedTool.Execute(ctx, pendingFunctionCall.Args)
			toolExecutionResponse := adktypes.FunctionResponse{Name: selectedTool.Name()}
			if toolErr != nil {
				toolExecutionResponse.Response = map[string]any{"error": toolErr.Error()}
			} else {
				toolExecutionResponse.Response = toolResultData.(map[string]any)
			}
			toolResponseMessage = adktypes.Message{
				Role:  "user",
				Parts: []adktypes.Part{{FunctionResponse: &toolExecutionResponse}},
			}
		}

		updatedHistory := append(history, latestMessage, llmResponseMsg)

		finalLlmResponseMsg, err := a.AgentLLMProviderField.GenerateContent(
			ctx,
			a.AgentModelNameField,
			a.AgentSystemInstructionField,
			a.AgentToolsField,
			updatedHistory,
			toolResponseMessage,
		)
		if err != nil {
			return adktypes.Message{}, fmt.Errorf("LLMProvider.GenerateContent after tool call failed for agent %s: %w", a.AgentNameField, err)
		}
		return finalLlmResponseMsg, nil
	}
	return llmResponseMsg, nil
}
