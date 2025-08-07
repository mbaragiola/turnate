package unit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"turnate/internal/config"
	"turnate/internal/database"
	"turnate/internal/middleware"
	"turnate/internal/models"
)

type MiddlewareTestSuite struct {
	suite.Suite
	db     *gorm.DB
	config *config.Config
	user   *models.User
}

func (suite *MiddlewareTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	
	// Create in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	suite.Require().NoError(err)
	
	suite.db = db
	database.DB = db
	
	// Run migrations
	err = models.AutoMigrate(db)
	suite.Require().NoError(err)
	
	// Create test config
	suite.config = &config.Config{
		JWTSecret: "test-secret-key-for-jwt",
		Port:      "8080",
	}
	
	// Create test user
	user := models.User{
		Username:    "middlewaretest",
		Email:       "middleware@example.com",
		DisplayName: "Middleware Test User",
		Role:        models.UserRoleNormal,
		IsActive:    true,
	}
	user.SetPassword("password123")
	
	err = db.Create(&user).Error
	suite.Require().NoError(err)
	
	suite.user = &user
}

func (suite *MiddlewareTestSuite) TestGenerateJWT() {
	t := suite.T()
	
	token, err := middleware.GenerateJWT(suite.user, suite.config)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	
	// Parse and verify the token
	parsedToken, err := jwt.ParseWithClaims(token, &middleware.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(suite.config.JWTSecret), nil
	})
	
	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)
	
	claims, ok := parsedToken.Claims.(*middleware.Claims)
	assert.True(t, ok)
	assert.Equal(t, suite.user.ID.String(), claims.UserID)
	assert.Equal(t, suite.user.Username, claims.Username)
	assert.Equal(t, string(suite.user.Role), claims.Role)
}

func (suite *MiddlewareTestSuite) TestAuthMiddlewareValid() {
	t := suite.T()
	
	// Generate valid token
	token, err := middleware.GenerateJWT(suite.user, suite.config)
	assert.NoError(t, err)
	
	// Create test router
	r := gin.New()
	r.Use(middleware.AuthMiddleware(suite.config))
	r.GET("/protected", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")
		role, _ := c.Get("role")
		
		c.JSON(http.StatusOK, gin.H{
			"user_id":  userID,
			"username": username,
			"role":     role,
		})
	})
	
	// Make request with valid token
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

func (suite *MiddlewareTestSuite) TestAuthMiddlewareNoToken() {
	t := suite.T()
	
	r := gin.New()
	r.Use(middleware.AuthMiddleware(suite.config))
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	// Make request without token
	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func (suite *MiddlewareTestSuite) TestAuthMiddlewareInvalidToken() {
	t := suite.T()
	
	r := gin.New()
	r.Use(middleware.AuthMiddleware(suite.config))
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	// Make request with invalid token
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func (suite *MiddlewareTestSuite) TestAuthMiddlewareExpiredToken() {
	t := suite.T()
	
	// Create expired token
	claims := &middleware.Claims{
		UserID:   suite.user.ID.String(),
		Username: suite.user.Username,
		Role:     string(suite.user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // Expired
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "turnate",
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(suite.config.JWTSecret))
	assert.NoError(t, err)
	
	r := gin.New()
	r.Use(middleware.AuthMiddleware(suite.config))
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	// Make request with expired token
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func (suite *MiddlewareTestSuite) TestAuthMiddlewareInactiveUser() {
	t := suite.T()
	
	// Create inactive user
	inactiveUser := models.User{
		Username:    "inactive",
		Email:       "inactive@example.com",
		DisplayName: "Inactive User",
		Role:        models.UserRoleNormal,
		IsActive:    false,
	}
	inactiveUser.SetPassword("password123")
	suite.db.Create(&inactiveUser)
	
	// Make sure the user is marked as inactive
	suite.db.Model(&inactiveUser).Update("is_active", false)
	
	// Generate token for inactive user  
	token, err := middleware.GenerateJWT(&inactiveUser, suite.config)
	assert.NoError(t, err)
	
	r := gin.New()
	r.Use(middleware.AuthMiddleware(suite.config))
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	// Make request with token for inactive user
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	// The middleware should reject inactive users
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func (suite *MiddlewareTestSuite) TestAdminMiddleware() {
	t := suite.T()
	
	// Create admin user
	adminUser := models.User{
		Username:    "admin",
		Email:       "admin@example.com",
		DisplayName: "Admin User",
		Role:        models.UserRoleAdmin,
		IsActive:    true,
	}
	adminUser.SetPassword("password123")
	suite.db.Create(&adminUser)
	
	adminToken, err := middleware.GenerateJWT(&adminUser, suite.config)
	assert.NoError(t, err)
	
	normalToken, err := middleware.GenerateJWT(suite.user, suite.config)
	assert.NoError(t, err)
	
	r := gin.New()
	r.Use(middleware.AuthMiddleware(suite.config))
	r.Use(middleware.AdminMiddleware())
	r.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})
	
	// Test with admin token
	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Test with normal user token
	req = httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+normalToken)
	
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func (suite *MiddlewareTestSuite) TestCORSMiddleware() {
	t := suite.T()
	
	r := gin.New()
	r.Use(middleware.CORSMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})
	
	// Test preflight request
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	
	// Test actual request
	req = httptest.NewRequest("GET", "/test", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func (suite *MiddlewareTestSuite) TestSecurityHeaders() {
	t := suite.T()
	
	r := gin.New()
	r.Use(middleware.SecurityHeaders())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
}

func (suite *MiddlewareTestSuite) TestLastSeenUpdate() {
	t := suite.T()
	
	originalLastSeen := suite.user.LastSeenAt
	
	// Generate token and make request
	token, err := middleware.GenerateJWT(suite.user, suite.config)
	assert.NoError(t, err)
	
	r := gin.New()
	r.Use(middleware.AuthMiddleware(suite.config))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Check that last seen was updated
	var updatedUser models.User
	suite.db.Where("id = ?", suite.user.ID).First(&updatedUser)
	
	if originalLastSeen == nil {
		assert.NotNil(t, updatedUser.LastSeenAt)
	} else {
		assert.True(t, updatedUser.LastSeenAt.After(*originalLastSeen))
	}
}

func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}