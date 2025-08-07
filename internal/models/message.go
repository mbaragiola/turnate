package models

type Message struct {
	BaseModel
	Content   string  `json:"content" gorm:"not null;type:text"`
	UserID    UUIDv7  `json:"user_id" gorm:"type:text;not null"`
	ChannelID UUIDv7  `json:"channel_id" gorm:"type:text;not null"`
	ThreadID  *UUIDv7 `json:"thread_id,omitempty" gorm:"type:text;index"`
	
	// Relationships
	User     User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Channel  Channel   `json:"channel,omitempty" gorm:"foreignKey:ChannelID"`
	Thread   *Message  `json:"thread,omitempty" gorm:"foreignKey:ThreadID"`
	Replies  []Message `json:"replies,omitempty" gorm:"foreignKey:ThreadID"`
}

type MessageResponse struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	DisplayName string `json:"display_name"`
	ChannelID string `json:"channel_id"`
	ThreadID  *string `json:"thread_id,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	ReplyCount int   `json:"reply_count,omitempty"`
}