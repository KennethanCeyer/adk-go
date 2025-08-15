package sequentialweather

import (
	"log"

	"github.com/KennethanCeyer/adk-go/agents"
	"github.com/KennethanCeyer/adk-go/agents/interfaces"
	"github.com/KennethanCeyer/adk-go/examples"
	"github.com/KennethanCeyer/adk-go/llmproviders"
	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
	"github.com/KennethanCeyer/adk-go/tools"
	"github.com/KennethanCeyer/adk-go/tools/example"
)

func init() {
	geminiProvider, err := llmproviders.NewGeminiLLMProvider()
	if err != nil {
		log.Fatalf("Failed to create GeminiLLMProvider: %v. Ensure GEMINI_API_KEY is set.", err)
	}

	// 1. Greeting Agent
	greetingInstruction := &modelstypes.Message{
		Parts: []modelstypes.Part{{Text: new(string)}},
	}
	*greetingInstruction.Parts[0].Text = "You are a friendly greeter. Your job is to greet the user and then clearly repeat their weather-related query to pass to the next step. For example, if the user says 'weather in london', you should respond with something like 'Hello! You want to know the weather for: london'."
	greetingAgent := agents.NewBaseLlmAgent(
		"GreeterAgent",
		"A sub-agent that handles greetings.",
		"gemini-1.5-pro-latest",
		greetingInstruction,
		geminiProvider,
		nil, // No tools for the greeter
	)

	weatherInstruction := &modelstypes.Message{
		Parts: []modelstypes.Part{{Text: new(string)}},
	}
	*weatherInstruction.Parts[0].Text = "You are a weather bot. You will receive input that contains a weather query, possibly prefixed by a greeting. Your task is to extract the city name from the query and use the `get_weather` tool to find the weather."
	weatherAgent := agents.NewBaseLlmAgent(
		"WeatherAgent",
		"A sub-agent that provides weather information.",
		"gemini-1.5-pro-latest",
		weatherInstruction,
		geminiProvider,
		[]tools.Tool{example.NewWeatherTool()},
	)

	sequentialAgent := agents.NewSequentialAgent(
		"SequentialWeatherWorkflow",
		"A workflow that first greets the user, then provides the weather.",
		[]interfaces.LlmAgent{greetingAgent, weatherAgent},
	)

	examples.RegisterAgent("sequentialWeather", sequentialAgent)
}
