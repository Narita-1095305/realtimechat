package main

import (
	"log"
	"os"

	"chatapp/internal/database"
	"github.com/gin-gonic/gin"
)

func main() {
	// データベース接続
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// 環境変数からポートを取得（デフォルト: 8080）
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Ginルーターを初期化
	r := gin.Default()

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
					"status": "error",
					"message": "Failed to get database instance",
				})
				return
			}
			
			if err := sqlDB.Ping(); err != nil {
				c.JSON(500, gin.H{
					"status": "error",
					"message": "Database connection failed",
				})
				return
			}
			
			c.JSON(200, gin.H{
				"status": "ok",
				"message": "Database connection is healthy",
			})
		})
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}