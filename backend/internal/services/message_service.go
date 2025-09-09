package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"real-time-chat/internal/domain"
	"real-time-chat/internal/usecase"
	"time"

	"github.com/google/uuid"
)

type messageService struct {
	messageRepo domain.MessageRepository
	convoRepo   domain.ConversationRepository
	userRepo    domain.UserRepository // Added for fetching sender details
}

func NewMessageService(messageRepo domain.MessageRepository, convoRepo domain.ConversationRepository, userRepo domain.UserRepository) usecase.MessageUseCase {
	return &messageService{messageRepo: messageRepo, convoRepo: convoRepo, userRepo: userRepo}
}

func (s *messageService) SaveMessage(ctx context.Context, message *domain.Message) (*domain.Message, error) {
	if len(message.Content) > 500 {
		return nil, errors.New("message content exceeds 500 characters")
	}

	message.ID = uuid.NewString()
	message.ServerTimestamp = time.Now().UTC()

	err := s.messageRepo.Create(ctx, message)
	if err != nil {
		return nil, err
	}

	// Populate sender details for the returned message
	sender, err := s.userRepo.FindByID(ctx, message.SenderID)
	if err != nil {
		log.Printf("Warning: Could not find sender for message %s: %v", message.ID, err)
		// Proceed without sender details if not found, or return an error depending on strictness
	} else {
		message.Sender = sender
	}

	return message, nil
}

func (s *messageService) GetMessagesForConversation(ctx context.Context, conversationID, userID string, before time.Time, limit int) ([]*domain.Message, error) {
	// Check if user is a participant in the conversation
	isParticipant, err := s.convoRepo.IsUserInConversation(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}
	if !isParticipant {
		return nil, errors.New("user is not a participant of this conversation")
	}

	if before.IsZero() {
		before = time.Now().UTC()
	}
	if limit == 0 {
		limit = 20
	}

	messages, err := s.messageRepo.FindByConversationID(ctx, conversationID, before, limit)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (s *messageService) GetConversationLastMessage(ctx context.Context, conversationID string) (*domain.Message, error) {
	return s.messageRepo.GetLastMessage(ctx, conversationID)
}
