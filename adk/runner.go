package adk

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/KennethanCeyer/adk-go/agents/interfaces"
	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
	"github.com/KennethanCeyer/adk-go/sessions"
)

type SimpleCLIRunner struct {
	AgentToRun interfaces.LlmAgent
	Session    *sessions.Session
}

func NewSimpleCLIRunner(agent interfaces.LlmAgent, sess *sessions.Session) (*SimpleCLIRunner, error) {
	if agent == nil {
		return nil, fmt.Errorf("agent cannot be nil")
	}
	if sess == nil {
		return nil, fmt.Errorf("session cannot be nil")
	}
	return &SimpleCLIRunner{AgentToRun: agent, Session: sess}, nil
}

func (r *SimpleCLIRunner) Start(ctx context.Context) {
	r.printAgentInfo()

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
		normalizedInput := strings.ToLower(userInputText)
		normalizedInput = strings.TrimRight(normalizedInput, "()")

		if normalizedInput == "exit" || normalizedInput == "quit" {
			fmt.Println("\nExiting agent.")
			return
		}
		if userInputText == "" {
			continue
		}

		userMessage := modelstypes.Message{Role: "user", Parts: []modelstypes.Part{{Text: &userInputText}}}

		agentResponse, err := r.AgentToRun.Process(ctx, r.Session.History, userMessage)
		if err != nil {
			if ctx.Err() != nil {
				log.Printf("Agent Process call failed due to context cancellation: %v", err)
				return
			}
			fmt.Printf("[%s-error]: I encountered an issue: %v\n", r.AgentToRun.GetName(), err)
			r.Session.History = append(r.Session.History, userMessage)
			sessions.Save(r.Session)
			continue
		}

		r.Session.History = append(r.Session.History, userMessage)
		if agentResponse != nil {
			r.Session.History = append(r.Session.History, *agentResponse)
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
		if len(r.Session.History) > maxHistoryTurns*2 {
			r.Session.History = r.Session.History[len(r.Session.History)-(maxHistoryTurns*2):]
		}

		sessions.Save(r.Session)
	}
}

func (r *SimpleCLIRunner) printAgentInfo() {
	fmt.Printf("--- Starting Agent: %s ---\n", r.AgentToRun.GetName())
	if desc := r.AgentToRun.GetDescription(); desc != "" {
		fmt.Printf("Description: %s\n", desc)
	}
	if model := r.AgentToRun.GetModelIdentifier(); strings.HasPrefix(model, "workflow-") {
		fmt.Println("Model: (This is a workflow agent)")
	} else {
		fmt.Printf("Model: %s\n", model)
	}
	if tools := r.AgentToRun.GetTools(); len(tools) > 0 {
		fmt.Println("Available Tools:")
		for _, tool := range tools {
			fmt.Printf("  - %s: %s\n", tool.Name(), tool.Description())
		}
	}
	fmt.Println("------------------------------------")
	fmt.Println("Type 'exit' or 'quit' to stop.")
}
