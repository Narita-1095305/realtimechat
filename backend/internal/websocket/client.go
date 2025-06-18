package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// Client is a middleman between the websocket connection and the hub
type Client struct {
	// The websocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte

	// The hub
	hub *Hub

	// Client ID
	ID string

	// User information
	UserID   uint
	Username string
	Email    string

	// Gin context for request handling
	ctx *gin.Context
}

// IncomingMessage represents a message received from the client
type IncomingMessage struct {
	Type    string `json:"type"`
	Channel string `json:"channel"`
	Content string `json:"content"`
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, userID uint, username, email string, ctx *gin.Context) *Client {
	return &Client{
		conn:     conn,
		send:     make(chan []byte, 256),
		hub:      hub,
		ID:       uuid.New().String(),
		UserID:   userID,
		Username: username,
		Email:    email,
		ctx:      ctx,
	}
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close(websocket.StatusInternalError, "Connection closed")
	}()

	// Set read deadline and message size limit
	ctx := c.ctx.Request.Context()

	for {
		// Read message from WebSocket
		_, message, err := c.conn.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
				websocket.CloseStatus(err) == websocket.StatusGoingAway {
				log.Printf("WebSocket closed normally for client %s", c.ID)
			} else {
				log.Printf("WebSocket read error for client %s: %v", c.ID, err)
			}
			break
		}

		log.Printf("📨 Received raw WebSocket message from client %s: %s", c.ID, string(message))

		// Parse incoming message
		var incomingMsg IncomingMessage
		if err := json.Unmarshal(message, &incomingMsg); err != nil {
			log.Printf("❌ Error parsing message from client %s: %v", c.ID, err)
			log.Printf("❌ Raw message that failed to parse: %s", string(message))
			continue
		}

		log.Printf("✅ Successfully parsed message from client %s: %+v", c.ID, incomingMsg)

		// Handle different message types
		c.handleMessage(incomingMsg)
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close(websocket.StatusInternalError, "Connection closed")
	}()

	ctx := c.ctx.Request.Context()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// The hub closed the channel
				c.conn.Close(websocket.StatusNormalClosure, "Channel closed")
				return
			}

			// Write message to WebSocket
			if err := c.conn.Write(ctx, websocket.MessageText, message); err != nil {
				log.Printf("WebSocket write error for client %s: %v", c.ID, err)
				return
			}

		case <-ticker.C:
			// Send ping
			if err := c.conn.Ping(ctx); err != nil {
				log.Printf("WebSocket ping error for client %s: %v", c.ID, err)
				return
			}
		}
	}
}

// handleMessage handles incoming messages from the client
func (c *Client) handleMessage(msg IncomingMessage) {
	switch msg.Type {
	case "chat_message":
		c.handleChatMessage(msg)
	case "ping":
		c.handlePing()
	case "get_users":
		c.handleGetUsers()
	default:
		log.Printf("Unknown message type from client %s: %s", c.ID, msg.Type)
	}
}

// handleChatMessage handles chat messages
func (c *Client) handleChatMessage(msg IncomingMessage) {
	log.Printf("🔍 handleChatMessage called by client %s (UserID: %d)", c.ID, c.UserID)
	log.Printf("🔍 Message details: Type=%s, Channel=%s, Content=%s", msg.Type, msg.Channel, msg.Content)

	if msg.Content == "" || msg.Channel == "" {
		log.Printf("❌ Invalid chat message from client %s: empty content or channel", c.ID)
		return
	}

	// Create message for broadcasting
	chatMsg := Message{
		Type:    "chat_message",
		Channel: msg.Channel,
		Data: ChatMessage{
			ID:        uint(time.Now().UnixNano()), // より精密なIDを使用
			Content:   msg.Content,
			Channel:   msg.Channel,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
			User: UserInfo{
				ID:       c.UserID,
				Username: c.Username,
			},
		},
		UserID: c.UserID,
		User: UserInfo{
			ID:       c.UserID,
			Username: c.Username,
		},
	}

	log.Printf("📤 Created chat message for broadcast: %+v", chatMsg)

	// Publish to Redis for distribution to all instances
	if err := c.hub.PublishMessage(msg.Channel, chatMsg); err != nil {
		log.Printf("❌ Error publishing message to Redis: %v", err)
	} else {
		log.Printf("✅ Message published to Redis successfully")
	}
}

// handlePing handles ping messages
func (c *Client) handlePing() {
	pongMsg := Message{
		Type: "pong",
		Data: map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	}

	if data, err := json.Marshal(pongMsg); err == nil {
		select {
		case c.send <- data:
		default:
			close(c.send)
		}
	}
}

// handleGetUsers handles requests for connected users
func (c *Client) handleGetUsers() {
	users := c.hub.GetConnectedUsers()

	usersMsg := Message{
		Type: "users_list",
		Data: map[string]interface{}{
			"users": users,
			"count": len(users),
		},
	}

	if data, err := json.Marshal(usersMsg); err == nil {
		select {
		case c.send <- data:
		default:
			close(c.send)
		}
	}
}

// Run starts the client's read and write pumps
func (c *Client) Run() {
	// Register client with hub
	c.hub.register <- c

	// Start pumps
	go c.writePump()
	c.readPump() // This blocks until connection is closed
}
