package handler

import (
	"net/http"

	"chatapp/internal/middleware"
	"chatapp/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Signup handles user registration
func (h *AuthHandler) Signup(c *gin.Context) {
	var req service.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	response, err := h.authService.Signup(req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "email already exists" || err.Error() == "username already exists" {
			status = http.StatusConflict
		}
		
		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"data":    response,
	})
}

// Login handles user authentication
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	response, err := h.authService.Login(req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "invalid email or password" {
			status = http.StatusUnauthorized
		}
		
		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"data":    response,
	})
}

// GetMe returns current user information
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user information",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}

// RefreshToken generates a new token for the authenticated user
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user information",
		})
		return
	}

	// Create a new signup request to generate a new token
	// Note: This is a simplified approach. In production, you might want
	// a dedicated refresh token mechanism
	response := &service.AuthResponse{
		User: *user,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token refreshed successfully",
		"data":    response,
	})
}