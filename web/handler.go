package web

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/KennethanCeyer/adk-go/agents/interfaces"
	"github.com/KennethanCeyer/adk-go/agents/invocation"
	modelstypes "github.com/KennethanCeyer/adk-go/models/types"
	"github.com/KennethanCeyer/adk-go/sessions"
	"github.com/gorilla/websocket"
)

type UIMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

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
	sess  *sessions.Session
	conn  *websocket.Conn
	mu    sync.Mutex // Protects concurrent writes to the WebSocket connection
}

// NewWebSocketHandler creates a new WebSocketHandler.
func NewWebSocketHandler(agent interfaces.LlmAgent, sess *sessions.Session) *WebSocketHandler {
	if sess.State == nil {
		sess.State = make(map[string]any)
	}
	return &WebSocketHandler{agent: agent, sess: sess}
}

// ServeWS handles WebSocket requests from the peer.
func (h *WebSocketHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()
	h.conn = conn

	log.Println("Client connected to WebSocket.")

	if err := h.sendInitialState(); err != nil {
		return
	}

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Read error: %v", err)
			}
			break
		}

		h.handleIncomingMessage(r.Context(), p)
	}
	log.Println("Client disconnected.")
}

// sendJSON safely sends a JSON message over the WebSocket connection.
func (h *WebSocketHandler) sendJSON(messageType string, payload any) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	msg := UIMessage{Type: messageType, Payload: payload}
	return h.conn.WriteJSON(msg)
}

func (h *WebSocketHandler) sendInitialState() error {
	infoPayload := map[string]string{
		"agentName":        h.agent.GetName(),
		"agentDescription": h.agent.GetDescription(),
		"agentType":        h.agent.GetModelIdentifier(),
		"sessionId":        h.sess.ID,
	}
	if err := h.sendJSON("system_info", infoPayload); err != nil {
		log.Printf("Error sending system info: %v", err)
		return err
	}

	if err := h.sendJSON("history", h.sess.History); err != nil {
		log.Printf("Error sending history: %v", err)
		return err
	}
	return nil
}

func (h *WebSocketHandler) handleIncomingMessage(ctx context.Context, msgBytes []byte) {
	log.Printf("Received from client: %s", msgBytes)
	userInputText := string(msgBytes)
	userMessage := modelstypes.Message{Role: "user", Parts: []modelstypes.Part{{Text: &userInputText}}}

	// Echo user message back to UI for immediate rendering.
	if err := h.sendJSON("user_message", userMessage); err != nil {
		log.Printf("Error echoing user message: %v", err)
		return
	}

	// Add user message to history before processing.
	h.sess.State["last_user_message"] = userInputText
	delete(h.sess.State, "last_agent_response_text") // Clear previous agent response
	h.sess.AddMessage(userMessage)

	// Create a context with the UI sender function.
	uiSender := func(messageType string, payload any) {
		_ = h.sendJSON(messageType, payload)
	}
	agentCtx := invocation.WithUISender(ctx, uiSender)

	// Process message with the agent.
	response, err := h.agent.Process(agentCtx, h.sess.GetHistory(), userMessage)
	if err != nil {
		log.Printf("Agent processing error: %v", err)
		errorText := "I encountered an error: " + err.Error()
		errorMsg := modelstypes.Message{Role: "model", Parts: []modelstypes.Part{{Text: &errorText}}}
		_ = h.sendJSON("agent_response", errorMsg)
		return
	}

	if response != nil {
		// Simulate state update for observability
		h.sess.State["last_agent_response_role"] = response.Role
		if len(response.Parts) > 0 {
			if response.Parts[0].Text != nil {
				h.sess.State["last_agent_response_text"] = *response.Parts[0].Text
			}
			if response.Parts[0].FunctionCall != nil {
				h.sess.State["last_agent_tool_call"] = response.Parts[0].FunctionCall
			}
		}
		h.sess.AddMessage(*response)
		if err := h.sendJSON("agent_response", *response); err != nil {
			log.Printf("Error sending agent response: %v", err)
		}
	}

	// Prune and save the session.
	h.sess.PruneHistory(maxHistoryTurns)
	sessions.Save(h.sess)
	// Send state update to client
	_ = h.sendJSON("state_update", h.sess.State)
}
