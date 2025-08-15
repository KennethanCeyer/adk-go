package examples

import (
	"log"
	"sort"
	"sync"

	"github.com/KennethanCeyer/adk-go/agents/interfaces"
)

var (
	agentRegistry = make(map[string]interfaces.LlmAgent)
	registryLock  = &sync.RWMutex{}
)

// RegisterAgent adds an agent to the central registry.
// This function is intended to be called from the init() function of each agent example package.
func RegisterAgent(name string, agent interfaces.LlmAgent) {
	registryLock.Lock()
	defer registryLock.Unlock()
	if _, exists := agentRegistry[name]; exists {
		log.Printf("Warning: Agent with name '%s' is already registered. Overwriting.", name)
	}
	agentRegistry[name] = agent
	log.Printf("Agent '%s' registered successfully.", name)
}

// GetAgent retrieves an agent from the registry by name.
func GetAgent(name string) (interfaces.LlmAgent, bool) {
	registryLock.RLock()
	defer registryLock.RUnlock()
	agent, found := agentRegistry[name]
	return agent, found
}

// ListAgents returns a sorted list of all registered agent names.
func ListAgents() []string {
	registryLock.RLock()
	defer registryLock.RUnlock()
	names := make([]string, 0, len(agentRegistry))
	for name := range agentRegistry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
