package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/KennethanCeyer/adk-go/adk"
	"github.com/KennethanCeyer/adk-go/agents"
	"github.com/KennethanCeyer/adk-go/examples/helloworld"
	"github.com/KennethanCeyer/adk-go/examples/sequential_weather"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "run":
		runCmd(os.Args[2:])
	// case "web":
	// 	// Placeholder for 'adk web'
	// 	fmt.Println("'web' command not yet implemented.")
	// case "eval":
	// 	// Placeholder for 'adk eval'
	// 	fmt.Println("'eval' command not yet implemented.")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: adk <command> [arguments]")
	fmt.Println("\nAvailable commands:")
	fmt.Println("  run                Run an agent")
	// fmt.Println("  web                Start the ADK web UI")
	// fmt.Println("  eval               Run evaluations for an agent")
	fmt.Println("\nFlags for run command:")
	fmt.Println("  -agent <name>      Name of the agent to run (helloworld or sequential_weather)")
}

func runCmd(args []string) {
	runFlagSet := flag.NewFlagSet("run", flag.ExitOnError)
	agentName := runFlagSet.String("agent", "helloworld", "Name of the agent to run (helloworld or sequential_weather)")

	err := runFlagSet.Parse(args)
	if err != nil {
		log.Fatalf("Error parsing flags for run command: %v", err)
	}

	var agentToRun agents.LlmAgent

	switch *agentName {
	case "helloworld":
		log.Println("Loading 'helloworld' agent...")
		agentToRun = helloworld.ConcreteLlmAgent
	case "sequential_weather":
		log.Println("Loading 'sequential_weather' agent...")
		agentToRun = sequential_weather.SequentialWeatherAgent
	default:
		log.Printf("Error: Unknown agent name '%s'", *agentName)
		printUsage()
		os.Exit(1)
	}

	if agentToRun == nil {
		log.Fatalf("Agent '%s' is not initialized. Check the corresponding examples/ package and ensure GEMINI_API_KEY is set.", *agentName)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	runner, err := adk.NewSimpleCLIRunner(agentToRun)
	if err != nil {
		log.Fatalf("Failed to create agent runner: %v", err)
	}

	log.Printf("Starting agent runner...")
	runner.Start(ctx)
	log.Println("Agent runner finished.")
}
