package service

import (
	"errors"
	"fmt"
	"time"

	"chatapp/internal/models"
	"chatapp/internal/repo"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo  *repo.UserRepository
	jwtSecret []byte
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type SignupRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

func NewAuthService(userRepo *repo.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
	}
}

// Signup creates a new user account
func (s *AuthService) Signup(req SignupRequest) (*AuthResponse, error) {
	// Check if email already exists
	exists, err := s.userRepo.EmailExists(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already exists")
	}

	// Check if username already exists
	exists, err = s.userRepo.UsernameExists(req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := s.userRepo.Create(&user); err != nil {
		return nil, err
	}

	// Generate JWT token
	token, err := s.generateToken(&user)
	if err != nil {
		return nil, err
	}

	// Remove password from response
	user.Password = ""

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(req LoginRequest) (*AuthResponse, error) {
	// Find user by email
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	// Remove password from response
	user.Password = ""

	return &AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

// GetUserByID retrieves user information by ID
func (s *AuthService) GetUserByID(userID uint) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Remove password from response
	user.Password = ""
	return user, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	// Debug: Log token parsing attempt
	fmt.Printf("Validating token - length: %d\n", len(tokenString))

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Debug: Check signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		fmt.Printf("JWT parsing error: %v\n", err)
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		fmt.Printf("Token validated successfully - UserID: %d, Username: %s\n", claims.UserID, claims.Username)
		return claims, nil
	}

	fmt.Printf("Token claims validation failed - token.Valid: %v\n", token.Valid)
	return nil, errors.New("invalid token")
}

// generateToken creates a new JWT token for the user
func (s *AuthService) generateToken(user *models.User) (string, error) {
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "chatapp",
			Subject:   "user_auth",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
