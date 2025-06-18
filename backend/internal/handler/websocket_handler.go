package handler

import (
	"net/http"

	"chatapp/internal/middleware"
	"chatapp/internal/service"
	ws "chatapp/internal/websocket"

	"github.com/gin-gonic/gin"
	"nhooyr.io/websocket"
)

type WebSocketHandler struct {
	hub         *ws.Hub
	authService *service.AuthService
}

func NewWebSocketHandler(hub *ws.Hub, authService *service.AuthService) *WebSocketHandler {
	return &WebSocketHandler{
		hub:         hub,
		authService: authService,
	}
}

// HandleWebSocket handles WebSocket connections
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Get user information from context (set by auth middleware)
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required for WebSocket connection",
		})
		return
	}

	username, _ := middleware.GetUsername(c)
	email, _ := middleware.GetEmail(c)

	// Upgrade HTTP connection to WebSocket
	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"}, // Allow all origins for development
		Subprotocols:   []string{"chat"},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to upgrade to WebSocket",
		})
		return
	}

	// Create new client
	client := ws.NewClient(h.hub, conn, userID, username, email, c)

	// Run client (this will block until connection is closed)
	client.Run()
}

// GetConnectedUsers returns the list of connected users
func (h *WebSocketHandler) GetConnectedUsers(c *gin.Context) {
	users := h.hub.GetConnectedUsers()
	count := h.hub.GetClientCount()

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"users": users,
			"count": count,
		},
	})
}

// GetHubStats returns hub statistics
func (h *WebSocketHandler) GetHubStats(c *gin.Context) {
	count := h.hub.GetClientCount()

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"connected_clients": count,
			"status":            "running",
		},
	})
}

// TestBroadcast sends a test message to all connected clients for debugging
func (h *WebSocketHandler) TestBroadcast(c *gin.Context) {
	testMsg := ws.Message{
		Type:    "system",
		Channel: "general",
		Data: map[string]interface{}{
			"message": "Test broadcast message",
			"debug":   true,
		},
	}

	if err := h.hub.PublishMessage("general", testMsg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to send test message",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Test message sent successfully",
		"data":    testMsg,
	})
}
