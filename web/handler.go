package web

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/KennethanCeyer/adk-go/agents/interfaces"
	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
	"github.com/gorilla/websocket"
)

const maxHistoryTurns = 20 // Limit conversation history to the last 20 turns (40 messages)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections for this example.
		// In a production environment, you should implement proper origin checking.
		return true
	},
}

// WebSocketHandler handles WebSocket connections.
type WebSocketHandler struct {
	agent interfaces.LlmAgent
}

// NewWebSocketHandler creates a new WebSocketHandler.
func NewWebSocketHandler(agent interfaces.LlmAgent) *WebSocketHandler {
	return &WebSocketHandler{agent: agent}
}

// ServeWS handles WebSocket requests from the peer.
func (h *WebSocketHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	log.Println("Client connected to WebSocket.")

	// Send an initial system message to identify the agent to the client.
	initialMessage := fmt.Sprintf("system:agent_name:%s", h.agent.GetName())
	if err := conn.WriteMessage(websocket.TextMessage, []byte(initialMessage)); err != nil {
		log.Printf("Failed to send initial agent name: %v", err)
		return
	}

	// For this example, each connection has its own conversation history.
	var history []modelstypes.Message

	for {
		// Read message from browser
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Read error: %v", err)
			}
			break // Exit loop on error
		}

		log.Printf("Received from client: %s", msg)
		userInputText := string(msg)
		userMessage := modelstypes.Message{Role: "user", Parts: []modelstypes.Part{{Text: &userInputText}}}

		// Process message with the agent
		agentResponse, err := h.agent.Process(context.Background(), history, userMessage)
		if err != nil {
			log.Printf("Agent processing error: %v", err)
			errMsg := []byte("Agent error: " + err.Error())
			if writeErr := conn.WriteMessage(websocket.TextMessage, errMsg); writeErr != nil {
				break
			}
			continue
		}

		history = append(history, userMessage)
		if agentResponse != nil {
			history = append(history, *agentResponse)
		}

		// Prune history to prevent it from growing indefinitely.
		if len(history) > maxHistoryTurns*2 {
			history = history[len(history)-(maxHistoryTurns*2):]
		}

		var responseText string
		if agentResponse != nil && len(agentResponse.Parts) > 0 && agentResponse.Parts[0].Text != nil {
			responseText = *agentResponse.Parts[0].Text
		}

		if err := conn.WriteMessage(websocket.TextMessage, []byte(responseText)); err != nil {
			break
		}
	}
	log.Println("Client disconnected.")
}
