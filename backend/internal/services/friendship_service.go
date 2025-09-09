package services

import (
	"context"
	"errors"
	"real-time-chat/internal/domain"
	"real-time-chat/internal/usecase"

	"github.com/google/uuid"
)

type friendshipService struct {
	friendshipRepo domain.FriendshipRepository
	userRepo       domain.UserRepository
	convoService   usecase.ConversationUseCase
	eventService   usecase.EventUseCase // Added EventService
}

func NewFriendshipService(friendshipRepo domain.FriendshipRepository, userRepo domain.UserRepository, convoService usecase.ConversationUseCase, eventService usecase.EventUseCase) usecase.FriendshipUseCase {
	return &friendshipService{friendshipRepo, userRepo, convoService, eventService}
}

func (s *friendshipService) SendRequest(ctx context.Context, requesterID, recipientUsername string) error {
	recipient, err := s.userRepo.FindByName(ctx, recipientUsername)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return errors.New("recipient user not found")
		}
		return err
	}
	if requesterID == recipient.ID {
		return errors.New("cannot send friend request to yourself")
	}

	exists, err := s.friendshipRepo.Exists(ctx, requesterID, recipient.ID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("friend request already exists or you are already friends")
	}

	request := &domain.Friendship{
		ID:      uuid.NewString(),
		UserID1: requesterID,
		UserID2: recipient.ID,
		Status:  domain.Pending,
	}
	if err := s.friendshipRepo.Create(ctx, request); err != nil {
		return err
	}

	// Create an event for the recipient
	requester, err := s.userRepo.FindByID(ctx, requesterID)
	if err != nil {
		return err // Should not happen
	}
	eventPayload := map[string]interface{}{
		"id": request.ID,
		"sender": map[string]string{
			"id":                request.UserID1,
			"username":          requester.Username,
			"profilePictureUrl": requester.ProfilePictureURL,
		},
		"status":    request.Status,
		"createdAt": request.CreatedAt,
	}
	s.eventService.CreateEvent(ctx, recipient.ID, domain.EventFriendRequest, eventPayload)

	return nil
}

func (s *friendshipService) RespondToRequest(ctx context.Context, userID, requestID string, status domain.FriendshipStatus) error {
	request, err := s.friendshipRepo.GetByID(ctx, requestID)
	if err != nil {
		return errors.New("request not found")
	}
	if request.UserID2 != userID {
		return errors.New("not authorized to respond to this request")
	}
	if request.Status != domain.Pending {
		return errors.New("request already handled")
	}

	if status != domain.Accepted && status != domain.Declined {
		return errors.New("invalid status, must be 'accepted' or 'declined'")
	}

	updatedRequest, err := s.friendshipRepo.UpdateStatus(ctx, requestID, status)
	if err != nil {
		return err
	}

	if status == domain.Accepted {
		// Create a one-on-one conversation
		_, err := s.convoService.CreateOneToOneConversation(ctx, updatedRequest.UserID1, updatedRequest.UserID2)
		if err != nil {
			return err
		}
		// Create an event for both users when friend request is accepted
		recipientUser, err := s.userRepo.FindByID(ctx, updatedRequest.UserID2)
		if err != nil { /* handle error */
		}
		requesterUser, err := s.userRepo.FindByID(ctx, updatedRequest.UserID1)
		if err != nil { /* handle error */
		}

		// Payload will be the accepted friendship details
		eventPayload := map[string]interface{}{
			"id":        updatedRequest.ID,
			"user1":     requesterUser,
			"user2":     recipientUser,
			"status":    updatedRequest.Status,
			"createdAt": updatedRequest.CreatedAt,
		}
		s.eventService.CreateEvent(ctx, updatedRequest.UserID1, domain.EventFriendAccepted, eventPayload)
		s.eventService.CreateEvent(ctx, updatedRequest.UserID2, domain.EventFriendAccepted, eventPayload)
	}

	return nil
}

func (s *friendshipService) GetPendingRequests(ctx context.Context, userID string) ([]*domain.FriendshipRequest, error) {
	return s.friendshipRepo.GetPendingRequests(ctx, userID)
}

func (s *friendshipService) GetFriends(ctx context.Context, userID string) ([]*domain.User, error) {
	return s.friendshipRepo.GetFriends(ctx, userID)
}
