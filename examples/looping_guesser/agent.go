package looping_guesser

import (
	"strings"

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
		examples.RegisterAgent("looping_guesser", nil, err)
		return
	}

	guesserInstructionText := `You are a number guessing bot playing a game. Your goal is to guess a secret number between 1 and 100.
You will be given the history of previous guesses and their results ('too_low' or 'too_high').
Based on the history, make the most logical next guess using a binary search strategy.
If the history is empty, your first guess must be 50.
Announce your new guess and then call the 'check_guess' tool with your guess.
If the tool response indicates you are 'correct', your final response MUST be "I guessed the number! I win!". Do not call any more tools.`
	guesserInstruction := &modelstypes.Message{Parts: []modelstypes.Part{{Text: &guesserInstructionText}}}

	guesserSubAgent := agents.NewBaseLlmAgent(
		"guesser_sub_agent",
		"An agent that makes a single guess in a number guessing game.",
		"gemini-2.5-flash",
		guesserInstruction,
		geminiProvider,
		[]tools.Tool{example.NewNumberGuesserTool()},
	)

	loopingGuesserAgent := agents.NewLoopAgent(
		"looping_guesser",
		"An agent that plays a number guessing game automatically by looping.",
		[]interfaces.LlmAgent{guesserSubAgent},
		10,
		func(latestResponse *modelstypes.Message) bool {
			if latestResponse != nil {
				for _, part := range latestResponse.Parts {
					if part.Text != nil && strings.Contains(*part.Text, "I win!") {
						return true
					}
				}
			}
			return false
		},
	)

	examples.RegisterAgent("looping_guesser", loopingGuesserAgent, nil)
}
