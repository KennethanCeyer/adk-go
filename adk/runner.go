package adk

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/KennethanCeyer/adk-go/agents"
	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
)

type SimpleCLIRunner struct {
	AgentToRun agents.LlmAgent
}

func NewSimpleCLIRunner(agent agents.LlmAgent) (*SimpleCLIRunner, error) {
	if agent == nil {
		return nil, fmt.Errorf("agent cannot be nil")
	}
	return &SimpleCLIRunner{AgentToRun: agent}, nil
}

func (r *SimpleCLIRunner) Start(ctx context.Context) {
	fmt.Printf("--- Starting Agent: %s ---\n", r.AgentToRun.GetName())
	if r.AgentToRun.GetDescription() != "" {
		fmt.Printf("Description: %s\n", r.AgentToRun.GetDescription())
	}
	fmt.Printf("Model: %s\n", r.AgentToRun.GetModelIdentifier())
	if len(r.AgentToRun.GetTools()) > 0 {
		fmt.Println("Available Tools:")
		for _, tool := range r.AgentToRun.GetTools() {
			fmt.Printf("  - %s: %s\n", tool.Name(), tool.Description())
		}
	}
	fmt.Println("------------------------------------")
	fmt.Println("Type 'exit' or 'quit' to stop.")

	var history []modelstypes.Message
	scanner := bufio.NewScanner(os.Stdin)

	for {
		select {
		case <-ctx.Done():
			log.Println("Runner loop: Context cancelled. Exiting loop.")
			return
		default:
		}

		fmt.Print("[user]: ")
		if !scanner.Scan() {
			if ctx.Err() != nil {
				log.Println("Runner loop: Context cancelled during input scan. Exiting.")
			} else if err := scanner.Err(); err != nil {
				log.Printf("Input error: %v. Exiting.", err)
			} else {
				log.Println("EOF received on input. Exiting.")
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

		userMessage := modelstypes.Message{Role: "user", Parts: []modelstypes.Part{{Text: &userInputText}}}

		currentHistoryForCall := make([]modelstypes.Message, len(history))
		copy(currentHistoryForCall, history)

		agentResponse, err := r.AgentToRun.Process(ctx, currentHistoryForCall, userMessage)
		if err != nil {
			if ctx.Err() != nil {
				log.Printf("Agent Process call failed due to context cancellation: %v", err)
				return
			}
			fmt.Printf("[%s-error]: I encountered an issue: %v\n", r.AgentToRun.Name(), err)
			history = append(history, userMessage)
			continue
		}

		history = append(history, userMessage)
		if agentResponse != nil {
			history = append(history, *agentResponse)
		}

		if agentResponse != nil && len(agentResponse.Parts) > 0 {
			var responseTexts []string
			for _, part := range agentResponse.Parts {
				if part.Text != nil {
					responseTexts = append(responseTexts, *part.Text)
				}
				if part.FunctionCall != nil {
					log.Printf("Runner: Agent response unexpectedly contained FunctionCall: Name=%s.", part.FunctionCall.Name)
				}
			}
			fmt.Printf("[%s]: %s\n", r.AgentToRun.GetName(), strings.Join(responseTexts, "\n"))
		} else {
			fmt.Printf("[%s]: (Agent returned no displayable content)\n", r.AgentToRun.GetName())
		}

		const maxHistoryTurns = 10
		if len(history) > maxHistoryTurns*2 {
			history = history[len(history)-(maxHistoryTurns*2):]
		}
	}
}
