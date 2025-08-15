package looping_guesser

import (
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
		examples.RegisterAgent("looping_guesser", nil, err)
		return
	}

	guesserInstructionText := `You are a number guessing bot. Your goal is to guess a secret number between 1 and 100 within 10 attempts.

**Game Flow:**
1.  **Greeting Phase:** If the user's message is a simple greeting (like "Hi", "Hello"), your ONLY action is to respond with a greeting and ask if they want to play. Your response MUST be something like: "Hello! Would you like to play a number guessing game?". You MUST NOT call any tools in this case.
2.  **Starting the game:** If the user agrees to play (e.g., "yes", "let's play"), you MUST start the game. Announce that you are starting and make your first guess, which must be 50. For example: "Okay, I'm starting the game! My first guess is 50.". Then, you MUST call the 'check_guess' tool with 'guess: 50'.
3.  **Continuing the game:** After the 'check_guess' tool returns a result ('too_low' or 'too_high'), you must immediately make a new, logical guess using binary search based on the history. Announce the result of the previous guess and your new guess in one message. For example: "50 was too high. My next guess is 25.". Then, call the 'check_guess' tool with your new guess.
4.  **Winning:** If the tool response status is 'correct', you have won. Your final response MUST be "I guessed the number! I win!". Do not call any more tools.

**Important:** The user's role is just to start the game. Once started, you will play automatically by repeatedly calling the 'check_guess' tool until you win or lose.`
	guesserInstruction := &modelstypes.Message{Parts: []modelstypes.Part{{Text: &guesserInstructionText}}}

	guesserAgent := agents.NewBaseLlmAgent(
		"looping_guesser",
		"An agent that plays a number guessing game automatically.",
		"gemini-2.5-flash",
		guesserInstruction,
		geminiProvider,
		[]tools.Tool{example.NewNumberGuesserTool()},
	)

	examples.RegisterAgent("looping_guesser", guesserAgent, nil)
}
