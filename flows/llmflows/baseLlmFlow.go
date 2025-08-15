package llmflows

import (
	"log"
	"reflect"
	"runtime"

	"github.com/KennethanCeyer/adk-go/agents/interfaces"
	"github.com/KennethanCeyer/adk-go/agents/invocation"
	"github.com/KennethanCeyer/adk-go/events"
	"github.com/KennethanCeyer/adk-go/flows/llmflows/processors"
	"github.com/KennethanCeyer/adk-go/models"
)

// BaseLlmFlow is the base for LLM-based flows.
type BaseLlmFlow struct {
	RequestProcessors  []processors.BaseLlmRequestProcessor
	ResponseProcessors []processors.BaseLlmResponseProcessor
}

// NewBaseLlmFlow creates a new BaseLlmFlow.
func NewBaseLlmFlow(
	requestProcessors []processors.BaseLlmRequestProcessor,
	responseProcessors []processors.BaseLlmResponseProcessor) *BaseLlmFlow {
	return &BaseLlmFlow{
		RequestProcessors:  requestProcessors,
		ResponseProcessors: responseProcessors,
	}
}

// RunAsync executes the LLM flow.
func (f *BaseLlmFlow) RunAsync(invocationCtx *invocation.InvocationContext, llmAgent interfaces.LlmAgent) (<-chan *events.Event, error) {
	outputEventCh := make(chan *events.Event)

	go func() {
		defer close(outputEventCh)

		llmReq := &models.LlmRequest{}
		if llmAgent.GetGenerateContentConfig() != nil {
			llmReq.Config = llmAgent.GetGenerateContentConfig()
		} else {
			llmReq.Config = &models.GenerateContentConfig{}
		}

		for _, processor := range f.RequestProcessors {
			processorName := runtime.FuncForPC(reflect.ValueOf(processor).Pointer()).Name()
			log.Printf("Running request processor: %s for agent: %s", processorName, llmAgent.GetName())
			eventCh, err := processor.RunAsync(invocationCtx, llmReq)
			if err != nil {
				// Handle error
				return
			}
			for event := range eventCh {
				outputEventCh <- event
			}
		}

		llmRespCh, err := llmAgent.PredictAsync(invocationCtx.Ctx, invocationCtx, llmReq)
		if err != nil {
			// Handle error
			return
		}

		modelResponseEvent, ok := <-llmRespCh
		if !ok {
			// Handle channel closed
			return
		}

		var llmResponse *models.LlmResponse
		if modelResponseEvent.LlmResponse != nil {
			llmResponse = modelResponseEvent.LlmResponse
		} else if modelResponseEvent.Content != nil {
			llmResponse = &models.LlmResponse{Content: modelResponseEvent.Content}
		} else {
			// Handle no response
			return
		}

		var accumulatedProcessorEvents []*events.Event
		for _, processor := range f.ResponseProcessors {
			processorName := runtime.FuncForPC(reflect.ValueOf(processor).Pointer()).Name()
			log.Printf("Running response processor: %s for agent: %s", processorName, llmAgent.GetName())
			eventCh, err := processor.RunAsync(invocationCtx, llmResponse, modelResponseEvent)
			if err != nil {
				// Handle error
				continue
			}
			for procEvent := range eventCh {
				accumulatedProcessorEvents = append(accumulatedProcessorEvents, procEvent)
			}
		}

		outputEventCh <- modelResponseEvent
		for _, procEvent := range accumulatedProcessorEvents {
			outputEventCh <- procEvent
		}
	}()

	return outputEventCh, nil
}
