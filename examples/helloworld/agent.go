package helloworld

import (
	"log"

	"github.com/KennethanCeyer/adk-go/agents"
	"github.com/KennethanCeyer/adk-go/examples"
	"github.com/KennethanCeyer/adk-go/llmproviders"
	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
	"github.com/KennethanCeyer/adk-go/tools"
)

func init() {
	geminiProvider, err := llmproviders.NewGeminiLLMProvider()
	if err != nil {
		log.Fatalf("Failed to create GeminiLLMProvider: %v. Ensure GEMINI_API_KEY is set.", err)
	}

	rollDieTool := tools.NewRollDieTool()
	agentTools := []tools.Tool{rollDieTool}

	systemInstructionText := "You are a friendly assistant. You can roll dice. When you use the roll_die tool, tell the user what was rolled and the die type (e.g., 'You rolled a 5 on a 6-sided die.')."
	systemInstruction := &modelstypes.Message{
		Parts: []modelstypes.Part{{Text: &systemInstructionText}},
	}

	agent := agents.NewBaseLlmAgent(
		"HelloWorldAgent",
		"A simple agent that can roll a die using a tool.",
		"gemini-1.5-pro-latest",
		systemInstruction,
		geminiProvider,
		agentTools,
	)
	examples.RegisterAgent("helloworld", agent)
}
