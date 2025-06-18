package service

import (
	"errors"
	"time"

	"chatapp/internal/models"
	"chatapp/internal/repo"
)

type MessageService struct {
	messageRepo *repo.MessageRepository
	userRepo    *repo.UserRepository
}

type CreateMessageRequest struct {
	Content string `json:"content" binding:"required,min=1,max=1000"`
	Channel string `json:"channel" binding:"required,min=1,max=50"`
}

type MessageResponse struct {
	ID        uint      `json:"id"`
	Content   string    `json:"content"`
	Channel   string    `json:"channel"`
	CreatedAt time.Time `json:"created_at"`
	User      UserInfo  `json:"user"`
}

type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

type MessagesListResponse struct {
	Messages []MessageResponse `json:"messages"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	Limit    int               `json:"limit"`
	HasMore  bool              `json:"has_more"`
}

type ChannelInfo struct {
	Name         string `json:"name"`
	MessageCount int64  `json:"message_count"`
	LastMessage  *MessageResponse `json:"last_message,omitempty"`
}

func NewMessageService(messageRepo *repo.MessageRepository, userRepo *repo.UserRepository) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		userRepo:    userRepo,
	}
}

// CreateMessage creates a new message
func (s *MessageService) CreateMessage(userID uint, req CreateMessageRequest) (*MessageResponse, error) {
	// Validate user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Create message
	message := models.Message{
		UserID:  userID,
		Content: req.Content,
		Channel: req.Channel,
	}

	if err := s.messageRepo.Create(&message); err != nil {
		return nil, err
	}

	// Return response with user info
	return &MessageResponse{
		ID:        message.ID,
		Content:   message.Content,
		Channel:   message.Channel,
		CreatedAt: message.CreatedAt,
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
		},
	}, nil
}

// GetMessagesByChannel retrieves messages for a specific channel with pagination
func (s *MessageService) GetMessagesByChannel(channel string, page, limit int) (*MessagesListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	offset := (page - 1) * limit

	// Get messages
	messages, err := s.messageRepo.GetByChannel(channel, offset, limit)
	if err != nil {
		return nil, err
	}

	// Get total count
	total, err := s.messageRepo.CountByChannel(channel)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	messageResponses := make([]MessageResponse, len(messages))
	for i, msg := range messages {
		messageResponses[i] = MessageResponse{
			ID:        msg.ID,
			Content:   msg.Content,
			Channel:   msg.Channel,
			CreatedAt: msg.CreatedAt,
			User: UserInfo{
				ID:       msg.User.ID,
				Username: msg.User.Username,
			},
		}
	}

	// Calculate if there are more messages
	hasMore := int64(offset+limit) < total

	return &MessagesListResponse{
		Messages: messageResponses,
		Total:    total,
		Page:     page,
		Limit:    limit,
		HasMore:  hasMore,
	}, nil
}

// GetRecentMessagesByChannel retrieves recent messages for a channel
func (s *MessageService) GetRecentMessagesByChannel(channel string, limit int) ([]MessageResponse, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}

	messages, err := s.messageRepo.GetRecentByChannel(channel, limit)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	messageResponses := make([]MessageResponse, len(messages))
	for i, msg := range messages {
		messageResponses[i] = MessageResponse{
			ID:        msg.ID,
			Content:   msg.Content,
			Channel:   msg.Channel,
			CreatedAt: msg.CreatedAt,
			User: UserInfo{
				ID:       msg.User.ID,
				Username: msg.User.Username,
			},
		}
	}

	return messageResponses, nil
}

// GetChannelInfo retrieves information about a channel
func (s *MessageService) GetChannelInfo(channel string) (*ChannelInfo, error) {
	// Get message count
	count, err := s.messageRepo.CountByChannel(channel)
	if err != nil {
		return nil, err
	}

	channelInfo := &ChannelInfo{
		Name:         channel,
		MessageCount: count,
	}

	// Get last message if exists
	if count > 0 {
		messages, err := s.messageRepo.GetRecentByChannel(channel, 1)
		if err == nil && len(messages) > 0 {
			msg := messages[0]
			channelInfo.LastMessage = &MessageResponse{
				ID:        msg.ID,
				Content:   msg.Content,
				Channel:   msg.Channel,
				CreatedAt: msg.CreatedAt,
				User: UserInfo{
					ID:       msg.User.ID,
					Username: msg.User.Username,
				},
			}
		}
	}

	return channelInfo, nil
}

// GetAvailableChannels returns list of available channels
func (s *MessageService) GetAvailableChannels() ([]ChannelInfo, error) {
	// For MVP, we only support "general" channel
	// In the future, this could be expanded to support multiple channels
	channels := []string{"general"}
	
	channelInfos := make([]ChannelInfo, len(channels))
	for i, channel := range channels {
		info, err := s.GetChannelInfo(channel)
		if err != nil {
			return nil, err
		}
		channelInfos[i] = *info
	}

	return channelInfos, nil
}

// DeleteMessage deletes a message (only by the author)
func (s *MessageService) DeleteMessage(messageID, userID uint) error {
	// Get message to verify ownership
	message, err := s.messageRepo.GetByID(messageID)
	if err != nil {
		return errors.New("message not found")
	}

	// Check if user is the author
	if message.UserID != userID {
		return errors.New("unauthorized: can only delete your own messages")
	}

	return s.messageRepo.Delete(messageID)
}