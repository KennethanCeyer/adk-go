package example

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
)

// HotelTool is a simple tool that returns mock hotel data.
type HotelTool struct{}

// NewHotelTool creates a HotelTool.
func NewHotelTool() *HotelTool {
	return &HotelTool{}
}

func (t *HotelTool) Name() string {
	return "find_hotels"
}

func (t *HotelTool) Description() string {
	return "Finds hotel options for a given destination and date."
}

func (t *HotelTool) Parameters() any {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"destination": {
				Type:        genai.TypeString,
				Description: "The destination city, e.g., 'Tokyo'.",
			},
			"date": {
				Type:        genai.TypeString,
				Description: "The check-in date in YYYY-MM-DD format.",
			},
		},
		Required: []string{"destination", "date"},
	}
}

func (t *HotelTool) Execute(ctx context.Context, args any) (any, error) {
	argsMap, ok := args.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("find_hotels: invalid arguments format, expected map[string]any, got %T", args)
	}

	destVal, ok := argsMap["destination"]
	if !ok {
		return nil, fmt.Errorf("find_hotels: missing 'destination' argument")
	}
	destStr, ok := destVal.(string)
	if !ok {
		return nil, fmt.Errorf("find_hotels: 'destination' argument must be a string, got %T", destVal)
	}

	dateVal, ok := argsMap["date"]
	if !ok {
		return nil, fmt.Errorf("find_hotels: missing 'date' argument")
	}
	dateStr, ok := dateVal.(string)
	if !ok {
		return nil, fmt.Errorf("find_hotels: 'date' argument must be a string, got %T", dateVal)
	}

	// Mock hotel data
	report := fmt.Sprintf("Found a room at the 'Cosmic Inn' in %s starting %s for $250/night.", destStr, dateStr)
	return map[string]any{"report": report}, nil
}
