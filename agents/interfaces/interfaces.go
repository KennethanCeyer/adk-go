package interfaces

import (
	"context"

	"github.com/KennethanCeyer/adk-go/llmproviders"
	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
	"github.com/KennethanCeyer/adk-go/tools"
)

// Agent is the base interface for all agents.
type Agent interface {
	GetName() string
	GetDescription() string
	// Other base agent methods like ParentAgent, SubAgents, etc. can be added here.
}

// LlmAgent defines agents that interact with LLMs.
type LlmAgent interface {
	Agent // Embeds base Agent capabilities

	GetModelIdentifier() string
	GetSystemInstruction() *modelstypes.Message
	GetTools() []tools.Tool
	GetLLMProvider() llmproviders.LLMProvider

	Process(
		ctx context.Context,
		history []modelstypes.Message,
		latestMessage modelstypes.Message,
	) (*modelstypes.Message, error)
}
