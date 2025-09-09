package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"real-time-chat/internal/domain"
	"real-time-chat/internal/usecase"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var (
	ErrGroupNotFound     = errors.New("group not found")
	ErrGroupSlugExists   = errors.New("group slug already exists")
	ErrGroupNameTooLong  = errors.New("group name must be at most 20 characters")
	ErrGroupSlugTooLong  = errors.New("group slug must be at most 20 characters")
	ErrInvalidGroupSlug  = errors.New("group slug can only contain lowercase alphabets, numbers, and underscore")
	ErrNotGroupOwner     = errors.New("only group owner can perform this action")
	ErrMemberNotFound    = errors.New("member not found in group")
	ErrAlreadyMember     = errors.New("user is already a member of this group")
	ErrCannotRemoveOwner = errors.New("cannot remove group owner")
	ErrMinGroupMembers   = errors.New("group must have at least one member")
)

var slugRegex = regexp.MustCompile("^[a-z0-9_]+$")

type groupService struct {
	groupRepo    domain.GroupRepository
	userRepo     domain.UserRepository
	convoService usecase.ConversationUseCase
	eventService usecase.EventUseCase
}

func NewGroupService(groupRepo domain.GroupRepository, userRepo domain.UserRepository, convoService usecase.ConversationUseCase, eventService usecase.EventUseCase) usecase.GroupUseCase {
	return &groupService{
		groupRepo:    groupRepo,
		userRepo:     userRepo,
		convoService: convoService,
		eventService: eventService,
	}
}

func (s *groupService) CreateGroup(ctx context.Context, ownerID, name, slug string, initialMembers []string) (*domain.Group, error) {
	if len(name) > 20 {
		return nil, ErrGroupNameTooLong
	}
	if len(slug) > 20 {
		return nil, ErrGroupSlugTooLong
	}
	if !slugRegex.MatchString(slug) {
		return nil, ErrInvalidGroupSlug
	}

	if _, err := s.groupRepo.FindBySlug(ctx, slug); err == nil {
		return nil, ErrGroupSlugExists
	}

	groupID := uuid.NewString()
	group := &domain.Group{
		ID:        groupID,
		Name:      name,
		Slug:      slug,
		OwnerID:   ownerID,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.groupRepo.Create(ctx, group); err != nil {
		return nil, err
	}

	allMembers := []string{ownerID} // Owner is always an initial member
	for _, memberID := range initialMembers {
		if memberID != ownerID {
			allMembers = append(allMembers, memberID)
		}
	}

	if err := s.convoService.CreateGroupConversation(ctx, groupID, allMembers); err != nil {
		return nil, fmt.Errorf("failed to set up group conversation: %w", err)
	}

	// Create event for all members
	eventPayload := map[string]interface{}{
		"group":   group,
		"members": allMembers, // Just IDs for payload
	}
	for _, memberID := range allMembers {
		s.eventService.CreateEvent(ctx, memberID, domain.EventGroupCreated, eventPayload)
	}

	return group, nil
}

func (s *groupService) GetGroupDetails(ctx context.Context, groupID string) (*domain.Group, []*domain.User, error) {
	group, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return nil, nil, ErrGroupNotFound
	}
	members, err := s.groupRepo.GetMembers(ctx, groupID)
	if err != nil {
		return nil, nil, err
	}
	return group, members, nil
}

func (s *groupService) JoinGroup(ctx context.Context, groupID, userID string) error {
	group, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return ErrGroupNotFound
	}

	isMember, err := s.convoService.IsUserInConversation(ctx, groupID, userID)
	if err != nil {
		return err
	}
	if isMember {
		return ErrAlreadyMember
	}

	if err := s.convoService.AddParticipant(ctx, groupID, userID); err != nil {
		return err
	}

	// Create event for the user who joined
	groupJSON, _ := json.Marshal(group)
	s.eventService.CreateEvent(ctx, userID, domain.EventGroupJoined, map[string]json.RawMessage{"group": groupJSON})

	return nil
}

func (s *groupService) LeaveGroup(ctx context.Context, groupID, userID string) error {
	group, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return ErrGroupNotFound
	}

	if group.OwnerID == userID {
		// Owner is leaving
		memberCount, err := s.groupRepo.CountMembers(ctx, groupID)
		if err != nil {
			return err
		}
		if memberCount == 1 { // Owner is the only member
			return s.deleteGroupInternal(ctx, groupID)
		}

		// Transfer ownership to the oldest member
		oldestMemberID, err := s.groupRepo.GetOldestMember(ctx, groupID)
		if err != nil {
			return err // Should not happen if memberCount > 1
		}
		if oldestMemberID == userID {
			// This scenario means the owner is also the oldest member and there's at least one other member.
			// The current logic for GetOldestMember will return the owner if they are the oldest.
			// We need a way to get the *next* oldest, or just pick any other member.
			// For simplicity, we'll assume `GetOldestMember` will return a different member if more exist.
			// A robust solution would need a more sophisticated query for a new owner.
			log.Printf("Owner %s is also the oldest member in group %s. Need better owner transfer logic.", userID, groupID)
			// For now, if oldest is owner, just assign to another arbitrary member (not robust)
			members, err := s.groupRepo.GetMembers(ctx, groupID)
			if err != nil {
				return err
			}
			newOwnerFound := false
			for _, m := range members {
				if m.ID != userID {
					oldestMemberID = m.ID
					newOwnerFound = true
					break
				}
			}
			if !newOwnerFound {
				return errors.New("could not find a new owner for the group")
			}
		}

		if err := s.groupRepo.UpdateOwner(ctx, groupID, oldestMemberID); err != nil {
			return err
		}
		log.Printf("Group %s owner transferred from %s to %s", groupID, userID, oldestMemberID)
	}

	if err := s.convoService.RemoveParticipant(ctx, groupID, userID); err != nil {
		return err
	}

	// Create event for the user who left
	groupJSON, _ := json.Marshal(group)
	s.eventService.CreateEvent(ctx, userID, domain.EventGroupLeft, map[string]json.RawMessage{"group": groupJSON})

	// Check if group should be deleted after member leaves
	memberCount, err := s.groupRepo.CountMembers(ctx, groupID)
	if err != nil {
		return err
	}
	if memberCount == 0 {
		log.Printf("Group %s is empty, deleting...", groupID)
		return s.deleteGroupInternal(ctx, groupID)
	}

	return nil
}

func (s *groupService) RemoveGroupMember(ctx context.Context, performingUserID, groupID, memberToRemoveID string) error {
	group, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return ErrGroupNotFound
	}
	if group.OwnerID != performingUserID {
		return ErrNotGroupOwner
	}
	if memberToRemoveID == performingUserID {
		return ErrCannotRemoveOwner
	}

	isMember, err := s.convoService.IsUserInConversation(ctx, groupID, memberToRemoveID)
	if err != nil {
		return err
	}
	if !isMember {
		return ErrMemberNotFound
	}

	if err := s.convoService.RemoveParticipant(ctx, groupID, memberToRemoveID); err != nil {
		return err
	}

	// Create event for the removed member
	_, err = s.userRepo.FindByID(ctx, memberToRemoveID)
	if err != nil { /* log error */
	}
	groupJSON, _ := json.Marshal(group)
	s.eventService.CreateEvent(ctx, memberToRemoveID, domain.EventGroupLeft, map[string]json.RawMessage{"group": groupJSON, "reason": json.RawMessage(`"removed_by_owner"`)})

	// Check if group should be deleted after member leaves
	memberCount, err := s.groupRepo.CountMembers(ctx, groupID)
	if err != nil {
		return err
	}
	if memberCount == 0 {
		log.Printf("Group %s is empty after member removal, deleting...", groupID)
		return s.deleteGroupInternal(ctx, groupID)
	}

	return nil
}

func (s *groupService) DeleteGroup(ctx context.Context, performingUserID, groupID string) error {
	group, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return ErrGroupNotFound
	}
	if group.OwnerID != performingUserID {
		return ErrNotGroupOwner
	}
	return s.deleteGroupInternal(ctx, groupID)
}

func (s *groupService) deleteGroupInternal(ctx context.Context, groupID string) error {
	// Get all members before deleting the conversation, to send events
	memberIDs, err := s.convoService.GetParticipantIDs(ctx, groupID)
	if err != nil {
		log.Printf("Failed to get members of group %s before deletion: %v", groupID, err)
	}

	// Deleting the conversation will cascade delete participants and messages
	if err := s.convoService.DeleteConversation(ctx, groupID); err != nil {
		return fmt.Errorf("failed to delete group conversation: %w", err)
	}
	if err := s.groupRepo.Delete(ctx, groupID); err != nil {
		return fmt.Errorf("failed to delete group from repository: %w", err)
	}

	// Create events for all members that the group was deleted
	eventPayload := map[string]interface{}{
		"groupId": groupID,
		"message": "Group has been deleted.",
	}
	for _, memberID := range memberIDs {
		s.eventService.CreateEvent(ctx, memberID, domain.EventConversationDeleted, eventPayload)
	}

	return nil
}
