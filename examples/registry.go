package examples

import (
	"log"
	"sort"
	"sync"

	"github.com/KennethanCeyer/adk-go/agents/interfaces"
)

type AgentDefinition struct {
	Name      string
	InitError string
}

var (
	mu               sync.RWMutex
	agents           = make(map[string]interfaces.LlmAgent)
	agentDefinitions = make(map[string]*AgentDefinition)
)

func RegisterAgent(name string, agent interfaces.LlmAgent, err error) {
	mu.Lock()
	defer mu.Unlock()

	var errMsg string
	if err != nil {
		errMsg = err.Error()
		log.Printf("Warning: Could not initialize '%s' agent: %v", name, err)
	}

	agentDefinitions[name] = &AgentDefinition{
		Name:      name,
		InitError: errMsg,
	}

	if agent != nil && err == nil {
		agents[name] = agent
	}
}

func GetAgent(name string) (interfaces.LlmAgent, bool) {
	mu.RLock()
	defer mu.RUnlock()
	agent, found := agents[name]
	return agent, found
}

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
