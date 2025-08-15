package invocation

import "github.com/KennethanCeyer/adk-go/agents/interfaces"

// InvocationContext holds the contextual information for a single agent invocation.
type InvocationContext struct {
	// Agent is the instance of the agent being invoked.
	Agent interfaces.LlmAgent
}
