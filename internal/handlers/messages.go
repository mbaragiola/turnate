package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	
	"turnate/internal/database"
	"turnate/internal/middleware"
	"turnate/internal/models"
)

type MessageHandler struct{}

func NewMessageHandler() *MessageHandler {
	return &MessageHandler{}
}

type CreateMessageRequest struct {
	Content  string  `json:"content" binding:"required,min=1,max=2000"`
	ThreadID *string `json:"thread_id,omitempty"`
}

func (h *MessageHandler) CreateMessage(c *gin.Context) {
	channelID := c.Param("id")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var req CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Verify channel exists and user has access
	var channel models.Channel
	if err := database.GetDB().Where("id = ?", channelID).First(&channel).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	// Check channel membership for private channels
	if channel.Type == models.ChannelTypePrivate && role != "admin" {
		var membership models.ChannelMember
		if err := database.GetDB().Where("channel_id = ? AND user_id = ?", channelID, userID).First(&membership).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to private channel"})
			return
		}
	}

	// For public channels, check if user is a member
	if channel.Type == models.ChannelTypePublic {
		var membership models.ChannelMember
		if err := database.GetDB().Where("channel_id = ? AND user_id = ?", channelID, userID).First(&membership).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Must join channel to post messages"})
			return
		}
	}

	// Convert user ID and channel ID to UUIDv7
	var userUUID, channelUUID models.UUIDv7
	if err := userUUID.Scan(userID.(string)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	if err := channelUUID.Scan(channelID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	message := models.Message{
		Content:   middleware.SanitizeString(req.Content),
		UserID:    userUUID,
		ChannelID: channelUUID,
	}

	// Handle threading
	if req.ThreadID != nil && *req.ThreadID != "" {
		// Verify thread message exists and belongs to same channel
		var threadMessage models.Message
		if err := database.GetDB().Where("id = ? AND channel_id = ?", *req.ThreadID, channelID).First(&threadMessage).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread message"})
			return
		}

		var threadUUID models.UUIDv7
		if err := threadUUID.Scan(*req.ThreadID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
			return
		}
		message.ThreadID = &threadUUID
	}

	if err := database.GetDB().Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
		return
	}

	// Load user data for response
	var user models.User
	database.GetDB().Where("id = ?", userID).First(&user)

	// Count replies if this is a thread
	replyCount := 0
	if message.ThreadID == nil {
		var count int64
		database.GetDB().Model(&models.Message{}).Where("thread_id = ?", message.ID).Count(&count)
		replyCount = int(count)
	}

	response := models.MessageResponse{
		ID:          message.ID.String(),
		Content:     message.Content,
		UserID:      user.ID.String(),
		Username:    user.Username,
		DisplayName: user.DisplayName,
		ChannelID:   message.ChannelID.String(),
		CreatedAt:   message.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   message.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		ReplyCount:  replyCount,
	}

	if message.ThreadID != nil {
		threadIDStr := message.ThreadID.String()
		response.ThreadID = &threadIDStr
	}

	c.JSON(http.StatusCreated, gin.H{"message": response})
}

func (h *MessageHandler) GetMessages(c *gin.Context) {
	channelID := c.Param("id")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 50
	}
	
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Verify channel exists and user has access
	var channel models.Channel
	if err := database.GetDB().Where("id = ?", channelID).First(&channel).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	// Check permissions
	if channel.Type == models.ChannelTypePrivate && role != "admin" {
		var membership models.ChannelMember
		if err := database.GetDB().Where("channel_id = ? AND user_id = ?", channelID, userID).First(&membership).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to private channel"})
			return
		}
	}

	if channel.Type == models.ChannelTypePublic {
		var membership models.ChannelMember
		if err := database.GetDB().Where("channel_id = ? AND user_id = ?", channelID, userID).First(&membership).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Must join channel to view messages"})
			return
		}
	}

	// Get messages (only top-level messages, not replies)
	var messages []models.Message
	if err := database.GetDB().
		Preload("User").
		Where("channel_id = ? AND thread_id IS NULL", channelID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	var messageResponses []models.MessageResponse
	for _, message := range messages {
		// Count replies
		var replyCount int64
		database.GetDB().Model(&models.Message{}).Where("thread_id = ?", message.ID).Count(&replyCount)

		response := models.MessageResponse{
			ID:          message.ID.String(),
			Content:     message.Content,
			UserID:      message.User.ID.String(),
			Username:    message.User.Username,
			DisplayName: message.User.DisplayName,
			ChannelID:   message.ChannelID.String(),
			CreatedAt:   message.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   message.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			ReplyCount:  int(replyCount),
		}

		messageResponses = append(messageResponses, response)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messageResponses)-1; i < j; i, j = i+1, j-1 {
		messageResponses[i], messageResponses[j] = messageResponses[j], messageResponses[i]
	}

	c.JSON(http.StatusOK, gin.H{"messages": messageResponses})
}

func (h *MessageHandler) GetThreadMessages(c *gin.Context) {
	channelID := c.Param("id")
	threadID := c.Param("threadId")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 50
	}
	
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Verify channel and thread
	var channel models.Channel
	if err := database.GetDB().Where("id = ?", channelID).First(&channel).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	var threadMessage models.Message
	if err := database.GetDB().Where("id = ? AND channel_id = ?", threadID, channelID).First(&threadMessage).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
		return
	}

	// Check permissions
	if channel.Type == models.ChannelTypePrivate && role != "admin" {
		var membership models.ChannelMember
		if err := database.GetDB().Where("channel_id = ? AND user_id = ?", channelID, userID).First(&membership).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to private channel"})
			return
		}
	}

	// Get thread replies
	var replies []models.Message
	if err := database.GetDB().
		Preload("User").
		Where("thread_id = ?", threadID).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&replies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch thread replies"})
		return
	}

	var replyResponses []models.MessageResponse
	for _, reply := range replies {
		threadIDStr := threadID
		response := models.MessageResponse{
			ID:          reply.ID.String(),
			Content:     reply.Content,
			UserID:      reply.User.ID.String(),
			Username:    reply.User.Username,
			DisplayName: reply.User.DisplayName,
			ChannelID:   reply.ChannelID.String(),
			ThreadID:    &threadIDStr,
			CreatedAt:   reply.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   reply.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		replyResponses = append(replyResponses, response)
	}

	c.JSON(http.StatusOK, gin.H{"replies": replyResponses})
}

func (h *MessageHandler) GetRecentMessages(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	// Get channels user is member of
	var channelIDs []string
	database.GetDB().Model(&models.ChannelMember{}).
		Where("user_id = ?", userID).
		Pluck("channel_id", &channelIDs)

	if len(channelIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"messages": []models.MessageResponse{}})
		return
	}

	// Get recent messages from user's channels
	var messages []models.Message
	if err := database.GetDB().
		Preload("User").
		Preload("Channel").
		Where("channel_id IN ? AND thread_id IS NULL", channelIDs).
		Where("created_at > ?", time.Now().Add(-24*time.Hour)).
		Order("created_at DESC").
		Limit(20).
		Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recent messages"})
		return
	}

	var messageResponses []models.MessageResponse
	for _, message := range messages {
		// Count replies
		var replyCount int64
		database.GetDB().Model(&models.Message{}).Where("thread_id = ?", message.ID).Count(&replyCount)

		response := models.MessageResponse{
			ID:          message.ID.String(),
			Content:     message.Content,
			UserID:      message.User.ID.String(),
			Username:    message.User.Username,
			DisplayName: message.User.DisplayName,
			ChannelID:   message.ChannelID.String(),
			CreatedAt:   message.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   message.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			ReplyCount:  int(replyCount),
		}

		messageResponses = append(messageResponses, response)
	}

	c.JSON(http.StatusOK, gin.H{"messages": messageResponses})
}