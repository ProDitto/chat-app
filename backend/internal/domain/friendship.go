package domain

import (
	"context"
	"time"
)

type FriendshipStatus string

const (
	Pending  FriendshipStatus = "pending"
	Accepted FriendshipStatus = "accepted"
	Declined FriendshipStatus = "declined"
)

type Friendship struct {
	ID        string           `json:"id"`
	UserID1   string           `json:"user_id1"` // Requester
	UserID2   string           `json:"user_id2"` // Recipient
	Status    FriendshipStatus `json:"status"`
	CreatedAt time.Time        `json:"created_at"`
}

type FriendshipRequest struct {
	ID        string    `json:"id"`
	Sender    User      `json:"sender"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type FriendshipRepository interface {
	Create(ctx context.Context, friendship *Friendship) error
	UpdateStatus(ctx context.Context, requestID string, status FriendshipStatus) (*Friendship, error)
	GetByID(ctx context.Context, requestID string) (*Friendship, error)
	GetPendingRequests(ctx context.Context, userID string) ([]*FriendshipRequest, error)
	GetFriends(ctx context.Context, userID string) ([]*User, error)
	Exists(ctx context.Context, userID1, userID2 string) (bool, error)
}
