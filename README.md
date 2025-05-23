# ADK-Go: Agent Development Kit in Go (Migration In Progress)

<html>
    <h2 align="center">
      <img src="https://raw.githubusercontent.com/google/adk-python/main/assets/agent-development-kit.png" width="256"/>
    </h2>
    <h3 align="center">
      An open-source, code-first Go toolkit for building, evaluating, and deploying sophisticated AI agents with flexibility and control.
    </h3>
    <h3 align="center">
      Important Links:
      <a href="https://google.github.io/adk-docs/">Docs</a>
    </h3>
</html>

👋 Welcome to ADK-Go!

This repository is currently undergoing a migration from its original Python-based Agent Development Kit (ADK) to Go. The primary goal is to create a robust, performant, and idiomatic Go version of the ADK.

**Please Note**: This is an active development project. Many features from the Python ADK are not yet implemented or are in the process of being ported. The initial focus is on establishing a core, runnable agent interaction loop.

This README provides instructions to set up and run a **foundational "Hello World" agent example**. This example demonstrates the basic structure, core ADK concepts in Go, and interaction with Google's Gemini LLM.

## Target Project Structure (For Runnable HelloWorld)

While the repository contains more files from the ongoing migration, the "Hello World" example relies on the following core structure:

```plaintext
adk-go/
├── cmd/
│   └── helloworld_runner/
│       └── main.go         # Executable for the HelloWorld agent
├── adk/
│   ├── agent.go            # Core Agent, Tool, LLMProvider interfaces; BaseAgent impl.
│   ├── runner.go           # SimpleCLIRunner for command-line interaction
│   └── types/
│       └── types.go        # Defines Message, Part, FunctionCall, FunctionResponse
├── examples/
│   └── helloworld/
│       └── agent.go        # HelloWorld agent definition and initialization
├── llmproviders/
│   └── gemini.go           # Gemini LLMProvider implementation
├── tools/
│   └── rolldie.go          # RollDieTool implementation
├── go.mod                  # Manages project dependencies
├── go.sum                  # Checksums for dependencies
└── README.md
```

As the migration progresses, other directories and files from your provided list (like `flows`, `auth`, `events`, etc.) will be populated and integrated.

Prerequisites

- **Go**: Version 1.20 or later. ([Installation Guide](https://go.dev/doc/install))
- **Git**: For cloning the repository.
- **Google Cloud Project**: A Google Cloud Project with the Vertex AI API (or Generative Language API via Google AI Studio) enabled.
- **Gemini API Key**: An API key for Google Gemini.
  - Important: Secure your API key. Do not commit it directly into code. Use environment variables.
- **Environment Variable**: Set the `GEMINI_API_KEY` environment variable.

## Setup & Running the HelloWorld Agent

1. **Clone the Repository**

   Open your terminal and clone the project repository:

   ```bash
   git clone https://github.com/KennethanCeyer/adk-go.git
   cd adk-go
   ```

2. **Tidy Dependencies**

   Ensure all necessary Go modules are downloaded and the `go.mod` and `go.sum` files are up-to-date.

   ```bash
   go mod tidy
   ```

   This command will automatically download `github.com/google/generative-ai-go/genai` and its dependencies based on the imports in the Go files.

3. **Set `GEMINI_API_KEY` Environment Variable (if not already done, as described in "Prerequisites")**

4. **Run the HelloWorld Agent**

   From the `adk-go` root directory

   ```bash
   go run ./cmd/helloworld_runner/main.go
   ```

5. **Interact with the Agent**

   Once the runner starts, it will prompt you for input. Try the following.

   ```bash
   2025/05/22 17:00:10 HelloWorldAgent initialized in examples/helloworld/agent.go.
   2025/05/22 17:00:10 HelloWorld Agent Runner starting... (Agent: HelloWorldAgent, Model: gemini-1.5-flash-latest)
   --- Starting Agent: HelloWorldAgent ---
   Type 'exit' or 'quit' to stop.
   [user]: Hello! Can you roll a 6-sided die for me?
   2025/05/22 17:00:22 RollDieTool: Executed. Rolled 2 (d6)
   [HelloWorldAgent]: You rolled a 2 on a 6-sided die.
   ```

   ![Helloworld example](./docs/helloworld_example.png)

## Next Steps & Contribution

This "Hello World" example serves as the initial building block. The next steps in the migration will involve:

- Implementing more core ADK features (e.g., advanced agent types, session management, event handling, complex flows).
- Porting additional tools and planners.
- Adding comprehensive tests.
