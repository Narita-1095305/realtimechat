package main

import (
	"log"
	"os"

	"chatapp/internal/database"
	"chatapp/internal/handler"
	"chatapp/internal/middleware"
	"chatapp/internal/repo"
	"chatapp/internal/service"
	ws "chatapp/internal/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	// データベース接続
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Redis接続
	if err := database.ConnectRedis(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer database.CloseRedis()

	// 環境変数からポートを取得（デフォルト: 8080）
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Ginルーターを初期化
	r := gin.Default()

	// CORS設定（開発用）
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// リポジトリ層の初期化
	userRepo := repo.NewUserRepository()
	messageRepo := repo.NewMessageRepository()

	// サービス層の初期化
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key-here")
	authService := service.NewAuthService(userRepo, jwtSecret)
	messageService := service.NewMessageService(messageRepo, userRepo)

	// WebSocketハブの初期化
	hub := ws.NewHub(database.RedisClient, database.RedisSubscriber, messageService)
	go hub.Run() // バックグラウンドでハブを実行

	// ミドルウェアの初期化
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// ハンドラーの初期化
	authHandler := handler.NewAuthHandler(authService)
	messageHandler := handler.NewMessageHandler(messageService)
	wsHandler := handler.NewWebSocketHandler(hub, authService)

	// ヘルスチェックエンドポイント
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Chat App Backend is running",
		})
	})

	// APIルートグループ
	api := r.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})

		// データベース接続テスト用エンドポイント
		api.GET("/db-status", func(c *gin.Context) {
			sqlDB, err := database.DB.DB()
			if err != nil {
				c.JSON(500, gin.H{
					"status":  "error",
					"message": "Failed to get database instance",
				})
				return
			}

			if err := sqlDB.Ping(); err != nil {
				c.JSON(500, gin.H{
					"status":  "error",
					"message": "Database connection failed",
				})
				return
			}

			c.JSON(200, gin.H{
				"status":  "ok",
				"message": "Database connection is healthy",
			})
		})

		// 認証エンドポイント
		auth := api.Group("/auth")
		{
			auth.POST("/signup", authHandler.Signup)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", authMiddleware.RequireAuth(), authHandler.GetMe)
			auth.POST("/refresh", authMiddleware.RequireAuth(), authHandler.RefreshToken)
		}

		// メッセージエンドポイント
		messages := api.Group("/messages")
		{
			// 認証が必要なエンドポイント
			messages.POST("", authMiddleware.RequireAuth(), messageHandler.CreateMessage)
			messages.DELETE("/:id", authMiddleware.RequireAuth(), messageHandler.DeleteMessage)

			// 認証がオプショナルなエンドポイント（将来的にプライベートチャンネル対応時に認証必須にする）
			messages.GET("", authMiddleware.OptionalAuth(), messageHandler.GetMessages)
			messages.GET("/recent", authMiddleware.OptionalAuth(), messageHandler.GetRecentMessages)
		}

		// チャンネルエンドポイント
		channels := api.Group("/channels")
		{
			channels.GET("", authMiddleware.OptionalAuth(), messageHandler.GetChannels)
			channels.GET("/:channel", authMiddleware.OptionalAuth(), messageHandler.GetChannelInfo)
		}

		// WebSocket関連エンドポイント
		wsGroup := api.Group("/ws")
		{
			wsGroup.GET("/users", wsHandler.GetConnectedUsers)
			wsGroup.GET("/stats", wsHandler.GetHubStats)
			wsGroup.POST("/test-broadcast", wsHandler.TestBroadcast)
		}
	}

	// WebSocketエンドポイント（認証必須）
	r.GET("/ws/chat", authMiddleware.RequireAuth(), wsHandler.HandleWebSocket)

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
