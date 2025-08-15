package tools

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	// "github.com/KennethanCeyer/adk-go/adk" // adk.Tool is implemented
)

type RollDieTool struct{}

func NewRollDieTool() *RollDieTool {
	return &RollDieTool{}
}

func (t *RollDieTool) Name() string { return "rollDie" }

func (t *RollDieTool) Description() string {
	return "Rolls a die with a specified number of sides and returns the result."
}

func (t *RollDieTool) Parameters() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"sides": map[string]any{
				"type":        "integer",
				"description": "The number of sides on the die (e.g., 6, 20). Must be positive.",
			},
		},
		"required": []string{"sides"},
	}
}

func (t *RollDieTool) Execute(ctx context.Context, args any) (any, error) {
	argsMap, ok := args.(map[string]any)
	if !ok { return nil, fmt.Errorf("rolldie: invalid args format, expected map[string]any, got %T", args) }
	sidesVal, ok := argsMap["sides"]
	if !ok { return nil, fmt.Errorf("rolldie: missing 'sides' argument") }

	var sides int64
	switch v := sidesVal.(type) {
	case float64:
		if v <= 0 || v != float64(int64(v)) { return nil, fmt.Errorf("rolldie: 'sides' (float64) must be positive integer, got %f", v) }
		sides = int64(v)
	case int:
		if v <= 0 { return nil, fmt.Errorf("rolldie: 'sides' (int) must be positive, got %d", v) }
		sides = int64(v)
	// Consider json.Number if LLM strictly adheres to JSON spec for numbers from schema
	default:
		return nil, fmt.Errorf("rolldie: 'sides' argument must be numeric, got type %T", sidesVal)
	}
	if sides <= 0 { return nil, fmt.Errorf("rolldie: 'sides' must be positive, got %d", sides) }

	nBig, err := rand.Int(rand.Reader, big.NewInt(sides))
	if err != nil { return nil, fmt.Errorf("rolldie: rand.Int failed: %w", err) }
	result := nBig.Int64() + 1
	log.Printf("RollDieTool: Executed. Rolled %d (d%d)", result, sides)
	return map[string]any{"result": result, "sidesRolled": sides}, nil
}
