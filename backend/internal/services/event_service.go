package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"real-time-chat/internal/domain"
	"real-time-chat/internal/usecase"
	"time"

	"github.com/google/uuid"
)

type eventService struct {
	eventRepo domain.EventRepository
	userRepo  domain.UserRepository // To enrich user data in events if needed
}

func NewEventService(eventRepo domain.EventRepository, userRepo domain.UserRepository) usecase.EventUseCase {
	return &eventService{eventRepo: eventRepo, userRepo: userRepo}
}

func (s *eventService) CreateEvent(ctx context.Context, userID string, eventType domain.EventType, payload interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	event := &domain.Event{
		ID:              uuid.NewString(),
		UserID:          userID,
		EventType:       eventType,
		Payload:         payloadBytes,
		ServerTimestamp: time.Now().UTC(),
	}

	return s.eventRepo.Create(ctx, event)
}

func (s *eventService) GetEventsForUser(ctx context.Context, userID string, sinceEventID string, limit int) ([]*domain.Event, error) {
	// If sinceEventID is provided, find its timestamp to use as a cursor
	var sinceTimestamp time.Time
	if sinceEventID != "" {
		event, err := s.eventRepo.GetEventByID(ctx, sinceEventID)
		if err != nil {
			if errors.Is(err, errors.New("event not found")) { // Specific error from repository
				return nil, errors.New("invalid since_event_id")
			}
			return nil, err
		}
		sinceTimestamp = event.ServerTimestamp
	} else {
		// If no sinceEventID, get events from the very beginning (or a reasonable default past)
		// For simplicity, if not provided, just get the latest `limit` events.
		// A more robust client would always provide the last known event ID.
		sinceTimestamp = time.Time{} // Zero time means no lower bound on timestamp for initial fetch
	}

	if limit == 0 {
		limit = 50 // Default limit for event feed
	}

	return s.eventRepo.GetEventsForUser(ctx, userID, sinceEventID, limit) // repository will handle timestamp/ID logic
}
