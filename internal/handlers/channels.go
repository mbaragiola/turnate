package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	
	"turnate/internal/database"
	"turnate/internal/middleware"
	"turnate/internal/models"
)

type ChannelHandler struct{}

func NewChannelHandler() *ChannelHandler {
	return &ChannelHandler{}
}

type CreateChannelRequest struct {
	Name        string                `json:"name" binding:"required,min=1,max=100"`
	Description string                `json:"description,omitempty"`
	Type        models.ChannelType    `json:"type,omitempty"`
}

type ChannelResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	CreatedBy   string `json:"created_by"`
	CreatedAt   string `json:"created_at"`
	MemberCount int    `json:"member_count"`
	IsMember    bool   `json:"is_member"`
}

func (h *ChannelHandler) CreateChannel(c *gin.Context) {
	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	// Clean and sanitize channel name
	channelName := middleware.SanitizeString(req.Name)
	channelName = strings.ToLower(strings.TrimSpace(channelName))
	channelName = strings.ReplaceAll(channelName, " ", "-")

	// Check if channel already exists
	var existingChannel models.Channel
	if err := database.GetDB().Where("name = ?", channelName).First(&existingChannel).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Channel already exists"})
		return
	}

	// Only admins can create private channels by default (can be modified)
	channelType := models.ChannelTypePublic
	if req.Type != "" {
		if req.Type == models.ChannelTypePrivate && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can create private channels"})
			return
		}
		channelType = req.Type
	}

	userIDStr := userID.(string)
	channel := models.Channel{
		Name:        channelName,
		Description: middleware.SanitizeString(req.Description),
		Type:        channelType,
		CreatedBy:   models.UUIDv7{},
	}

	// Convert string to UUIDv7
	if err := channel.CreatedBy.Scan(userIDStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := database.GetDB().Create(&channel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create channel"})
		return
	}

	// Add creator as member
	member := models.ChannelMember{
		ChannelID: channel.ID,
		UserID:    channel.CreatedBy,
	}
	database.GetDB().Create(&member)

	response := ChannelResponse{
		ID:          channel.ID.String(),
		Name:        channel.Name,
		Description: channel.Description,
		Type:        string(channel.Type),
		CreatedBy:   channel.CreatedBy.String(),
		CreatedAt:   channel.CreatedAt.Format("2006-01-02T15:04:05Z"),
		MemberCount: 1,
		IsMember:    true,
	}

	c.JSON(http.StatusCreated, gin.H{"channel": response, "message": "Channel created successfully! 🎉"})
}

func (h *ChannelHandler) GetChannels(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var channels []models.Channel
	query := database.GetDB()

	// Regular users only see public channels and private channels they're members of
	if role != "admin" {
		query = query.Where("type = ? OR id IN (SELECT channel_id FROM channel_members WHERE user_id = ? AND deleted_at IS NULL)", 
			models.ChannelTypePublic, userID)
	}

	if err := query.Find(&channels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch channels"})
		return
	}

	var channelResponses []ChannelResponse
	for _, channel := range channels {
		// Count members
		var memberCount int64
		database.GetDB().Model(&models.ChannelMember{}).Where("channel_id = ?", channel.ID).Count(&memberCount)

		// Check if user is member
		var membership models.ChannelMember
		isMember := database.GetDB().Where("channel_id = ? AND user_id = ?", channel.ID, userID).First(&membership).Error == nil

		channelResponses = append(channelResponses, ChannelResponse{
			ID:          channel.ID.String(),
			Name:        channel.Name,
			Description: channel.Description,
			Type:        string(channel.Type),
			CreatedBy:   channel.CreatedBy.String(),
			CreatedAt:   channel.CreatedAt.Format("2006-01-02T15:04:05Z"),
			MemberCount: int(memberCount),
			IsMember:    isMember,
		})
	}

	c.JSON(http.StatusOK, gin.H{"channels": channelResponses})
}

func (h *ChannelHandler) GetChannel(c *gin.Context) {
	channelID := c.Param("id")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var channel models.Channel
	if err := database.GetDB().Where("id = ?", channelID).First(&channel).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	// Check permissions for private channels
	if channel.Type == models.ChannelTypePrivate && role != "admin" {
		var membership models.ChannelMember
		if err := database.GetDB().Where("channel_id = ? AND user_id = ?", channelID, userID).First(&membership).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to private channel"})
			return
		}
	}

	// Count members
	var memberCount int64
	database.GetDB().Model(&models.ChannelMember{}).Where("channel_id = ?", channel.ID).Count(&memberCount)

	// Check if user is member
	var membership models.ChannelMember
	isMember := database.GetDB().Where("channel_id = ? AND user_id = ?", channel.ID, userID).First(&membership).Error == nil

	response := ChannelResponse{
		ID:          channel.ID.String(),
		Name:        channel.Name,
		Description: channel.Description,
		Type:        string(channel.Type),
		CreatedBy:   channel.CreatedBy.String(),
		CreatedAt:   channel.CreatedAt.Format("2006-01-02T15:04:05Z"),
		MemberCount: int(memberCount),
		IsMember:    isMember,
	}

	c.JSON(http.StatusOK, gin.H{"channel": response})
}

func (h *ChannelHandler) JoinChannel(c *gin.Context) {
	channelID := c.Param("id")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var channel models.Channel
	if err := database.GetDB().Where("id = ?", channelID).First(&channel).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	// Check if it's a private channel and user is not admin
	if channel.Type == models.ChannelTypePrivate && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot join private channel"})
		return
	}

	// Check if already a member
	var existingMembership models.ChannelMember
	if err := database.GetDB().Where("channel_id = ? AND user_id = ?", channelID, userID).First(&existingMembership).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Already a member of this channel"})
		return
	}

	// Create membership
	var userUUID models.UUIDv7
	if err := userUUID.Scan(userID.(string)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	member := models.ChannelMember{
		ChannelID: channel.ID,
		UserID:    userUUID,
	}

	if err := database.GetDB().Create(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join channel"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully joined channel! 🎉"})
}

func (h *ChannelHandler) LeaveChannel(c *gin.Context) {
	channelID := c.Param("id")
	userID, _ := c.Get("user_id")

	// Check if member
	var membership models.ChannelMember
	if err := database.GetDB().Where("channel_id = ? AND user_id = ?", channelID, userID).First(&membership).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not a member of this channel"})
		return
	}

	// Don't allow leaving general channel
	var channel models.Channel
	if database.GetDB().Where("id = ? AND name = ?", channelID, "general").First(&channel).Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot leave the general channel"})
		return
	}

	if err := database.GetDB().Delete(&membership).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to leave channel"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully left channel! 👋"})
}

func (h *ChannelHandler) GetChannelMembers(c *gin.Context) {
	channelID := c.Param("id")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var channel models.Channel
	if err := database.GetDB().Where("id = ?", channelID).First(&channel).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	// Check permissions for private channels
	if channel.Type == models.ChannelTypePrivate && role != "admin" {
		var membership models.ChannelMember
		if err := database.GetDB().Where("channel_id = ? AND user_id = ?", channelID, userID).First(&membership).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to private channel"})
			return
		}
	}

	var members []models.User
	if err := database.GetDB().
		Joins("JOIN channel_members ON users.id = channel_members.user_id").
		Where("channel_members.channel_id = ? AND channel_members.deleted_at IS NULL", channelID).
		Select("users.id, users.username, users.display_name, users.role, users.is_active, users.last_seen_at").
		Find(&members).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch channel members"})
		return
	}

	var memberProfiles []UserProfile
	for _, member := range members {
		memberProfiles = append(memberProfiles, UserProfile{
			ID:          member.ID.String(),
			Username:    member.Username,
			DisplayName: member.DisplayName,
			Role:        string(member.Role),
			IsActive:    member.IsActive,
		})
	}

	c.JSON(http.StatusOK, gin.H{"members": memberProfiles})
}