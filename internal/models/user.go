package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserRole string

const (
	UserRoleAdmin  UserRole = "admin"
	UserRoleNormal UserRole = "normal"
)

type User struct {
	BaseModel
	Username     string    `json:"username" gorm:"uniqueIndex;not null;size:50"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null;size:100"`
	Password     string    `json:"-" gorm:"not null"`
	DisplayName  string    `json:"display_name" gorm:"size:100"`
	Role         UserRole  `json:"role" gorm:"default:'normal'"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	LastSeenAt   *time.Time `json:"last_seen_at"`
	
	// Relationships
	Messages        []Message        `json:"messages,omitempty" gorm:"foreignKey:UserID"`
	ChannelMembers  []ChannelMember  `json:"channel_members,omitempty" gorm:"foreignKey:UserID"`
}

func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}