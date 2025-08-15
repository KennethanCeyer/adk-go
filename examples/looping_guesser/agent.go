package looping_guesser

import (
	"log"

	"github.com/KennethanCeyer/adk-go/agents"
	"github.com/KennethanCeyer/adk-go/examples"
	"github.com/KennethanCeyer/adk-go/llmproviders"
	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
	"github.com/KennethanCeyer/adk-go/tools"
	"github.com/KennethanCeyer/adk-go/tools/example"
)

func init() {
	geminiProvider, err := llmproviders.NewGeminiLLMProvider()
	if err != nil {
		log.Printf("Warning: Could not initialize 'looping_guesser' agent: %v. Ensure GEMINI_API_KEY is set.", err)
		return
	}

	guesserInstructionText := `You are a number guessing bot that plays a number guessing game with the user. Your goal is to guess a secret number between 1 and 100 within 10 attempts.

**Game Flow:**
1.  **Initiation:** When the user starts the conversation (e.g., with "Hi", "guess the number", or any other message), you MUST begin the game by making your first guess.
2.  **First Guess:** Your first guess MUST always be 50. Use the 'check_guess' tool with the argument 'guess: 50'.
3.  **Subsequent Guesses:** For every subsequent turn, you will receive a tool response with a 'status' of 'too_low', 'too_high', or 'correct'.
    - Based on this status and the history of your guesses, you must make a new, logical guess to narrow down the range of possible numbers.
    - Announce the result of the previous guess (e.g., "50 was too high.") and then immediately make your next guess by calling the 'check_guess' tool again.
4.  **Winning:** If the tool response status is 'correct', you have won. Your final response MUST be "I guessed the number! I win!".
5.  **Losing:** You have a maximum of 10 tool calls to 'check_guess'. If you have not guessed the number after 10 attempts, your final response MUST be "I have failed to guess the number in 10 tries. I admit defeat.".

**Important Rules:**
- Always use the 'check_guess' tool to make a guess.
- Do not make up results. Your actions are driven by the tool responses.`
	guesserInstruction := &modelstypes.Message{Parts: []modelstypes.Part{{Text: &guesserInstructionText}}}

	guesserAgent := agents.NewBaseLlmAgent(
		"looping_guesser",
		"An agent that plays a number guessing game.",
		"gemini-2.5-flash",
		guesserInstruction,
		geminiProvider,
		[]tools.Tool{example.NewNumberGuesserTool()},
	)

	examples.RegisterAgent("looping_guesser", guesserAgent)
}
