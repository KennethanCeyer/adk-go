package web

import (
	_ "embed"
	"log"
	"net/http"

	"github.com/KennethanCeyer/adk-go/examples"
)

//go:embed index.html
var indexHTML []byte

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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
	})

	log.Printf("Starting web server for agent '%s' on %s", agentName, addr)
	log.Printf("Open http://localhost%s in your browser.", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("ListenAndServe error: %v", err)
	}
}
