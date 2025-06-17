package models

import (
	"time"
	"gorm.io/gorm"
)

// User represents a user in the chat application
type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Username  string         `gorm:"uniqueIndex;not null;size:50" json:"username"`
	Email     string         `gorm:"uniqueIndex;not null;size:100" json:"email"`
	Password  string         `gorm:"not null;size:255" json:"-"` // パスワードはJSONに含めない
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	
	// リレーション
	Messages []Message `gorm:"foreignKey:UserID" json:"messages,omitempty"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}