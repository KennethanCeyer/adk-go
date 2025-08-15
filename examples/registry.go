package examples

import (
	"log"
	"sort"
	"sync"

	"github.com/KennethanCeyer/adk-go/agents/interfaces"
)

// AgentDefinition holds the metadata for an agent, including any initialization error.
type AgentDefinition struct {
	Name      string
	InitError string
}

var (
	mu               sync.RWMutex
	agents           = make(map[string]interfaces.LlmAgent)
	agentDefinitions = make(map[string]*AgentDefinition)
)

// RegisterAgent attempts to register an agent. If the agent is nil (due to an init error),
// it still records the definition with the error message so the UI can show its status.
func RegisterAgent(name string, agent interfaces.LlmAgent, err error) {
	mu.Lock()
	defer mu.Unlock()

	var errMsg string
	if err != nil {
		errMsg = err.Error()
		// This log matches the format from previous issues, which is helpful for debugging.
		log.Printf("Warning: Could not initialize '%s' agent: %v", name, err)
	}

	// Always register the definition so the UI can see it.
	agentDefinitions[name] = &AgentDefinition{
		Name:      name,
		InitError: errMsg,
	}

	// Only register the agent instance if it was created successfully.
	if agent != nil && err == nil {
		agents[name] = agent
	}
}

// GetAgent retrieves an agent from the registry by name.
func GetAgent(name string) (interfaces.LlmAgent, bool) {
	mu.RLock()
	defer mu.RUnlock()
	agent, found := agents[name]
	return agent, found
}

// ListAgents returns a sorted list of all *successfully initialized* agent names.
func ListAgents() []string {
	mu.RLock()
	defer mu.RUnlock()
	names := make([]string, 0, len(agents))
	for name := range agents {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetAllAgentDefinitions returns all registered agent definitions, including those that failed to initialize.
// This is used by the web UI to show the status of all agents.
func GetAllAgentDefinitions() []*AgentDefinition {
	mu.RLock()
	defer mu.RUnlock()
	defs := make([]*AgentDefinition, 0, len(agentDefinitions))
	for _, def := range agentDefinitions {
		defs = append(defs, def)
	}
	// Sort for consistent ordering in the UI
	sort.Slice(defs, func(i, j int) bool {
		return defs[i].Name < defs[j].Name
	})
	return defs
}
