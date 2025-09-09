package domain

import (
	"context"
	"encoding/json"
	"time"
)

type EventType string

const (
	EventNewMessage          EventType = "new_message"
	EventFriendRequest       EventType = "friend_request"
	EventFriendAccepted      EventType = "friend_accepted"
	EventGameInvite          EventType = "game_invite"
	EventGameUpdate          EventType = "game_update"
	EventGroupCreated        EventType = "group_created"
	EventGroupJoined         EventType = "group_joined"
	EventGroupLeft           EventType = "group_left"
	EventConversationDeleted EventType = "conversation_deleted"
)

type Event struct {
	ID              string          `json:"id"`
	UserID          string          `json:"user_id"` // The user this event is primarily relevant to
	EventType       EventType       `json:"event_type"`
	Payload         json.RawMessage `json:"payload"`
	ServerTimestamp time.Time       `json:"server_timestamp"`
}

type EventRepository interface {
	Create(ctx context.Context, event *Event) error
	GetEventsForUser(ctx context.Context, userID string, sinceEventID string, limit int) ([]*Event, error)
	GetEventByID(ctx context.Context, eventID string) (*Event, error)
}
