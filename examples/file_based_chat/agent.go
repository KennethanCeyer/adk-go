package file_based_chat

import (
	"github.com/KennethanCeyer/adk-go/agents"
	"github.com/KennethanCeyer/adk-go/examples"
	"github.com/KennethanCeyer/adk-go/llmproviders"
	"github.com/KennethanCeyer/adk-go/models/types"
	"github.com/KennethanCeyer/adk-go/tools"
)

func init() {
	provider, err := llmproviders.NewGeminiLLMProvider()
	if err != nil {
		// Register the agent with its initialization error.
		examples.RegisterAgent("file_based_chat", nil, err)
		return
	}

	systemText := "You are a helpful assistant that specializes in reading and writing local files. When the conversation starts with a simple greeting, introduce yourself and your capabilities. For example: 'Hello! I can read and write files for you. You can ask me to do things like: \\\"read notes.txt\\\" or \\\"write 'Hello World' to a new file named welcome.txt\\\". What would you like to do?'. For other requests, use the provided tools to manage files as requested by the user."
	systemInstruction := &types.Message{
		Role: "system",
		Parts: []types.Part{
			{Text: &systemText},
		},
	}

	agent := agents.NewBaseLlmAgent(
		"file_based_chat",
		"An agent that can read and write local files.",
		"gemini-2.5-flash",
		systemInstruction,
		provider,
		[]tools.Tool{NewReadFileTool(), NewWriteFileTool()},
	)

	examples.RegisterAgent("file_based_chat", agent, nil)
}
