package handler

import (
	"net/http"
	"strconv"

	"chatapp/internal/middleware"
	"chatapp/internal/service"
	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	messageService *service.MessageService
}

func NewMessageHandler(messageService *service.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

// CreateMessage handles message creation
func (h *MessageHandler) CreateMessage(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req service.CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	message, err := h.messageService.CreateMessage(userID, req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Message created successfully",
		"data":    message,
	})
}

// GetMessages handles message retrieval with pagination
func (h *MessageHandler) GetMessages(c *gin.Context) {
	// Get channel from query parameter (default: general)
	channel := c.DefaultQuery("channel", "general")

	// Get pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "50")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}

	messages, err := h.messageService.GetMessagesByChannel(channel, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve messages",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": messages,
	})
}

// GetRecentMessages handles recent message retrieval
func (h *MessageHandler) GetRecentMessages(c *gin.Context) {
	// Get channel from query parameter (default: general)
	channel := c.DefaultQuery("channel", "general")

	// Get limit parameter
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}

	messages, err := h.messageService.GetRecentMessagesByChannel(channel, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve recent messages",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": messages,
	})
}

// GetChannels handles channel list retrieval
func (h *MessageHandler) GetChannels(c *gin.Context) {
	channels, err := h.messageService.GetAvailableChannels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve channels",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": channels,
	})
}

// GetChannelInfo handles channel information retrieval
func (h *MessageHandler) GetChannelInfo(c *gin.Context) {
	channel := c.Param("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Channel name is required",
		})
		return
	}

	channelInfo, err := h.messageService.GetChannelInfo(channel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve channel information",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": channelInfo,
	})
}

// DeleteMessage handles message deletion
func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	messageIDStr := c.Param("id")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid message ID",
		})
		return
	}

	err = h.messageService.DeleteMessage(uint(messageID), userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "message not found" {
			status = http.StatusNotFound
		} else if err.Error() == "unauthorized: can only delete your own messages" {
			status = http.StatusForbidden
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Message deleted successfully",
	})
}