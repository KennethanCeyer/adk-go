package sessions

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
	"github.com/google/uuid"
)

var store = &SessionStore{
	mu:              sync.RWMutex{},
	sessions:        make(map[string]*Session),
	sessionsByAgent: make(map[string]map[string]struct{}),
}

type SessionStore struct {
	mu              sync.RWMutex
	sessions        map[string]*Session
	sessionsByAgent map[string]map[string]struct{} // agentName -> sessionID -> empty struct
}

func GetOrCreate(agentName, sessionID string) *Session {
	store.mu.Lock()
	defer store.mu.Unlock()

	if sessionID != "" {
		if sess, found := store.sessions[sessionID]; found {
			return sess
		}
	}

	newID := uuid.New().String()
	sess := &Session{
		ID:             newID,
		AgentName:      agentName,
		State:          make(map[string]any),
		History:        []modelstypes.Message{},
		LastUpdateTime: time.Now(),
	}

	store.sessions[newID] = sess
	if _, ok := store.sessionsByAgent[agentName]; !ok {
		store.sessionsByAgent[agentName] = make(map[string]struct{})
	}
	store.sessionsByAgent[agentName][newID] = struct{}{}
	return sess
}

func Get(sessionID string) (*Session, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	sess, found := store.sessions[sessionID]
	if !found {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	return sess, nil
}

func Save(sess *Session) {
	store.mu.Lock()
	defer store.mu.Unlock()
	sess.LastUpdateTime = time.Now()
	store.sessions[sess.ID] = sess
}

func ListByAgent(agentName string) []string {
	store.mu.RLock()
	defer store.mu.RUnlock()

	agentSessionMap, found := store.sessionsByAgent[agentName]
	if !found {
		return []string{}
	}

	// Collect all sessions for the agent
	agentSessions := make([]*Session, 0, len(agentSessionMap))
	for id := range agentSessionMap {
		if sess, ok := store.sessions[id]; ok {
			agentSessions = append(agentSessions, sess)
		}
	}

	// Sort sessions by LastUpdateTime in descending order (newest first)
	sort.Slice(agentSessions, func(i, j int) bool {
		return agentSessions[i].LastUpdateTime.After(agentSessions[j].LastUpdateTime)
	})

	// Extract the sorted IDs
	ids := make([]string, 0, len(agentSessions))
	for _, sess := range agentSessions {
		ids = append(ids, sess.ID)
	}
	return ids
}

func Delete(sessionID string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	session, found := store.sessions[sessionID]
	if !found {
		return nil // Not an error if it's already gone.
	}

	delete(store.sessions, sessionID)

	if agentSessions, ok := store.sessionsByAgent[session.AgentName]; ok {
		delete(agentSessions, session.ID)
		if len(agentSessions) == 0 {
			delete(store.sessionsByAgent, session.AgentName)
		}
	}
	log.Printf("Deleted session: %s", sessionID)
	return nil
}
