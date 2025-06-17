package repo

import (
	"chatapp/internal/database"
	"chatapp/internal/models"
	"gorm.io/gorm"
)

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository() *MessageRepository {
	return &MessageRepository{
		db: database.DB,
	}
}

// Create creates a new message
func (r *MessageRepository) Create(message *models.Message) error {
	return r.db.Create(message).Error
}

// GetByID retrieves a message by ID
func (r *MessageRepository) GetByID(id uint) (*models.Message, error) {
	var message models.Message
	err := r.db.Preload("User").First(&message, id).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// GetByChannel retrieves messages by channel with pagination
func (r *MessageRepository) GetByChannel(channel string, offset, limit int) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.Preload("User").
		Where("channel = ?", channel).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

// GetRecentByChannel retrieves recent messages by channel
func (r *MessageRepository) GetRecentByChannel(channel string, limit int) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.Preload("User").
		Where("channel = ?", channel).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error
	
	// Reverse the slice to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	
	return messages, err
}

// GetByUserID retrieves messages by user ID
func (r *MessageRepository) GetByUserID(userID uint, offset, limit int) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.Preload("User").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

// Update updates a message
func (r *MessageRepository) Update(message *models.Message) error {
	return r.db.Save(message).Error
}

// Delete soft deletes a message
func (r *MessageRepository) Delete(id uint) error {
	return r.db.Delete(&models.Message{}, id).Error
}

// CountByChannel counts messages in a channel
func (r *MessageRepository) CountByChannel(channel string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Message{}).Where("channel = ?", channel).Count(&count).Error
	return count, err
}

// List retrieves all messages with pagination
func (r *MessageRepository) List(offset, limit int) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error
	return messages, err
}