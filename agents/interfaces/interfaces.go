package interfaces

import (
	"context"

	"github.com/KennethanCeyer/adk-go/llmproviders"
	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
	"github.com/KennethanCeyer/adk-go/tools"
)

type Agent interface {
	GetName() string
	GetDescription() string
}

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
