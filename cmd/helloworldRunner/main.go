package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/KennethanCeyer/adk-go/examples/helloworld" // Corrected path
	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	// Ensure cancel is called when main exits, to clean up any resources
	// associated with the context, although os.Exit will terminate directly.
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	// Notify sigChan for SIGINT (Ctrl+C) and SIGTERM.
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Goroutine to handle graceful shutdown on receiving a signal.
	go func() {
		sig := <-sigChan
		log.Printf("Shutdown signal received: %v. Cancelling context...", sig)
		cancel() // Signal cancellation to all parts of the application using this context.

		// Allow a short period for cleanup before forcing exit.
		// This helps if any goroutines are performing final actions.
		time.Sleep(500 * time.Millisecond)

		log.Println("Forcing exit due to signal.")
		os.Exit(0) // Exit the program. Status 0 for normal shutdown.
	}()

	if helloworld.ConcreteLlmAgent == nil {
		log.Fatal("HelloWorld runner: ConcreteLlmAgent is not initialized. Check init() in examples/helloworld/agent.go and ensure GEMINI_API_KEY is set.")
	}
	agentToRun := helloworld.ConcreteLlmAgent

	fmt.Printf("--- Starting Agent: %s ---\n", agentToRun.GetName())
	if agentToRun.GetDescription() != "" {
		fmt.Printf("Description: %s\n", agentToRun.GetDescription())
	}
	fmt.Printf("Model: %s\n", agentToRun.GetModelIdentifier())
	if len(agentToRun.GetTools()) > 0 {
		fmt.Println("Available Tools:")
		for _, tool := range agentToRun.GetTools() {
			fmt.Printf("  - %s: %s\n", tool.Name(), tool.Description())
		}
	} else {
		fmt.Println("No tools configured for this agent.")
	}
	fmt.Println("Type 'exit' or 'quit' to stop.")
	fmt.Println("------------------------------------")

	var history []modelstypes.Content
	scanner := bufio.NewScanner(os.Stdin)

	for {
		// Check if the context has been cancelled at the start of each loop iteration.
		select {
		case <-ctx.Done():
			log.Println("Runner loop: Context cancelled. Exiting loop.")
			// When context is done, main function will return, leading to program exit.
			return
		default:
			// Continue if context is not yet done.
		}

		fmt.Print("[user]: ")
		if !scanner.Scan() {
			// scanner.Scan() returns false on EOF or error, including when os.Stdin is closed
			// due to context cancellation affecting underlying operations or a direct EOF.
			if ctx.Err() != nil {
				log.Println("Runner loop: Context cancelled during input scan. Exiting.")
			} else if err := scanner.Err(); err != nil {
				log.Printf("Input error: %v. Exiting.", err)
			} else {
				log.Println("EOF received on input. Exiting.")
			}
			return // Exit main loop and function.
		}
		userInputText := strings.TrimSpace(scanner.Text())

		if strings.ToLower(userInputText) == "exit" || strings.ToLower(userInputText) == "quit" {
			fmt.Println("User requested exit.")
			cancel() // Signal cancellation.
			// The select statement at the top of the loop will catch this on the next iteration,
			// or the signal handler's os.Exit will terminate.
			// Explicit return here also works to exit the loop immediately.
			return
		}
		if userInputText == "" {
			continue
		}

		currentUserContent := modelstypes.Content{
			Role:  "user",
			Parts: []modelstypes.Part{{Text: &userInputText}},
		}

		// The agent's Process method should also respect the passed context.
		agentResponseContent, err := agentToRun.Process(ctx, history, currentUserContent)
		if err != nil {
			if ctx.Err() != nil { // Check if the error from Process was due to context cancellation
				log.Printf("Agent Process call failed due to context cancellation: %v", err)
				return // Exit loop
			}
			log.Printf("Error from agent.Process: %v", err)
			fmt.Printf("[%s-error]: I encountered an issue: %v\n", agentToRun.GetName(), err)
			history = append(history, currentUserContent)
			continue
		}

		history = append(history, currentUserContent)
		if agentResponseContent != nil {
			history = append(history, *agentResponseContent)
		}

		if agentResponseContent != nil && len(agentResponseContent.Parts) > 0 {
			var responseTexts []string
			for _, part := range agentResponseContent.Parts {
				if part.Text != nil {
					responseTexts = append(responseTexts, *part.Text)
				}
				// This check is for debugging; ideally, Process handles all FunctionCalls internally.
				if part.FunctionCall != nil {
					log.Printf("Runner: Agent response unexpectedly contained FunctionCall: Name=%s.", part.FunctionCall.Name)
				}
			}
			fmt.Printf("[%s]: %s\n", agentToRun.GetName(), strings.Join(responseTexts, "\n"))
		} else {
			fmt.Printf("[%s]: (Agent returned no displayable content)\n", agentToRun.GetName())
		}

		const maxHistoryTurns = 5
		if len(history) > maxHistoryTurns*2 { // Each turn has user + model message
			history = history[len(history)-(maxHistoryTurns*2):]
		}
	}
	// This log might not be reached if os.Exit is called by the signal handler.
	// log.Println("HelloWorld Agent Runner loop finished normally.")
}
