package services

import (
	"context"
	"errors"
	"fmt"
	"real-time-chat/internal/domain"
	"real-time-chat/internal/usecase"

	"github.com/google/uuid"
)

type conversationService struct {
	convoRepo domain.ConversationRepository
	userRepo  domain.UserRepository
	groupRepo domain.GroupRepository
}

func NewConversationService(convoRepo domain.ConversationRepository, userRepo domain.UserRepository, groupRepo domain.GroupRepository) usecase.ConversationUseCase {
	return &conversationService{convoRepo: convoRepo, userRepo: userRepo, groupRepo: groupRepo}
}

func (s *conversationService) GetUserConversations(ctx context.Context, userID string) ([]*domain.Conversation, error) {
	convos, err := s.convoRepo.FindForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Populate conversation names for one-on-one chats and group details
	for _, convo := range convos {
		if convo.Type == domain.TypeOneToOne {
			participantIDs, err := s.convoRepo.GetParticipantIDs(ctx, convo.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get participants for convo %s: %w", convo.ID, err)
			}
			for _, pid := range participantIDs {
				if pid != userID {
					otherUser, err := s.userRepo.FindByID(ctx, pid)
					if err != nil {
						// This could happen if a user was deleted but their conversation entry wasn't cleaned up.
						// For robustness, handle gracefully.
						otherUser = &domain.User{ID: pid, Username: "Unknown User"}
					}
					convo.Name = otherUser.Username
					convo.Participants = []*domain.User{otherUser} // Only the other participant for display
					break
				}
			}
		} else if convo.Type == domain.TypeGroup {
			group, err := s.groupRepo.FindByID(ctx, convo.ID) // Assuming groupID == conversationID for groups
			if err != nil {
				// Similar handling for missing group
				group = &domain.Group{ID: convo.ID, Name: "Unknown Group"}
			}
			convo.Name = group.Name
			convo.Group = group
		}
	}
	return convos, nil
}

func (s *conversationService) AddParticipant(ctx context.Context, conversationID, userID string) error {
	return s.convoRepo.AddParticipant(ctx, conversationID, userID)
}

func (s *conversationService) RemoveParticipant(ctx context.Context, conversationID, userID string) error {
	return s.convoRepo.RemoveParticipant(ctx, conversationID, userID)
}

func (s *conversationService) GetParticipantIDs(ctx context.Context, conversationID string) ([]string, error) {
	return s.convoRepo.GetParticipantIDs(ctx, conversationID)
}

func (s *conversationService) MarkConversationAsRead(ctx context.Context, conversationID, userID string) error {
	return s.convoRepo.UpdateLastRead(ctx, conversationID, userID)
}

func (s *conversationService) CreateOneToOneConversation(ctx context.Context, userID1, userID2 string) (string, error) {
	existingConvoID, err := s.convoRepo.FindOneToOne(ctx, userID1, userID2)
	if err != nil {
		return "", err
	}
	if existingConvoID != "" {
		return existingConvoID, nil // Conversation already exists
	}

	convo := &domain.Conversation{
		ID:   uuid.NewString(),
		Type: domain.TypeOneToOne,
	}
	convoID, err := s.convoRepo.Create(ctx, convo)
	if err != nil {
		return "", err
	}
	if err := s.convoRepo.AddParticipant(ctx, convoID, userID1); err != nil {
		return "", err
	}
	if err := s.convoRepo.AddParticipant(ctx, convoID, userID2); err != nil {
		return "", err
	}
	return convoID, nil
}

func (s *conversationService) DeleteOneToOneConversation(ctx context.Context, conversationID, userID string) error {
	convo, err := s.convoRepo.FindByID(ctx, conversationID)
	if err != nil {
		return errors.New("conversation not found")
	}
	if convo.Type != domain.TypeOneToOne {
		return errors.New("only one-on-one conversations can be deleted this way")
	}

	participantIDs, err := s.convoRepo.GetParticipantIDs(ctx, conversationID)
	if err != nil {
		return err
	}

	isParticipant := false
	for _, pid := range participantIDs {
		if pid == userID {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		return errors.New("user is not a participant of this conversation")
	}

	// Remove the specific user from the conversation. The conversation will persist for the other user.
	return s.convoRepo.RemoveParticipant(ctx, conversationID, userID)
}

func (s *conversationService) GetConversationByID(ctx context.Context, conversationID string) (*domain.Conversation, error) {
	// This fetches basic convo details, without specific user's unread count
	convo, err := s.convoRepo.FindByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	return convo, nil
}

func (s *conversationService) IsUserInConversation(ctx context.Context, conversationID, userID string) (bool, error) {
	return s.convoRepo.IsUserInConversation(ctx, conversationID, userID)
}

func (s *conversationService) CreateGroupConversation(ctx context.Context, groupID string, participantIDs []string) error {
	convo := &domain.Conversation{
		ID:   groupID, // Group ID is also the Conversation ID
		Type: domain.TypeGroup,
	}
	_, err := s.convoRepo.Create(ctx, convo)
	if err != nil {
		return fmt.Errorf("failed to create conversation for group: %w", err)
	}
	for _, userID := range participantIDs {
		if err := s.convoRepo.AddParticipant(ctx, groupID, userID); err != nil {
			return fmt.Errorf("failed to add participant %s to group conversation: %w", userID, err)
		}
	}
	return nil
}

func (s *conversationService) DeleteConversation(ctx context.Context, conversationID string) error {
	return s.convoRepo.Delete(ctx, conversationID)
}
