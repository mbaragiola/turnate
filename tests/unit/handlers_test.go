package unit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"turnate/internal/config"
	"turnate/internal/database"
	"turnate/internal/handlers"
	"turnate/internal/middleware"
	"turnate/internal/models"
)

type HandlersTestSuite struct {
	suite.Suite
	db         *gorm.DB
	router     *gin.Engine
	config     *config.Config
	testUser   *models.User
	testToken  string
}

func (suite *HandlersTestSuite) SetupSuite() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	
	// Create in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	suite.Require().NoError(err)
	
	suite.db = db
	database.DB = db // Set global DB for handlers
	
	// Run migrations
	err = models.AutoMigrate(db)
	suite.Require().NoError(err)
	
	err = models.CreateIndexes(db)
	suite.Require().NoError(err)
	
	// Create test config
	suite.config = &config.Config{
		JWTSecret: "test-secret-key",
		Port:      "8080",
	}
	
	// Setup router
	suite.setupRouter()
	
	// Create test user and token
	suite.createTestUser()
}

func (suite *HandlersTestSuite) setupRouter() {
	r := gin.New()
	
	// Create handlers
	authHandler := handlers.NewAuthHandler(suite.config)
	userHandler := handlers.NewUserHandler()
	channelHandler := handlers.NewChannelHandler()
	messageHandler := handlers.NewMessageHandler()
	
	// Public routes
	r.POST("/api/v1/auth/register", authHandler.Register)
	r.POST("/api/v1/auth/login", authHandler.Login)
	
	// Protected routes
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(suite.config))
	{
		protected.GET("/users/me", authHandler.Profile)
		// User routes
		users := protected.Group("/users")
		{
			users.GET("", userHandler.GetUsers)
			users.PATCH("/:id", userHandler.UpdateUser)
		}
		
		// Channel routes with message sub-routes
		channels := protected.Group("/channels")
		{
			channels.GET("", channelHandler.GetChannels)
			channels.POST("", channelHandler.CreateChannel)
			channels.GET("/:id", channelHandler.GetChannel)
			channels.POST("/:id/join", channelHandler.JoinChannel)
			
			// Message routes under channels
			channels.POST("/:id/messages", messageHandler.CreateMessage)
			channels.GET("/:id/messages", messageHandler.GetMessages)
		}
	}
	
	suite.router = r
}

func (suite *HandlersTestSuite) createTestUser() {
	user := models.User{
		Username:    "testuser",
		Email:       "test@example.com",
		DisplayName: "Test User",
		Role:        models.UserRoleNormal,
		IsActive:    true,
	}
	user.SetPassword("password123")
	
	err := suite.db.Create(&user).Error
	suite.Require().NoError(err)
	
	suite.testUser = &user
	
	// Generate token
	token, err := middleware.GenerateJWT(&user, suite.config)
	suite.Require().NoError(err)
	suite.testToken = token
}

func (suite *HandlersTestSuite) TearDownTest() {
	// Clean up data between tests
	suite.db.Exec("DELETE FROM messages")
	suite.db.Exec("DELETE FROM channel_members") 
	suite.db.Exec("DELETE FROM channels WHERE name != 'general'")
	suite.db.Exec("DELETE FROM users WHERE username != 'testuser'")
}

func (suite *HandlersTestSuite) makeRequest(method, url string, body interface{}, token string) *httptest.ResponseRecorder {
	var bodyReader *bytes.Reader
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(bodyBytes)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}
	
	req := httptest.NewRequest(method, url, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	return w
}

func (suite *HandlersTestSuite) TestUserRegistration() {
	t := suite.T()
	
	registerData := map[string]interface{}{
		"username":     "newuser",
		"email":        "newuser@example.com",
		"password":     "password123",
		"display_name": "New User",
	}
	
	w := suite.makeRequest("POST", "/api/v1/auth/register", registerData, "")
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Contains(t, response, "token")
	assert.Contains(t, response, "user")
	assert.Contains(t, response, "message")
	
	// Verify user was created in database
	var user models.User
	err = suite.db.Where("username = ?", "newuser").First(&user).Error
	assert.NoError(t, err)
	assert.Equal(t, "newuser", user.Username)
	assert.Equal(t, "newuser@example.com", user.Email)
}

func (suite *HandlersTestSuite) TestUserLogin() {
	t := suite.T()
	
	loginData := map[string]interface{}{
		"username": "testuser",
		"password": "password123",
	}
	
	w := suite.makeRequest("POST", "/api/v1/auth/login", loginData, "")
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Contains(t, response, "token")
	assert.Contains(t, response, "user")
	assert.Contains(t, response, "message")
}

func (suite *HandlersTestSuite) TestInvalidLogin() {
	t := suite.T()
	
	loginData := map[string]interface{}{
		"username": "testuser",
		"password": "wrongpassword",
	}
	
	w := suite.makeRequest("POST", "/api/v1/auth/login", loginData, "")
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Contains(t, response, "error")
}

func (suite *HandlersTestSuite) TestGetProfile() {
	t := suite.T()
	
	w := suite.makeRequest("GET", "/api/v1/users/me", nil, suite.testToken)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Contains(t, response, "user")
	user := response["user"].(map[string]interface{})
	assert.Equal(t, "testuser", user["username"])
}

func (suite *HandlersTestSuite) TestUnauthorizedAccess() {
	t := suite.T()
	
	w := suite.makeRequest("GET", "/api/v1/users/me", nil, "")
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func (suite *HandlersTestSuite) TestChannelCreation() {
	t := suite.T()
	
	channelData := map[string]interface{}{
		"name":        "test-channel",
		"description": "A test channel",
		"type":        "public",
	}
	
	w := suite.makeRequest("POST", "/api/v1/channels", channelData, suite.testToken)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Contains(t, response, "channel")
	assert.Contains(t, response, "message")
	
	channel := response["channel"].(map[string]interface{})
	assert.Equal(t, "test-channel", channel["name"])
	assert.Equal(t, "public", channel["type"])
}

func (suite *HandlersTestSuite) TestGetChannels() {
	t := suite.T()
	
	// Create a test channel first
	channel := models.Channel{
		Name:      "get-test",
		Type:      models.ChannelTypePublic,
		CreatedBy: suite.testUser.ID,
	}
	suite.db.Create(&channel)
	
	// Add user as member
	member := models.ChannelMember{
		ChannelID: channel.ID,
		UserID:    suite.testUser.ID,
	}
	suite.db.Create(&member)
	
	w := suite.makeRequest("GET", "/api/v1/channels", nil, suite.testToken)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Contains(t, response, "channels")
	channels := response["channels"].([]interface{})
	assert.GreaterOrEqual(t, len(channels), 1)
}

func (suite *HandlersTestSuite) TestMessageCreation() {
	t := suite.T()
	
	// Create a channel and add user as member
	channel := models.Channel{
		Name:      "msg-test",
		Type:      models.ChannelTypePublic,
		CreatedBy: suite.testUser.ID,
	}
	suite.db.Create(&channel)
	
	member := models.ChannelMember{
		ChannelID: channel.ID,
		UserID:    suite.testUser.ID,
	}
	suite.db.Create(&member)
	
	messageData := map[string]interface{}{
		"content": "Hello, test message!",
	}
	
	url := "/api/v1/channels/" + channel.ID.String() + "/messages"
	w := suite.makeRequest("POST", url, messageData, suite.testToken)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Contains(t, response, "message")
	message := response["message"].(map[string]interface{})
	assert.Equal(t, "Hello, test message!", message["content"])
}

func (suite *HandlersTestSuite) TestGetMessages() {
	t := suite.T()
	
	// Create a channel, add user as member, and create a message
	channel := models.Channel{
		Name:      "getmsg-test",
		Type:      models.ChannelTypePublic,
		CreatedBy: suite.testUser.ID,
	}
	suite.db.Create(&channel)
	
	member := models.ChannelMember{
		ChannelID: channel.ID,
		UserID:    suite.testUser.ID,
	}
	suite.db.Create(&member)
	
	message := models.Message{
		Content:   "Test message to get",
		UserID:    suite.testUser.ID,
		ChannelID: channel.ID,
	}
	suite.db.Create(&message)
	
	url := "/api/v1/channels/" + channel.ID.String() + "/messages"
	w := suite.makeRequest("GET", url, nil, suite.testToken)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Contains(t, response, "messages")
	messages := response["messages"].([]interface{})
	assert.GreaterOrEqual(t, len(messages), 1)
}

func (suite *HandlersTestSuite) TestInvalidChannelAccess() {
	t := suite.T()
	
	// Try to access non-existent channel
	w := suite.makeRequest("GET", "/api/v1/channels/invalid-id/messages", nil, suite.testToken)
	
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func (suite *HandlersTestSuite) TestMessageWithoutMembership() {
	t := suite.T()
	
	// Create a channel without adding user as member
	channel := models.Channel{
		Name:      "no-member-test",
		Type:      models.ChannelTypePublic,
		CreatedBy: suite.testUser.ID,
	}
	suite.db.Create(&channel)
	
	messageData := map[string]interface{}{
		"content": "Should not work",
	}
	
	url := "/api/v1/channels/" + channel.ID.String() + "/messages"
	w := suite.makeRequest("POST", url, messageData, suite.testToken)
	
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func (suite *HandlersTestSuite) TestUpdateUser() {
	t := suite.T()
	
	updateData := map[string]interface{}{
		"display_name": "Updated Name",
	}
	
	url := "/api/v1/users/" + suite.testUser.ID.String()
	w := suite.makeRequest("PATCH", url, updateData, suite.testToken)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Contains(t, response, "user")
	user := response["user"].(map[string]interface{})
	assert.Equal(t, "Updated Name", user["display_name"])
}

func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}