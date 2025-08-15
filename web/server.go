package web

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/KennethanCeyer/adk-go/examples"
)

//go:embed static
var staticFiles embed.FS

// StartServer initializes and starts the web server.
func StartServer(addr, agentName string) {
	agent, found := examples.GetAgent(agentName)
	if !found {
		log.Fatalf("Error: Agent '%s' not found for web server.", agentName)
	}
	if agent == nil {
		log.Fatalf("Agent '%s' is not initialized. Check the corresponding examples/ package and ensure GEMINI_API_KEY is set.", agentName)
	}

	// Create a new handler for each request, passing the agent instance.
	handler := NewWebSocketHandler(agent)

	http.HandleFunc("/ws", handler.ServeWS)

	// Serve static files from the embedded filesystem.
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("Failed to create sub filesystem for static files: %v", err)
	}
	http.Handle("/", http.FileServer(http.FS(staticFS)))

	log.Printf("Starting web server for agent '%s' on %s", agentName, addr)
	log.Printf("Open http://localhost%s in your browser.", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("ListenAndServe error: %v", err)
	}
}
