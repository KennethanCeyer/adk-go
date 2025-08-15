package web

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/KennethanCeyer/adk-go/examples"
	"github.com/KennethanCeyer/adk-go/sessions"
	"github.com/KennethanCeyer/adk-go/web/graph"
)

//go:embed index.html
var indexHTML []byte

//go:embed logo.png
var logoPNG []byte

// StartServer initializes and starts the web server.
func StartServer(addr string) {
	// The WebSocket handler is now created per-connection, based on the requested agent.
	http.HandleFunc("/ws", serveWS)

	http.HandleFunc("/api/sessions", handleListSessions)
	http.HandleFunc("/api/session/", handleGetSessionState) // Note the trailing slash
	http.HandleFunc("/api/details", handleDetails)
	http.HandleFunc("/api/graph", handleAgentGraph)
	http.HandleFunc("/api/agents", handleListAgents)

	// Serve the embedded logo.
	http.HandleFunc("/logo.png", handleLogo)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(indexHTML)
	})

	log.Printf("Starting web server on %s", addr)
	log.Printf("Open http://localhost%s in your browser.", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("ListenAndServe error: %v", err)
	}
}

func handleListAgents(w http.ResponseWriter, r *http.Request) {
	agentNames := examples.ListAgents()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(agentNames); err != nil {
		log.Printf("Error encoding agent list: %v", err)
		http.Error(w, "Failed to list agents", http.StatusInternalServerError)
	}
}

func handleListSessions(w http.ResponseWriter, r *http.Request) {
	agentName := r.URL.Query().Get("agent")
	if agentName == "" {
		http.Error(w, "Missing 'agent' query parameter", http.StatusBadRequest)
		return
	}
	sessionIDs := sessions.ListByAgent(agentName)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sessionIDs); err != nil {
		log.Printf("Error encoding session list: %v", err)
		http.Error(w, "Failed to list sessions", http.StatusInternalServerError)
	}
}

func handleGetSessionState(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/session/"), "/")
	if len(parts) < 1 || parts[0] == "" {
		http.Error(w, "Missing session ID", http.StatusBadRequest)
		return
	}
	sessionID := parts[0]

	session, err := sessions.Get(sessionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Session with ID '%s' not found", sessionID), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(session.State); err != nil {
		log.Printf("Error encoding session state: %v", err)
		http.Error(w, "Failed to encode session state", http.StatusInternalServerError)
	}
}

func handleDetails(w http.ResponseWriter, r *http.Request) {
	agentName := r.URL.Query().Get("agent")
	toolName := r.URL.Query().Get("tool")

	if agentName == "" {
		http.Error(w, "Missing 'agent' query parameter", http.StatusBadRequest)
		return
	}

	agent, found := examples.GetAgent(agentName)
	if !found {
		http.Error(w, "Agent not found", http.StatusNotFound)
		return
	}

	var details struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"`
	}

	if toolName != "" {
		for _, tool := range agent.GetTools() {
			if tool.Name() == toolName {
				details.Name = tool.Name()
				details.Description = tool.Description()
				details.Type = "Tool"
				break
			}
		}
	} else {
		details.Name = agent.GetName()
		details.Description = agent.GetDescription()
		details.Type = agent.GetModelIdentifier()
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(details); err != nil {
		http.Error(w, "Failed to encode details", http.StatusInternalServerError)
	}
}

func handleAgentGraph(w http.ResponseWriter, r *http.Request) {
	agentName := r.URL.Query().Get("agent")
	if agentName == "" {
		http.Error(w, "Missing 'agent' query parameter", http.StatusBadRequest)
		return
	}
	agent, found := examples.GetAgent(agentName)
	if !found {
		http.Error(w, "Agent not found", http.StatusNotFound)
		return
	}
	dotSource := graph.Build(agent)
	w.Header().Set("Content-Type", "text/vnd.graphviz")
	_, _ = w.Write([]byte(dotSource))
}

func handleLogo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	_, _ = w.Write(logoPNG)
}

// serveWS handles WebSocket requests from the peer.
func serveWS(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	agentName := query.Get("agent")
	sessionID := query.Get("sessionId")

	if agentName == "" { http.Error(w, "Missing 'agent' query parameter", http.StatusBadRequest); return }

	agent, found := examples.GetAgent(agentName)
	if !found || agent == nil {
		log.Printf("WebSocket connection request for unknown or uninitialized agent: %s", agentName)
		http.Error(w, fmt.Sprintf("Agent '%s' not found or not initialized", agentName), http.StatusNotFound)
		return
	}

	// Get an existing session for the agent, or create a new one.
	// This allows users to continue conversations by using the same session ID.
	currentSession := sessions.GetOrCreate(agentName, sessionID)
	log.Printf("WebSocket connected for agent '%s' with session ID '%s'", agentName, currentSession.ID)

	handler := NewWebSocketHandler(agent, currentSession)
	handler.ServeWS(w, r)
}
