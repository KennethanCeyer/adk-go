package example

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
)

type FlightTool struct{}

func NewFlightTool() *FlightTool {
	return &FlightTool{}
}

func (t *FlightTool) Name() string {
	return "find_flights"
}

func (t *FlightTool) Description() string {
	return "Finds flight options for a given destination and date."
}

func (t *FlightTool) Parameters() any {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"destination": {
				Type:        genai.TypeString,
				Description: "The destination city, e.g., 'Tokyo'.",
			},
			"date": {
				Type:        genai.TypeString,
				Description: "The date of travel in YYYY-MM-DD format.",
			},
		},
		Required: []string{"destination", "date"},
	}
}

func (t *FlightTool) Execute(ctx context.Context, args any) (any, error) {
	argsMap, ok := args.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("find_flights: invalid arguments format, expected map[string]any, got %T", args)
	}

	destVal, ok := argsMap["destination"]
	if !ok {
		return nil, fmt.Errorf("find_flights: missing 'destination' argument")
	}
	destStr, ok := destVal.(string)
	if !ok {
		return nil, fmt.Errorf("find_flights: 'destination' argument must be a string, got %T", destVal)
	}

	dateVal, ok := argsMap["date"]
	if !ok {
		return nil, fmt.Errorf("find_flights: missing 'date' argument")
	}
	dateStr, ok := dateVal.(string)
	if !ok {
		return nil, fmt.Errorf("find_flights: 'date' argument must be a string, got %T", dateVal)
	}

	// Mock flight data
	report := fmt.Sprintf("Found a round-trip flight to %s on %s for $1200 on 'Galaxy Airlines'.", destStr, dateStr)
	return map[string]any{"report": report}, nil
}
