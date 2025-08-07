package models

type ChannelType string

const (
	ChannelTypePublic  ChannelType = "public"
	ChannelTypePrivate ChannelType = "private"
)

type Channel struct {
	BaseModel
	Name        string      `json:"name" gorm:"not null;size:100"`
	Description string      `json:"description" gorm:"size:500"`
	Type        ChannelType `json:"type" gorm:"default:'public'"`
	CreatedBy   UUIDv7      `json:"created_by" gorm:"type:text;not null"`
	
	// Relationships
	Creator     User            `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
	Members     []ChannelMember `json:"members,omitempty" gorm:"foreignKey:ChannelID"`
	Messages    []Message       `json:"messages,omitempty" gorm:"foreignKey:ChannelID"`
}

type ChannelMember struct {
	BaseModel
	ChannelID UUIDv7 `json:"channel_id" gorm:"type:text;not null"`
	UserID    UUIDv7 `json:"user_id" gorm:"type:text;not null"`
	
	// Relationships
	Channel User `json:"channel,omitempty" gorm:"foreignKey:ChannelID"`
	User    User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// Ensure unique channel membership
func (ChannelMember) TableName() string {
	return "channel_members"
}