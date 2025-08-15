package sequential_weather

import (
	"log"

	"github.com/KennethanCeyer/adk-go/agents"
	"github.com/KennethanCeyer/adk-go/llmproviders"
	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
	"github.com/KennethanCeyer/adk-go/tools"
)

var SequentialWeatherAgent agents.LlmAgent

func init() {
	geminiProvider, err := llmproviders.NewGeminiLLMProvider()
	if err != nil {
		log.Fatalf("Failed to create GeminiLLMProvider: %v. Ensure GEMINI_API_KEY is set.", err)
	}

	// --- Define Specialist Sub-Agents ---

	// 1. Greeting Agent
	greetingInstruction := &modelstypes.Message{
		Parts: []modelstypes.Part{{Text: new(string)}},
	}
	*greetingInstruction.Parts[0].Text = "You are a friendly greeter. Just say hello and ask how you can help with the weather."
	greetingAgent := agents.NewBaseLlmAgent(
		"GreeterAgent",
		"A sub-agent that handles greetings.",
		"gemini-2.5-pro",
		greetingInstruction,
		geminiProvider,
		nil, // No tools for the greeter
	)

	// 2. Weather Agent
	weatherInstruction := &modelstypes.Message{
		Parts: []modelstypes.Part{{Text: new(string)}},
	}
	// Using RollDie as a stand-in for a real weather tool for this example.
	*weatherInstruction.Parts[0].Text = "You are a weather bot. The user has already been greeted. Use the roll_die tool to simulate getting weather for 'New York'."
	weatherAgent := agents.NewBaseLlmAgent(
		"WeatherAgent",
		"A sub-agent that provides weather information.",
		"gemini-2.5-pro",
		weatherInstruction,
		geminiProvider,
		[]tools.Tool{tools.NewRollDieTool()},
	)

	// --- Define the Sequential Workflow Agent ---
	SequentialWeatherAgent = agents.NewSequentialAgent(
		"SequentialWeatherWorkflow",
		"A workflow that first greets the user, then provides the weather.",
		[]agents.LlmAgent{greetingAgent, weatherAgent},
	)

	log.Println("SequentialWeatherAgent initialized in examples/sequential_weather/agent.go.")
}
