package redis

import (
	"context"
	"real-time-chat/internal/domain"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisTokenRepository struct {
	client *redis.Client
}

func NewRedisTokenRepository(client *redis.Client) domain.TokenRepository {
	return &RedisTokenRepository{client: client}
}

func (r *RedisTokenRepository) StoreRefreshToken(ctx context.Context, userID, tokenID string, expiresIn time.Duration) error {
	return r.client.Set(ctx, tokenID, userID, expiresIn).Err()
}

func (r *RedisTokenRepository) GetRefreshTokenUserID(ctx context.Context, tokenID string) (string, error) {
	return r.client.Get(ctx, tokenID).Result()
}

func (r *RedisTokenRepository) DeleteRefreshToken(ctx context.Context, tokenID string) error {
	return r.client.Del(ctx, tokenID).Err()
}

// StoreOTP stores an OTP for a given email with an expiry.
func (r *RedisTokenRepository) StoreOTP(ctx context.Context, email, otp string, expiresIn time.Duration) error {
	otpKey := "otp:" + email
	return r.client.Set(ctx, otpKey, otp, expiresIn).Err()
}

// GetOTP retrieves an OTP for a given email.
func (r *RedisTokenRepository) GetOTP(ctx context.Context, email string) (string, error) {
	otpKey := "otp:" + email
	return r.client.Get(ctx, otpKey).Result()
}

// DeleteOTP deletes an OTP for a given email.
func (r *RedisTokenRepository) DeleteOTP(ctx context.Context, email string) error {
	otpKey := "otp:" + email
	return r.client.Del(ctx, otpKey).Err()
}
