package llmproviders

import (
	"context"

	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
	"github.com/KennethanCeyer/adk-go/tools"
)

// LLMProvider defines the interface for interacting with a Large Language Model.
type LLMProvider interface {
	GenerateContent(
		ctx context.Context,
		modelName string,
		systemInstruction *modelstypes.Message,
		tools []tools.Tool,
		history []modelstypes.Message,
		latestMessage modelstypes.Message,
	) (*modelstypes.Message, error)
}
