package domain

import (
	"context"
	"time"
)

type Message struct {
	ID              string    `json:"id"`
	ConversationID  string    `json:"conversation_id"`
	SenderID        string    `json:"sender_id"`
	Content         string    `json:"content"`
	ServerTimestamp time.Time `json:"server_timestamp"`
	Sender          *User     `json:"sender,omitempty"`
}

type MessageRepository interface {
	Create(ctx context.Context, message *Message) error
	FindByConversationID(ctx context.Context, conversationID string, before time.Time, limit int) ([]*Message, error)
	GetLastMessage(ctx context.Context, conversationID string) (*Message, error)
}
