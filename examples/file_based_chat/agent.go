package file_based_chat

import (
	"log"

	"github.com/KennethanCeyer/adk-go/agents"
	"github.com/KennethanCeyer/adk-go/examples"
	"github.com/KennethanCeyer/adk-go/llmproviders"
	"github.com/KennethanCeyer/adk-go/models/types"
	"github.com/KennethanCeyer/adk-go/tools"
)

func init() {
	provider, err := llmproviders.NewGeminiLLMProvider()
	if err != nil {
		log.Printf("failed to create gemini provider for file_based_chat: %v", err)
		return
	}

	systemText := "You are a helpful assistant with the ability to read and write files. Use the provided tools to manage files based on the user's request."
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

	examples.RegisterAgent("file_based_chat", agent)
}
