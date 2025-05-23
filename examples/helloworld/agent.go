package helloworld

import (
	"fmt"
	"log"

	"github.com/KennethanCeyer/adk-go/adk"
	"github.com/KennethanCeyer/adk-go/llmproviders"
	"github.com/KennethanCeyer/adk-go/tools"
)

var Agent adk.Agent

func init() {
	geminiProvider, err := llmproviders.NewGeminiLLMProvider()
	if err != nil {
		panic(fmt.Sprintf("helloworld.init: Failed to create GeminiLLMProvider: %v. Check GEMINI_API_KEY.", err))
	}

	rollDieTool := tools.NewRollDieTool()
	agentTools := []adk.Tool{rollDieTool}

	Agent = adk.NewBaseAgent(
		"HelloWorldAgent",
		"You are a friendly assistant. You can roll dice. When you use the roll_die tool, tell the user what was rolled and the die type (e.g., 'You rolled a 5 on a 6-sided die.').",
		"gemini-1.5-flash-latest",
		geminiProvider,
		agentTools,
	)
	log.Println("HelloWorldAgent initialized in examples/helloworld/agent.go.")
}
