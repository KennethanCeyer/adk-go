package sequential_weather

import (
	"github.com/KennethanCeyer/adk-go/agents"
	"github.com/KennethanCeyer/adk-go/examples"
	"github.com/KennethanCeyer/adk-go/llmproviders"
	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
	"github.com/KennethanCeyer/adk-go/tools"
	"github.com/KennethanCeyer/adk-go/tools/example"
)

func init() {
	provider, err := llmproviders.NewGeminiLLMProvider()
	if err != nil {
		examples.RegisterAgent("sequential_weather", nil, err)
		return
	}

	// This example uses a single, capable agent instead of a sequential workflow for efficiency.
	instructionText := "You are a friendly and helpful weather assistant. When the conversation starts with a simple greeting, introduce yourself and ask which city's weather they'd like to know. For example: 'Hello! I can get the latest weather report for you. Which city are you interested in?'. If the user asks for the weather directly, use the `getWeather` tool to provide the information. If the user is just making small talk, respond conversationally. Be concise and friendly."
	systemInstruction := &modelstypes.Message{Parts: []modelstypes.Part{{Text: &instructionText}}}

	weatherAgent := agents.NewBaseLlmAgent(
		"sequential_weather", // Keep the name for consistency with the command
		"A friendly assistant that can provide weather information using a tool.",
		"gemini-2.5-flash",
		systemInstruction,
		provider,
		[]tools.Tool{example.NewWeatherTool()},
	)

	examples.RegisterAgent("sequential_weather", weatherAgent, nil)
}
