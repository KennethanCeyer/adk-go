package processors

import (
	"fmt"
	"log"
	"strings"

	"github.com/KennethanCeyer/adk-go/agents/invocation"
	"github.com/KennethanCeyer/adk-go/events"
	"github.com/KennethanCeyer/adk-go/models"
	"github.com/KennethanCeyer/adk-go/models/types"
)

type IdentityLlmRequestProcessor struct{}

func (p *IdentityLlmRequestProcessor) RunAsync(invocationCtx *invocation.InvocationContext, llmReq *models.LlmRequest) (<-chan *events.Event, error) {
	outCh := make(chan *events.Event)
	go func() {
		defer close(outCh)

		if invocationCtx == nil || invocationCtx.Agent == nil {
			log.Println("Error: InvocationContext or Agent is nil in IdentityLlmRequestProcessor")
			return
		}
		agent := invocationCtx.Agent

		var instructions []string
		instructions = append(instructions, fmt.Sprintf("You are an agent. Your internal name is \"%s\".", agent.GetName()))
		if desc := agent.GetDescription(); desc != "" {
			instructions = append(instructions, fmt.Sprintf(" The description about you is \"%s\"", desc))
		}

		if llmReq.Config == nil {
			llmReq.Config = &models.GenerateContentConfig{}
		}
		if llmReq.Config.SystemInstruction == nil {
			llmReq.Config.SystemInstruction = &models.Content{}
		}
		currentSysInstructionText := ""
		if len(llmReq.Config.SystemInstruction.Parts) > 0 && llmReq.Config.SystemInstruction.Parts[0].Text != nil {
			currentSysInstructionText = *llmReq.Config.SystemInstruction.Parts[0].Text
		}

		addedInstruction := strings.Join(instructions, "")
		if currentSysInstructionText != "" {
			currentSysInstructionText += "\n\n"
		}
		newInstructionText := currentSysInstructionText + addedInstruction
		llmReq.Config.SystemInstruction.Parts = []types.Part{{Text: &newInstructionText}}
	}()
	return outCh, nil
}
