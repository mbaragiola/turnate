package models

import (
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Channel{},
		&ChannelMember{},
		&Message{},
	)
}

func CreateIndexes(db *gorm.DB) error {
	// Create unique index for channel membership
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_channel_members_unique ON channel_members (channel_id, user_id) WHERE deleted_at IS NULL").Error; err != nil {
		return err
	}
	
	// Create indexes for better query performance
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_messages_channel_created ON messages (channel_id, created_at) WHERE deleted_at IS NULL").Error; err != nil {
		return err
	}
	
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_messages_thread_created ON messages (thread_id, created_at) WHERE deleted_at IS NULL AND thread_id IS NOT NULL").Error; err != nil {
		return err
	}
	
	return nil
}