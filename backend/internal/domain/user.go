package domain

import (
	"context"
	"time"
)

type User struct {
	ID                string    `json:"id"`
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	PasswordHash      string    `json:"-"`
	ProfilePictureURL string    `json:"profile_picture_url"`
	IsVerified        bool      `json:"is_verified"`
	CreatedAt         time.Time `json:"created_at"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	FindByName(ctx context.Context, name string) (*User, error)
	UpdateVerificationStatus(ctx context.Context, userID string, isVerified bool) error
	UpdateProfilePicture(ctx context.Context, userID, url string) error
	UpdatePassword(ctx context.Context, userID, newPasswordHash string) error
	// For simplicity, profile picture and password updates are here.
	// In a larger app, these might be in a separate ProfileRepository.
}

// TokenRepository is for refresh tokens and OTPs
type TokenRepository interface {
	StoreRefreshToken(ctx context.Context, userID, tokenID string, expiresIn time.Duration) error
	GetRefreshTokenUserID(ctx context.Context, tokenID string) (string, error)
	DeleteRefreshToken(ctx context.Context, tokenID string) error
	StoreOTP(ctx context.Context, email, otp string, expiresIn time.Duration) error
	GetOTP(ctx context.Context, email string) (string, error)
	DeleteOTP(ctx context.Context, email string) error
}
