package sessions

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// SessionStore defines the interface for storing and retrieving conversation sessions.
type SessionStore interface {
	Get(id string) (*Session, bool)
	Save(session *Session)
	ListByAgent(agentName string) []string
}

const sessionDir = ".sessions"

// globalStore is now a file-based store.
var globalStore = NewFileSessionStore(sessionDir)

// FileSessionStore is a thread-safe, file-based implementation of SessionStore.
type FileSessionStore struct {
	basePath string
	lock     sync.RWMutex // Used to protect access to the filesystem map.
}

// NewFileSessionStore creates a new file-based session store.
func NewFileSessionStore(basePath string) *FileSessionStore {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		log.Fatalf("Failed to create session directory at %s: %v", basePath, err)
	}
	return &FileSessionStore{
		basePath: basePath,
	}
}

// Get retrieves a session by its ID.
func (s *FileSessionStore) Get(id string) (*Session, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	filePath := filepath.Join(s.basePath, id+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false
		}
		log.Printf("Error reading session file %s: %v", filePath, err)
		return nil, false
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		log.Printf("Error unmarshaling session file %s: %v", filePath, err)
		return nil, false
	}
	return &session, true
}

// Save stores a session.
func (s *FileSessionStore) Save(session *Session) {
	s.lock.Lock()
	defer s.lock.Unlock()

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		log.Printf("Error marshaling session %s: %v", session.ID, err)
		return
	}

	filePath := filepath.Join(s.basePath, session.ID+".json")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		log.Printf("Error writing session file %s: %v", filePath, err)
	}
}

// ListByAgent returns a list of session IDs for a given agent, sorted by modification time.
func (s *FileSessionStore) ListByAgent(agentName string) []string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	type sessionInfo struct {
		ID      string
		ModTime time.Time
	}
	var sessions []sessionInfo

	files, err := os.ReadDir(s.basePath)
	if err != nil {
		log.Printf("Error reading session directory %s: %v", s.basePath, err)
		return nil
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}
		filePath := filepath.Join(s.basePath, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		var sess Session
		if err := json.Unmarshal(data, &sess); err == nil && sess.AgentName == agentName {
			info, _ := file.Info()
			sessions = append(sessions, sessionInfo{ID: sess.ID, ModTime: info.ModTime()})
		}
	}

	sort.Slice(sessions, func(i, j int) bool { return sessions[i].ModTime.After(sessions[j].ModTime) })

	var ids []string
	for _, s := range sessions {
		ids = append(ids, s.ID)
	}
	return ids
}

// GetOrCreate retrieves a session by ID, or creates a new one if the ID is empty or not found.
func GetOrCreate(agentName, id string) *Session {
	if s, found := globalStore.Get(id); id != "" && found {
		// Ensure the session belongs to the correct agent to prevent mix-ups.
		if s.AgentName == agentName {
			return s
		}
	}
	s := New(agentName)
	globalStore.Save(s)
	return s
}

// Get retrieves a session by ID from the global store, returning an error if not found.
func Get(id string) (*Session, error) { if s, found := globalStore.Get(id); found { return s, nil }; return nil, fmt.Errorf("session with ID '%s' not found", id) }

// Save saves a session to the global store.
func Save(s *Session) {
	globalStore.Save(s)
}

// ListByAgent lists sessions for an agent from the global store.
func ListByAgent(agentName string) []string {
	return globalStore.ListByAgent(agentName)
}
