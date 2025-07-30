package utils

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocketManager manages WebSocket connections
type WebSocketManager struct {
	connections map[string]map[*websocket.Conn]bool // boardID -> connections
	mutex       sync.RWMutex
	upgrader    websocket.Upgrader
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type    string      `json:"type"`
	BoardID string      `json:"boardId,omitempty"`
	IdeaID  string      `json:"ideaId,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// FeedbackAnimation represents feedback animation data
type FeedbackAnimation struct {
	IdeaID       string `json:"ideaId"`
	FeedbackType string `json:"feedbackType"`
	Emoji        string `json:"emoji,omitempty"`
	Timestamp    int64  `json:"timestamp"`
}

var wsManager *WebSocketManager

// InitWebSocketManager initializes the WebSocket manager
func InitWebSocketManager() {
	wsManager = &WebSocketManager{
		connections: make(map[string]map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
		},
	}
}

// HandleWebSocket handles WebSocket connections
func HandleWebSocket(c *gin.Context) {
	boardID := c.Param("boardId")
	if boardID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Board ID required"})
		return
	}

	conn, err := wsManager.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Add connection to manager
	wsManager.addConnection(boardID, conn)
	defer wsManager.removeConnection(boardID, conn)

	log.Printf("WebSocket connected for board: %s", boardID)

	// Handle incoming messages (ping/pong, etc.)
	for {
		var msg WebSocketMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle different message types
		switch msg.Type {
		case "ping":
			conn.WriteJSON(WebSocketMessage{Type: "pong"})
		}
	}
}

// addConnection adds a WebSocket connection for a board
func (wsm *WebSocketManager) addConnection(boardID string, conn *websocket.Conn) {
	wsm.mutex.Lock()
	defer wsm.mutex.Unlock()

	if wsm.connections[boardID] == nil {
		wsm.connections[boardID] = make(map[*websocket.Conn]bool)
	}
	wsm.connections[boardID][conn] = true
}

// removeConnection removes a WebSocket connection
func (wsm *WebSocketManager) removeConnection(boardID string, conn *websocket.Conn) {
	wsm.mutex.Lock()
	defer wsm.mutex.Unlock()

	if wsm.connections[boardID] != nil {
		delete(wsm.connections[boardID], conn)
		if len(wsm.connections[boardID]) == 0 {
			delete(wsm.connections, boardID)
		}
	}
}

// BroadcastToBoard sends a message to all connections for a specific board
func (wsm *WebSocketManager) BroadcastToBoard(boardID string, message WebSocketMessage) {
	wsm.mutex.RLock()
	connections := wsm.connections[boardID]
	wsm.mutex.RUnlock()

	if connections == nil {
		return
	}

	// Create a copy of connections to avoid holding the lock during broadcast
	connList := make([]*websocket.Conn, 0, len(connections))
	for conn := range connections {
		connList = append(connList, conn)
	}

	// Broadcast to all connections
	for _, conn := range connList {
		err := conn.WriteJSON(message)
		if err != nil {
			log.Printf("WebSocket write error: %v", err)
			// Remove failed connection
			wsm.removeConnection(boardID, conn)
			conn.Close()
		}
	}
}

// BroadcastFeedbackAnimation broadcasts feedback animation to admin board
func BroadcastFeedbackAnimation(boardID, ideaID, feedbackType string, emoji string) {
	if wsManager == nil {
		return
	}

	animation := FeedbackAnimation{
		IdeaID:       ideaID,
		FeedbackType: feedbackType,
		Emoji:        emoji,
		Timestamp:    getCurrentTimestamp(),
	}

	message := WebSocketMessage{
		Type:    "feedback_animation",
		BoardID: boardID,
		IdeaID:  ideaID,
		Data:    animation,
	}

	wsManager.BroadcastToBoard(boardID, message)
	log.Printf("Feedback animation broadcasted: Board=%s, Idea=%s, Type=%s",
		boardID, ideaID, feedbackType)
}

// BroadcastIdeaUpdate broadcasts idea updates to all board connections
func BroadcastIdeaUpdate(boardID, ideaID string, updateData interface{}) {
	if wsManager == nil {
		return
	}

	message := WebSocketMessage{
		Type:    "idea_update",
		BoardID: boardID,
		IdeaID:  ideaID,
		Data:    updateData,
	}

	wsManager.BroadcastToBoard(boardID, message)
}

// getCurrentTimestamp returns current timestamp in milliseconds
func getCurrentTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
