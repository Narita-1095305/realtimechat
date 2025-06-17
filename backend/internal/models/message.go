package models

import (
	"time"
	"gorm.io/gorm"
)

// Message represents a chat message
type Message struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Content   string         `gorm:"not null;type:text" json:"content"`
	Channel   string         `gorm:"not null;size:50;index;default:'general'" json:"channel"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	
	// リレーション
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName specifies the table name for Message model
func (Message) TableName() string {
	return "messages"
}