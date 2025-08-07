package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"turnate/internal/models"
)

type ModelsTestSuite struct {
	suite.Suite
	db *gorm.DB
}

func (suite *ModelsTestSuite) SetupSuite() {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	suite.Require().NoError(err)
	
	suite.db = db
	
	// Run migrations
	err = models.AutoMigrate(db)
	suite.Require().NoError(err)
	
	err = models.CreateIndexes(db)
	suite.Require().NoError(err)
}

func (suite *ModelsTestSuite) TearDownTest() {
	// Clean up data between tests
	suite.db.Exec("DELETE FROM messages")
	suite.db.Exec("DELETE FROM channel_members")
	suite.db.Exec("DELETE FROM channels")
	suite.db.Exec("DELETE FROM users")
}

func (suite *ModelsTestSuite) TestUserModel() {
	t := suite.T()
	
	// Test creating a user
	user := models.User{
		Username:    "testuser",
		Email:       "test@example.com",
		DisplayName: "Test User",
		Role:        models.UserRoleNormal,
		IsActive:    true,
	}
	
	err := user.SetPassword("password123")
	assert.NoError(t, err)
	
	err = suite.db.Create(&user).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, user.ID)
	
	// Test password verification
	assert.True(t, user.CheckPassword("password123"))
	assert.False(t, user.CheckPassword("wrongpassword"))
	
	// Test password is not stored in plain text
	assert.NotEqual(t, "password123", user.Password)
	
	// Test role methods
	assert.False(t, user.IsAdmin())
	
	user.Role = models.UserRoleAdmin
	assert.True(t, user.IsAdmin())
}

func (suite *ModelsTestSuite) TestUserValidation() {
	t := suite.T()
	
	// Test unique username constraint
	user1 := models.User{
		Username:    "duplicate",
		Email:       "user1@example.com",
		DisplayName: "User 1",
		Role:        models.UserRoleNormal,
	}
	user1.SetPassword("password123")
	
	err := suite.db.Create(&user1).Error
	assert.NoError(t, err)
	
	user2 := models.User{
		Username:    "duplicate",
		Email:       "user2@example.com",
		DisplayName: "User 2",
		Role:        models.UserRoleNormal,
	}
	user2.SetPassword("password123")
	
	err = suite.db.Create(&user2).Error
	assert.Error(t, err)
	
	// Test unique email constraint
	user3 := models.User{
		Username:    "different",
		Email:       "user1@example.com",
		DisplayName: "User 3",
		Role:        models.UserRoleNormal,
	}
	user3.SetPassword("password123")
	
	err = suite.db.Create(&user3).Error
	assert.Error(t, err)
}

func (suite *ModelsTestSuite) TestChannelModel() {
	t := suite.T()
	
	// Create a user first
	user := models.User{
		Username: "channelcreator",
		Email:    "creator@example.com",
		Role:     models.UserRoleNormal,
	}
	user.SetPassword("password123")
	err := suite.db.Create(&user).Error
	assert.NoError(t, err)
	
	// Test creating a channel
	channel := models.Channel{
		Name:        "test-channel",
		Description: "Test channel description",
		Type:        models.ChannelTypePublic,
		CreatedBy:   user.ID,
	}
	
	err = suite.db.Create(&channel).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, channel.ID)
	
	// Test channel types
	assert.Equal(t, models.ChannelTypePublic, channel.Type)
	
	privateChannel := models.Channel{
		Name:      "private-channel",
		Type:      models.ChannelTypePrivate,
		CreatedBy: user.ID,
	}
	
	err = suite.db.Create(&privateChannel).Error
	assert.NoError(t, err)
	assert.Equal(t, models.ChannelTypePrivate, privateChannel.Type)
}

func (suite *ModelsTestSuite) TestChannelMembership() {
	t := suite.T()
	
	// Create users and channel
	user1 := models.User{Username: "user1", Email: "user1@example.com", Role: models.UserRoleNormal}
	user1.SetPassword("password")
	suite.db.Create(&user1)
	
	user2 := models.User{Username: "user2", Email: "user2@example.com", Role: models.UserRoleNormal}
	user2.SetPassword("password")
	suite.db.Create(&user2)
	
	channel := models.Channel{Name: "membership-test", Type: models.ChannelTypePublic, CreatedBy: user1.ID}
	suite.db.Create(&channel)
	
	// Add members
	member1 := models.ChannelMember{ChannelID: channel.ID, UserID: user1.ID}
	err := suite.db.Create(&member1).Error
	assert.NoError(t, err)
	
	member2 := models.ChannelMember{ChannelID: channel.ID, UserID: user2.ID}
	err = suite.db.Create(&member2).Error
	assert.NoError(t, err)
	
	// Test duplicate membership (should fail due to unique constraint)
	duplicateMember := models.ChannelMember{ChannelID: channel.ID, UserID: user1.ID}
	err = suite.db.Create(&duplicateMember).Error
	assert.Error(t, err)
	
	// Count members
	var memberCount int64
	suite.db.Model(&models.ChannelMember{}).Where("channel_id = ?", channel.ID).Count(&memberCount)
	assert.Equal(t, int64(2), memberCount)
}

func (suite *ModelsTestSuite) TestMessageModel() {
	t := suite.T()
	
	// Create user and channel
	user := models.User{Username: "msguser", Email: "msg@example.com", Role: models.UserRoleNormal}
	user.SetPassword("password")
	suite.db.Create(&user)
	
	channel := models.Channel{Name: "msg-channel", Type: models.ChannelTypePublic, CreatedBy: user.ID}
	suite.db.Create(&channel)
	
	// Test creating a message
	message := models.Message{
		Content:   "Hello, world!",
		UserID:    user.ID,
		ChannelID: channel.ID,
	}
	
	err := suite.db.Create(&message).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, message.ID)
	
	// Test thread reply
	reply := models.Message{
		Content:   "Reply to hello",
		UserID:    user.ID,
		ChannelID: channel.ID,
		ThreadID:  &message.ID,
	}
	
	err = suite.db.Create(&reply).Error
	assert.NoError(t, err)
	assert.Equal(t, message.ID, *reply.ThreadID)
	
	// Count replies
	var replyCount int64
	suite.db.Model(&models.Message{}).Where("thread_id = ?", message.ID).Count(&replyCount)
	assert.Equal(t, int64(1), replyCount)
}

func (suite *ModelsTestSuite) TestUUIDv7Generation() {
	t := suite.T()
	
	// Test UUIDv7 generation
	uuid1 := models.NewUUIDv7()
	uuid2 := models.NewUUIDv7()
	
	assert.NotEqual(t, uuid1, uuid2)
	assert.NotEmpty(t, uuid1.String())
	assert.NotEmpty(t, uuid2.String())
	
	// Test UUIDv7 in database
	user := models.User{
		Username: "uuidtest",
		Email:    "uuid@example.com",
		Role:     models.UserRoleNormal,
	}
	user.SetPassword("password")
	
	err := suite.db.Create(&user).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, user.ID.String())
}

func (suite *ModelsTestSuite) TestSoftDelete() {
	t := suite.T()
	
	// Create and delete a user
	user := models.User{
		Username: "deletetest",
		Email:    "delete@example.com",
		Role:     models.UserRoleNormal,
	}
	user.SetPassword("password")
	
	err := suite.db.Create(&user).Error
	assert.NoError(t, err)
	
	// Soft delete
	err = suite.db.Delete(&user).Error
	assert.NoError(t, err)
	
	// Should not find user in normal query
	var foundUser models.User
	err = suite.db.Where("username = ?", "deletetest").First(&foundUser).Error
	assert.Error(t, err)
	
	// Should find user with unscoped query
	err = suite.db.Unscoped().Where("username = ?", "deletetest").First(&foundUser).Error
	assert.NoError(t, err)
	assert.NotNil(t, foundUser.DeletedAt)
}

func TestModelsTestSuite(t *testing.T) {
	suite.Run(t, new(ModelsTestSuite))
}