package sessions

import (
	"github.com/KennethanCeyer/adk-go/models/types"
	"github.com/google/uuid"
)

// Session holds the state of a single conversation with an agent.
type Session struct {
	ID        string          `json:"id"`
	AgentName string          `json:"agent_name"`
	History   []types.Message `json:"history"`
	State     map[string]any  `json:"state"`
}

// New creates a new Session with a unique ID.
func New(agentName string) *Session {
	return &Session{
		ID:        uuid.NewString(),
		AgentName: agentName,
		History:   []types.Message{},
		State:     make(map[string]any),
	}
}

// AddMessage appends a message to the session's history.
func (s *Session) AddMessage(msg types.Message) { s.History = append(s.History, msg) }

// GetHistory returns the conversation history.
func (s *Session) GetHistory() []types.Message { return s.History }

// PruneHistory limits the history to a specific number of turns.
func (s *Session) PruneHistory(maxTurns int) {
	maxMessages := maxTurns * 2
	if len(s.History) > maxMessages {
		s.History = s.History[len(s.History)-maxMessages:]
	}
}
