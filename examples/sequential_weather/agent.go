package sequential_weather

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
		log.Printf("Warning: Could not initialize 'sequential_weather' agent: %v. Ensure GEMINI_API_KEY is set.", err)
		return
	}

	// 1. Greeting Agent
	greetingText := "You are a friendly greeter. Your job is to greet the user and then clearly repeat their weather-related query to pass to the next step. For example, if the user says 'weather in london', you should respond with something like 'Hello! You want to know the weather for: london'."
	greetingInstruction := &modelstypes.Message{Parts: []modelstypes.Part{{Text: &greetingText}}}
	greetingAgent := agents.NewBaseLlmAgent(
		"GreeterAgent",
		"A sub-agent that handles greetings.",
		"gemini-1.5-pro-latest",
		greetingInstruction,
		geminiProvider,
		nil,
	)

	// 2. Weather Agent
	weatherText := "You are a weather bot. You will receive input that contains a weather query, possibly prefixed by a greeting. Your task is to extract the city name from the query and use the `getWeather` tool to find the weather."
	weatherInstruction := &modelstypes.Message{Parts: []modelstypes.Part{{Text: &weatherText}}}
	weatherAgent := agents.NewBaseLlmAgent(
		"WeatherAgent",
		"A sub-agent that provides weather information.",
		"gemini-1.5-pro-latest",
		weatherInstruction,
		geminiProvider,
		[]tools.Tool{example.NewWeatherTool()},
	)

	// 3. Sequential Workflow Agent
	sequentialAgent := agents.NewSequentialAgent(
		"sequential_weather",
		"A workflow that first greets the user, then provides the weather.",
		[]interfaces.LlmAgent{greetingAgent, weatherAgent},
	)

	examples.RegisterAgent("sequential_weather", sequentialAgent)
}
