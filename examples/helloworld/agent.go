package helloworld

import (
	"github.com/KennethanCeyer/adk-go/agents"
	"github.com/KennethanCeyer/adk-go/examples"
	"github.com/KennethanCeyer/adk-go/llmproviders"
	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
	"github.com/KennethanCeyer/adk-go/tools"
)

func init() {
	provider, err := llmproviders.NewGeminiLLMProvider()
	if err != nil {
		examples.RegisterAgent("helloworld", nil, err)
		return
	}

	rollDieTool := tools.NewRollDieTool()
	agentTools := []tools.Tool{rollDieTool}

	systemInstructionText := "You are a friendly assistant named HelloWorldAgent. Your special ability is to roll dice. When the conversation starts with a simple greeting, introduce yourself and ask if the user wants to roll a die. For example: 'Hi there! I'm the HelloWorldAgent. I can roll dice for you. Would you like to roll one?'. For other requests, use the rollDie tool and report the result clearly, like 'You rolled a 5 on a 6-sided die.'."
	systemInstruction := &modelstypes.Message{
		Parts: []modelstypes.Part{{Text: &systemInstructionText}},
	}

	agent := agents.NewBaseLlmAgent(
		"helloworld",
		"A simple agent that can roll a die using a tool.",
		"gemini-2.5-flash",
		systemInstruction,
		provider,
		agentTools,
	)
	examples.RegisterAgent("helloworld", agent, nil)
}
