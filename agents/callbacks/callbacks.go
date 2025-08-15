package callbacks

import (
	"github.com/KennethanCeyer/adk-go/models"
	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
	"github.com/KennethanCeyer/adk-go/sessions"
	"github.com/KennethanCeyer/adk-go/tools"
)

type CallbackContext struct {
	AgentName    string
	InvocationID string
	SessionState *sessions.Session
	UserContent  *modelstypes.Message
}

// Callback function types
type BeforeAgentCallback func(ctx *CallbackContext) *modelstypes.Message
type AfterAgentCallback func(ctx *CallbackContext, finalResponse *modelstypes.Message) *modelstypes.Message

type BeforeModelCallback func(ctx *CallbackContext, req *models.LlmRequest) *models.LlmResponse
type AfterModelCallback func(ctx *CallbackContext, res *models.LlmResponse) *models.LlmResponse

type BeforeToolCallback func(ctx *CallbackContext, tool tools.Tool, args map[string]any) map[string]any
type AfterToolCallback func(ctx *CallbackContext, tool tools.Tool, args map[string]any, result map[string]any) map[string]any
