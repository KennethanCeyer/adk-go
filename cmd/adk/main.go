package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/KennethanCeyer/adk-go/adk"
	"github.com/KennethanCeyer/adk-go/examples"
	"github.com/KennethanCeyer/adk-go/sessions"
	"github.com/KennethanCeyer/adk-go/web"

	// Import example packages to trigger their init() functions.
	_ "github.com/KennethanCeyer/adk-go/examples/file_based_chat"
	_ "github.com/KennethanCeyer/adk-go/examples/financial_analyst"
	_ "github.com/KennethanCeyer/adk-go/examples/helloworld"
	_ "github.com/KennethanCeyer/adk-go/examples/looping_guesser"
	_ "github.com/KennethanCeyer/adk-go/examples/parallel_trip_planner"
	_ "github.com/KennethanCeyer/adk-go/examples/sequential_weather"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	// User-friendly check: if the first argument is a flag, they probably forgot the 'run' or 'web' command.
	if strings.HasPrefix(command, "-") {
		fmt.Printf("Error: Missing command. Did you mean 'adk run %s'?\n\n", strings.Join(os.Args[1:], " "))
		printUsage()
		os.Exit(1)
	}

	switch command {
	case "run":
		runCmd(os.Args[2:])
	case "web":
		webCmd(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: adk <command> [arguments]")
	fmt.Println("\nAvailable commands:")
	fmt.Println("  run                Run an agent in the command line")
	fmt.Println("  web                Start a web server with a UI for an agent")
	fmt.Println("\nRun 'adk <command> -h' for more information on a specific command.")
	fmt.Println("\nAvailable agents for 'run' and 'web' commands:")
	fmt.Printf("  %s\n", strings.Join(examples.ListAgents(), ", "))
}

func runCmd(args []string) {
	runFlagSet := flag.NewFlagSet("run", flag.ContinueOnError)
	agentName := newAgentFlag(runFlagSet)
	sessionID := runFlagSet.String("session-id", "", "ID of a previous session to resume.")

	err := runFlagSet.Parse(args)
	if err != nil {
		log.Fatalf("Error parsing flags for run command: %v", err)
	}

	var currentSession *sessions.Session
	if *sessionID != "" {
		log.Printf("Attempting to resume session '%s'...", *sessionID)
		currentSession, err = sessions.Get(*sessionID)
		if err != nil {
			log.Fatalf("Error resuming session: %v", err)
		}
		log.Printf("Successfully resumed session '%s'.", *sessionID)
	} else {
		currentSession = sessions.GetOrCreate(*agentName, *sessionID)
		log.Printf("Started new session with ID: %s", currentSession.ID)
		log.Println("To resume this session later, use: -session-id=" + currentSession.ID)
	}

	log.Printf("Loading '%s' agent...", *agentName)
	agentToRun, found := examples.GetAgent(*agentName)
	if !found {
		log.Printf("Error: Unknown agent name '%s'.", *agentName)
		printUsage()
		os.Exit(1)
	}

	if agentToRun == nil {
		log.Fatalf("Agent '%s' is not initialized. Check the corresponding examples/ package and ensure GEMINI_API_KEY is set.", *agentName)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	runner, err := adk.NewSimpleCLIRunner(agentToRun, currentSession)
	if err != nil {
		log.Fatalf("Failed to create agent runner: %v", err)
	}

	log.Printf("Starting agent runner...")
	runner.Start(ctx)
	log.Println("Agent runner finished.")
}

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

// newAgentFlag adds a standardized -agent flag to a given FlagSet.
func newAgentFlag(fs *flag.FlagSet) *string {
	availableAgents := examples.ListAgents()
	defaultAgent := "helloworld"
	if len(availableAgents) == 0 {
		log.Fatal("No agents are registered. Please check the 'examples' packages.")
	}
	if !contains(availableAgents, defaultAgent) {
		if len(availableAgents) > 0 {
			defaultAgent = availableAgents[0]
		}
	}
	usage := fmt.Sprintf("Name of the agent to run. Available: [%s]", strings.Join(availableAgents, ", "))
	return fs.String("agent", defaultAgent, usage)
}

func webCmd(args []string) {
	webFlagSet := flag.NewFlagSet("web", flag.ContinueOnError)
	port := webFlagSet.String("port", "8080", "Port to run the web server on")

	err := webFlagSet.Parse(args)
	if err != nil {
		log.Fatalf("Error parsing flags for web command: %v", err)
	}

	addr := ":" + *port
	web.StartServer(addr)
}
