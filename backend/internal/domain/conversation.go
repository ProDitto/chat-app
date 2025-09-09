package domain

import (
	"context"
	"time"
)

type ConversationType string

// const (
// 	TypeOneToOne     ConversationType = "one-on-one"
// 	TypeGroup        ConversationType = "group"
// )

type Conversation struct {
	ID                string           `json:"id"`
	Type              ConversationType `json:"type"`
	CreatedAt         time.Time        `json:"created_at"`
	Name              string           `json:"name,omitempty"`              // For groups or derived for 1-on-1
	LastMessage       *Message         `json:"last_message,omitempty"`
	Participants      []*User          `json:"participants,omitempty"`
	UnreadCount       int              `json:"unread_count"`
	Group             *Group           `json:"group,omitempty"` // Only for Group conversations
}

type ConversationRepository interface {
	Create(ctx context.Context, conversation *Conversation) (string, error)
	AddParticipant(ctx context.Context, conversationID, userID string) error
	RemoveParticipant(ctx context.Context, conversationID, userID string) error
	GetParticipantIDs(ctx context.Context, conversationID string) ([]string, error)
	FindForUser(ctx context.Context, userID string) ([]*Conversation, error)
	FindByID(ctx context.Context, conversationID string) (*Conversation, error)
	FindOneToOne(ctx context.Context, userID1, userID2 string) (string, error)
	UpdateLastRead(ctx context.Context, conversationID, userID string) error
	Delete(ctx context.Context, conversationID string) error
	GetLastReadTimestamp(ctx context.Context, conversationID, userID string) (time.Time, error)
}
