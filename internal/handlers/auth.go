package handlers

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	
	"turnate/internal/config"
	"turnate/internal/database"
	"turnate/internal/middleware"
	"turnate/internal/models"
)

type AuthHandler struct {
	Config *config.Config
}

type RegisterRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=50"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	DisplayName string `json:"display_name,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token   string      `json:"token"`
	User    UserProfile `json:"user"`
	Message string      `json:"message"`
}

type UserProfile struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	IsActive    bool   `json:"is_active"`
}

func NewAuthHandler(config *config.Config) *AuthHandler {
	return &AuthHandler{Config: config}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Validate username format
	if !isValidUsername(req.Username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username must contain only letters, numbers, and underscores"})
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := database.GetDB().Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
		return
	}

	// Create new user (sanitize inputs)
	user := models.User{
		Username:    strings.ToLower(middleware.SanitizeString(req.Username)),
		Email:       strings.ToLower(middleware.SanitizeString(req.Email)),
		DisplayName: middleware.SanitizeString(req.DisplayName),
		Role:        models.UserRoleNormal,
		IsActive:    true,
	}

	if req.DisplayName == "" {
		user.DisplayName = req.Username
	}

	if err := user.SetPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	if err := database.GetDB().Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Add user to general channel
	var generalChannel models.Channel
	if err := database.GetDB().Where("name = ?", "general").First(&generalChannel).Error; err == nil {
		member := models.ChannelMember{
			ChannelID: generalChannel.ID,
			UserID:    user.ID,
		}
		database.GetDB().Create(&member)
	}

	// Generate JWT token
	token, err := middleware.GenerateJWT(&user, h.Config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	response := AuthResponse{
		Token: token,
		User: UserProfile{
			ID:          user.ID.String(),
			Username:    user.Username,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			Role:        string(user.Role),
			IsActive:    user.IsActive,
		},
		Message: "Registration successful! Welcome to Turnate! 🎉",
	}

	c.JSON(http.StatusCreated, response)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	var user models.User
	sanitizedUsername := strings.ToLower(middleware.SanitizeString(req.Username))
	if err := database.GetDB().Where("username = ? OR email = ?", sanitizedUsername, sanitizedUsername).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is disabled"})
		return
	}

	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	token, err := middleware.GenerateJWT(&user, h.Config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	response := AuthResponse{
		Token: token,
		User: UserProfile{
			ID:          user.ID.String(),
			Username:    user.Username,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			Role:        string(user.Role),
			IsActive:    user.IsActive,
		},
		Message: "Login successful! Welcome back! 👋",
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) Profile(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	user := userInterface.(*models.User)
	profile := UserProfile{
		ID:          user.ID.String(),
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Role:        string(user.Role),
		IsActive:    user.IsActive,
	}

	c.JSON(http.StatusOK, gin.H{"user": profile})
}

func isValidUsername(username string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, username)
	return matched
}