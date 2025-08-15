package example

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
)

type WeatherTool struct{}

// NewWeatherTool creates a WeatherTool.
func NewWeatherTool() *WeatherTool {
	return &WeatherTool{}
}

func (t *WeatherTool) Name() string {
	return "getWeather"
}

func (t *WeatherTool) Description() string {
	return "Gets the current weather for a specified city."
}

func (t *WeatherTool) Parameters() any {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"city": {
				Type:        genai.TypeString,
				Description: "The city name, e.g., 'London' or 'Tokyo'.",
			},
		},
		Required: []string{"city"},
	}
}

func (t *WeatherTool) Execute(ctx context.Context, args any) (any, error) {
	argsMap, ok := args.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("getWeather: invalid arguments format, expected map[string]any, got %T", args)
	}

	cityVal, ok := argsMap["city"]
	if !ok {
		return nil, fmt.Errorf("getWeather: missing 'city' argument")
	}
	cityStr, ok := cityVal.(string)
	if !ok {
		return nil, fmt.Errorf("getWeather: 'city' argument must be a string, got %T", cityVal)
	}

	// Mock weather data
	report := fmt.Sprintf("The weather in %s is 72°F and sunny.", cityStr)
	return map[string]any{"report": report}, nil
}
