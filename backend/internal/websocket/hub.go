package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"chatapp/internal/service"

	"github.com/go-redis/redis/v8"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Redis client for pub/sub
	redisClient *redis.Client

	// Redis subscriber client (separate instance for subscribe operations)
	redisSubscriber *redis.Client

	// Message service for database operations
	messageService *service.MessageService

	// Mutex for thread-safe operations
	mutex sync.RWMutex

	// Context for graceful shutdown
	ctx context.Context
}

// Message represents a WebSocket message
type Message struct {
	Type    string      `json:"type"`
	Channel string      `json:"channel,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	UserID  uint        `json:"user_id,omitempty"`
	User    UserInfo    `json:"user,omitempty"`
}

// UserInfo represents user information in messages
type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

// ChatMessage represents a chat message
type ChatMessage struct {
	ID        uint     `json:"id"`
	Content   string   `json:"content"`
	Channel   string   `json:"channel"`
	CreatedAt string   `json:"created_at"`
	User      UserInfo `json:"user"`
}

// NewHub creates a new WebSocket hub
func NewHub(redisClient *redis.Client, redisSubscriber *redis.Client, messageService *service.MessageService) *Hub {
	return &Hub{
		clients:         make(map[*Client]bool),
		broadcast:       make(chan []byte),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		redisClient:     redisClient,
		redisSubscriber: redisSubscriber,
		messageService:  messageService,
		ctx:             context.Background(),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	// Start Redis subscriber in a separate goroutine
	go h.subscribeToRedis()

	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.clients[client] = true
	log.Printf("Client registered: %s (User ID: %d)", client.ID, client.UserID)

	// Send welcome message
	welcomeMsg := Message{
		Type: "system",
		Data: map[string]interface{}{
			"message": "Connected to chat",
			"user_id": client.UserID,
		},
	}

	if data, err := json.Marshal(welcomeMsg); err == nil {
		select {
		case client.send <- data:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}

	// Notify other clients about user joining
	joinMsg := Message{
		Type:    "user_joined",
		Channel: "general",
		Data: map[string]interface{}{
			"message": client.Username + " joined the chat",
		},
		User: UserInfo{
			ID:       client.UserID,
			Username: client.Username,
		},
	}

	h.publishToRedis("chat:general", joinMsg)
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)
		log.Printf("Client unregistered: %s (User ID: %d)", client.ID, client.UserID)

		// Notify other clients about user leaving
		leaveMsg := Message{
			Type:    "user_left",
			Channel: "general",
			Data: map[string]interface{}{
				"message": client.Username + " left the chat",
			},
			User: UserInfo{
				ID:       client.UserID,
				Username: client.Username,
			},
		}

		h.publishToRedis("chat:general", leaveMsg)
	}
}

// broadcastMessage broadcasts a message to all clients
func (h *Hub) broadcastMessage(message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	log.Printf("📢 Broadcasting message to %d clients: %s", len(h.clients), string(message))

	successCount := 0
	failureCount := 0

	for client := range h.clients {
		select {
		case client.send <- message:
			successCount++
		default:
			log.Printf("❌ Failed to send message to client %s, closing connection", client.ID)
			close(client.send)
			delete(h.clients, client)
			failureCount++
		}
	}

	log.Printf("📢 Broadcast complete - Success: %d, Failed: %d", successCount, failureCount)
}

// PublishMessage publishes a message to Redis for distribution
func (h *Hub) PublishMessage(channel string, msg Message) error {
	return h.publishToRedis("chat:"+channel, msg)
}

// publishToRedis publishes a message to Redis
func (h *Hub) publishToRedis(channel string, msg Message) error {
	log.Printf("📡 Publishing message to Redis channel: %s", channel)
	log.Printf("📡 Message content: %+v", msg)

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("❌ Failed to marshal message for Redis: %v", err)
		return err
	}

	log.Printf("📡 Marshaled message data: %s", string(data))

	err = h.redisClient.Publish(h.ctx, channel, data).Err()
	if err != nil {
		log.Printf("❌ Failed to publish to Redis: %v", err)
	} else {
		log.Printf("✅ Message published to Redis successfully")
	}

	return err
}

// subscribeToRedis subscribes to Redis channels for message distribution
func (h *Hub) subscribeToRedis() {
	log.Printf("🔔 Starting Redis subscription to chat:* channels")

	// Redis接続をテスト (Subscribe用のクライアントを使用)
	ctx := context.Background()
	if err := h.redisSubscriber.Ping(ctx).Err(); err != nil {
		log.Printf("❌ Redis subscriber ping failed: %v", err)
		return
	}
	log.Printf("✅ Redis subscriber ping successful")

	// サブスクリプションを開始 (Subscribe専用クライアントを使用)
	// パターンマッチングにはPSubscribeを使用
	pubsub := h.redisSubscriber.PSubscribe(h.ctx, "chat:*")
	defer pubsub.Close()

	log.Printf("✅ Redis pattern subscription created successfully with subscriber client")

	// サブスクリプションが正常に作成されたかテスト
	if _, err := pubsub.Receive(ctx); err != nil {
		log.Printf("❌ Failed to receive pattern subscription confirmation: %v", err)
		return
	}
	log.Printf("✅ Redis pattern subscription confirmed")

	ch := pubsub.Channel()
	log.Printf("📻 Listening for Redis messages on channel...")

	for msg := range ch {
		log.Printf("📨 Received Redis message on channel %s: %s", msg.Channel, msg.Payload)
		log.Printf("📨 Broadcasting to %d connected clients", len(h.clients))
		h.broadcast <- []byte(msg.Payload)
	}

	log.Printf("⚠️ Redis subscription channel closed")
}

// GetConnectedUsers returns the list of connected users
func (h *Hub) GetConnectedUsers() []UserInfo {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	users := make([]UserInfo, 0, len(h.clients))
	userMap := make(map[uint]bool) // To avoid duplicates

	for client := range h.clients {
		if !userMap[client.UserID] {
			users = append(users, UserInfo{
				ID:       client.UserID,
				Username: client.Username,
			})
			userMap[client.UserID] = true
		}
	}

	return users
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}
