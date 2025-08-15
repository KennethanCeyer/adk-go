package example

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/generative-ai-go/genai"
)

type NumberGuesserTool struct {
	secret int
}

func NewNumberGuesserTool() *NumberGuesserTool {
	// Seed the random number generator
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return &NumberGuesserTool{
		secret: r.Intn(100) + 1,
	}
}

func (t *NumberGuesserTool) Name() string {
	return "check_guess"
}

func (t *NumberGuesserTool) Description() string {
	return "Checks a guessed number against the secret number. The secret is between 1 and 100."
}

func (t *NumberGuesserTool) Parameters() any {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"guess": {
				Type:        genai.TypeInteger,
				Description: "The number to guess.",
			},
		},
		Required: []string{"guess"},
	}
}

func (t *NumberGuesserTool) Execute(ctx context.Context, args any) (any, error) {
	argsMap, ok := args.(map[string]any)
	if !ok { return nil, fmt.Errorf("check_guess: invalid arguments format, expected map[string]any, got %T", args) }
	guessVal, ok := argsMap["guess"]
	if !ok { return nil, fmt.Errorf("check_guess: missing 'guess' argument") }
	// The JSON unmarshaling from the LLM often results in float64 for numbers.
	guessFloat, ok := guessVal.(float64)
	if !ok { return nil, fmt.Errorf("check_guess: 'guess' argument must be a number, got %T", guessVal) }
	guess := int(guessFloat)

	var status string
	if guess < t.secret { status = "too_low"
	} else if guess > t.secret { status = "too_high"
	} else { status = "correct" }
	return map[string]any{"status": status}, nil
}
