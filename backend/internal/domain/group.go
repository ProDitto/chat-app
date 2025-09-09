package domain

import (
	"context"
	"time"
)

type Group struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	OwnerID   string    `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

type GroupRepository interface {
	Create(ctx context.Context, group *Group) error
	FindByID(ctx context.Context, groupID string) (*Group, error)
	FindBySlug(ctx context.Context, slug string) (*Group, error)
	GetMembers(ctx context.Context, groupID string) ([]*User, error)
	AddMember(ctx context.Context, groupID, userID string) error
	RemoveMember(ctx context.Context, groupID, userID string) error
	CountMembers(ctx context.Context, groupID string) (int, error)
	Delete(ctx context.Context, groupID string) error
	UpdateOwner(ctx context.Context, groupID, newOwnerID string) error
	GetOldestMember(ctx context.Context, groupID string) (string, error)
}
