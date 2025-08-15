package parallel_trip_planner

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
		log.Printf("Warning: Could not initialize 'parallel_trip_planner' agent: %v. Ensure GEMINI_API_KEY is set.", err)
		return
	}

	// 1. Flight Agent
	flightInstructionText := "You are a flight booking assistant. Your goal is to find flights using the `find_flights` tool. To do this, you need a destination city and a travel date. If the user provides both, call the tool. If any information is missing, ask the user for it. Do not make up information. Be concise."
	flightInstruction := &modelstypes.Message{Parts: []modelstypes.Part{{Text: &flightInstructionText}}}
	flightAgent := agents.NewBaseLlmAgent(
		"FlightAgent",
		"A sub-agent that finds flights.",
		"gemini-1.5-pro-latest",
		flightInstruction,
		geminiProvider,
		[]tools.Tool{example.NewFlightTool()},
	)

	// 2. Hotel Agent
	hotelInstructionText := "You are a hotel booking assistant. Your goal is to find hotels using the `find_hotels` tool. To do this, you need a destination city and a check-in date. If the user provides both, call the tool. If any information is missing, ask the user for it. Do not make up information. Be concise."
	hotelInstruction := &modelstypes.Message{Parts: []modelstypes.Part{{Text: &hotelInstructionText}}}
	hotelAgent := agents.NewBaseLlmAgent(
		"HotelAgent",
		"A sub-agent that finds hotels.",
		"gemini-1.5-pro-latest",
		hotelInstruction,
		geminiProvider,
		[]tools.Tool{example.NewHotelTool()},
	)

	// 3. Parallel Workflow Agent (Trip Planner)
	synthesisInstructionText := "You are a helpful travel planner. You will be given information about flights and hotels. Combine this information into a single, easy-to-read travel plan summary for the user. Be friendly and confirm the details you've found."
	synthesisInstruction := &modelstypes.Message{Parts: []modelstypes.Part{{Text: &synthesisInstructionText}}}

	tripPlannerAgent := agents.NewParallelAgent(
		"parallel_trip_planner",
		"A workflow that finds flights and hotels in parallel and synthesizes a travel plan.",
		"gemini-1.5-pro-latest",
		synthesisInstruction,
		geminiProvider,
		[]interfaces.LlmAgent{flightAgent, hotelAgent},
	)

	examples.RegisterAgent("parallel_trip_planner", tripPlannerAgent)
}
