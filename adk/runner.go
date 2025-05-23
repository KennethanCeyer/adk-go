package adk

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	adktypes "github.com/KennethanCeyer/adk-go/adk/types"
)

type SimpleCLIRunner struct {
	AgentToRun Agent
}

func NewSimpleCLIRunner(agent Agent) (*SimpleCLIRunner, error) {
	if agent == nil {
		return nil, fmt.Errorf("agent cannot be nil")
	}
	return &SimpleCLIRunner{AgentToRun: agent}, nil
}

func (r *SimpleCLIRunner) Start(ctx context.Context) {
	fmt.Printf("--- Starting Agent: %s ---\n", r.AgentToRun.Name())
	fmt.Println("Type 'exit' or 'quit' to stop.")

	var history []adktypes.Message
	scanner := bufio.NewScanner(os.Stdin)

	for {
		select {
		case <-ctx.Done():
			log.Println("Runner context cancelled, shutting down.")
			return
		default:
		}

		fmt.Print("[user]: ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				log.Printf("Error reading input: %v", err)
			}
			return
		}
		userInputText := strings.TrimSpace(scanner.Text())

		if strings.ToLower(userInputText) == "exit" || strings.ToLower(userInputText) == "quit" {
			fmt.Println("Exiting agent.")
			return
		}
		if userInputText == "" {
			continue
		}

		userMessage := adktypes.Message{Role: "user", Parts: []adktypes.Part{{Text: &userInputText}}}

		currentHistoryForCall := make([]adktypes.Message, len(history))
		copy(currentHistoryForCall, history)

		agentResponse, err := r.AgentToRun.Process(ctx, currentHistoryForCall, userMessage)
		if err != nil {
			log.Printf("Error from agent %s: %v", r.AgentToRun.Name(), err)
			fmt.Printf("[%s-error]: Sorry, I encountered an error: %v\n", r.AgentToRun.Name(), err)
			continue
		}

		history = append(history, userMessage)
		history = append(history, agentResponse)

		var responseTexts []string
		for _, part := range agentResponse.Parts {
			if part.Text != nil {
				responseTexts = append(responseTexts, *part.Text)
			}
			if part.FunctionCall != nil {
				responseTexts = append(responseTexts, fmt.Sprintf("[INFO: Agent should have handled FunctionCall: %s. This indicates an issue in BaseAgent.Process if seen by user.]", part.FunctionCall.Name))
			}
		}
		fmt.Printf("[%s]: %s\n", r.AgentToRun.Name(), strings.Join(responseTexts, "\n"))

		const maxHistoryItems = 20
		if len(history) > maxHistoryItems {
			history = history[len(history)-maxHistoryItems:]
		}
	}
}
