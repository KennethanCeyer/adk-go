package sequential_weather

import (
	"log"

	"github.com/KennethanCeyer/adk-go/agents"
	"github.com/KennethanCeyer/adk-go/examples"
	"github.com/KennethanCeyer/adk-go/llmproviders"
	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
	"github.com/KennethanCeyer/adk-go/tools"
	"github.com/KennethanCeyer/adk-go/tools/example"
)

func init() {
	geminiProvider, err := llmproviders.NewGeminiLLMProvider()
	if err != nil {
		log.Printf("Warning: Could not initialize 'sequential_weather' agent: %v. Ensure GEMINI_API_KEY is set.", err)
		return
	}

	// This example is now a single, more capable agent instead of a sequential workflow.
	// This provides a more natural and efficient user experience for this specific use case,
	// addressing the issue of running unnecessary sub-agents for simple greetings.
	instructionText := "You are a friendly and helpful weather assistant. If the user asks for the weather, use the `getWeather` tool to provide the information. If the user is just making small talk, respond conversationally. Be concise and friendly."
	systemInstruction := &modelstypes.Message{Parts: []modelstypes.Part{{Text: &instructionText}}}

	weatherAgent := agents.NewBaseLlmAgent(
		"sequential_weather", // Keep the name for consistency with the command
		"A friendly assistant that can provide weather information using a tool.",
		"gemini-2.5-flash",
		systemInstruction,
		geminiProvider,
		[]tools.Tool{example.NewWeatherTool()},
	)

	examples.RegisterAgent("sequential_weather", weatherAgent)
}
