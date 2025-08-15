package example

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
)

// WeatherTool is a simple tool that returns mock weather data.
type WeatherTool struct{}

// NewWeatherTool creates a WeatherTool.
func NewWeatherTool() *WeatherTool {
	return &WeatherTool{}
}

func (t *WeatherTool) Name() string {
	return "get_weather"
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
				Description: "The city to get the weather for, e.g., 'San Francisco'.",
			},
		},
		Required: []string{"city"},
	}
}

func (t *WeatherTool) Execute(ctx context.Context, args any) (any, error) {
	argsMap, ok := args.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("get_weather: invalid arguments format, expected map[string]any, got %T", args)
	}

	cityVal, ok := argsMap["city"]
	if !ok {
		return nil, fmt.Errorf("get_weather: missing 'city' argument")
	}

	cityStr, ok := cityVal.(string)
	if !ok {
		return nil, fmt.Errorf("get_weather: 'city' argument must be a string, got %T", cityVal)
	}

	// Mock weather data
	mockWeatherDB := map[string]string{
		"new york": "Sunny with a temperature of 25°C.",
		"london":   "Cloudy with a temperature of 15°C.",
		"tokyo":    "Light rain and a temperature of 18°C.",
		"seoul":    "Clear skies with a temperature of 22°C.",
	}

	if report, found := mockWeatherDB[strings.ToLower(cityStr)]; found {
		return map[string]any{"report": report}, nil
	}

	return map[string]any{"report": fmt.Sprintf("Sorry, I don't have weather information for '%s'.", cityStr)}, nil
}
