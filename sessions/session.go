package sessions

import (
	"time"

	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
)

type Session struct {
	ID             string
	AgentName      string
	State          map[string]any
	History        []modelstypes.Message
	LastUpdateTime time.Time
}

func (s *Session) AddMessage(msg modelstypes.Message) {
	s.History = append(s.History, msg)
}

func (s *Session) PruneHistory(maxTurns int) {
	// Implementation for history pruning can be added here.
}
