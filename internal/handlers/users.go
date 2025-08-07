package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	
	"turnate/internal/database"
	"turnate/internal/models"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	var users []models.User
	if err := database.GetDB().Select("id, username, display_name, role, is_active, last_seen_at").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	var userProfiles []UserProfile
	for _, user := range users {
		userProfiles = append(userProfiles, UserProfile{
			ID:          user.ID.String(),
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Role:        string(user.Role),
			IsActive:    user.IsActive,
		})
	}

	c.JSON(http.StatusOK, gin.H{"users": userProfiles})
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	userID := c.Param("id")
	
	var user models.User
	if err := database.GetDB().Select("id, username, display_name, role, is_active, last_seen_at").Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	profile := UserProfile{
		ID:          user.ID.String(),
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Role:        string(user.Role),
		IsActive:    user.IsActive,
	}

	c.JSON(http.StatusOK, gin.H{"user": profile})
}

type UpdateUserRequest struct {
	DisplayName *string           `json:"display_name,omitempty"`
	Role        *models.UserRole  `json:"role,omitempty"`
	IsActive    *bool            `json:"is_active,omitempty"`
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	currentUserID, _ := c.Get("user_id")
	currentRole, _ := c.Get("role")
	
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	var user models.User
	if err := database.GetDB().Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Only admins can update role and is_active, users can update their own display_name
	if currentRole != "admin" && userID != currentUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// Non-admins can only update display_name
	if currentRole != "admin" {
		if req.Role != nil || req.IsActive != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can update role and status"})
			return
		}
	}

	// Update fields
	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}
	
	if currentRole == "admin" {
		if req.Role != nil {
			user.Role = *req.Role
		}
		if req.IsActive != nil {
			user.IsActive = *req.IsActive
		}
	}

	if err := database.GetDB().Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	profile := UserProfile{
		ID:          user.ID.String(),
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Role:        string(user.Role),
		IsActive:    user.IsActive,
	}

	c.JSON(http.StatusOK, gin.H{"user": profile, "message": "User updated successfully! ✅"})
}