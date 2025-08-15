package models

import (
	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
	"github.com/KennethanCeyer/adk-go/tools"
)

type LlmRequest struct {
	ModelIdentifier   string
	SystemInstruction *modelstypes.Message
	Tools             []tools.Tool
	History           []modelstypes.Message
	LatestMessage     modelstypes.Message
}

type LlmResponse struct {
	Content *modelstypes.Message
}
