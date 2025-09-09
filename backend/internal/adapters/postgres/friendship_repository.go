package postgres

import (
	"context"
	"real-time-chat/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresFriendshipRepository struct {
	db *pgxpool.Pool
}

func NewPostgresFriendshipRepository(db *pgxpool.Pool) domain.FriendshipRepository {
	return &PostgresFriendshipRepository{db: db}
}

func (r *PostgresFriendshipRepository) Create(ctx context.Context, f *domain.Friendship) error {
	query := `INSERT INTO friendships (id, user_id1, user_id2, status) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, query, f.ID, f.UserID1, f.UserID2, f.Status)
	return err
}

func (r *PostgresFriendshipRepository) UpdateStatus(ctx context.Context, requestID string, status domain.FriendshipStatus) (*domain.Friendship, error) {
	var f domain.Friendship
	query := `UPDATE friendships SET status = $1 WHERE id = $2 RETURNING id, user_id1, user_id2, status, created_at`
	err := r.db.QueryRow(ctx, query, status, requestID).Scan(&f.ID, &f.UserID1, &f.UserID2, &f.Status, &f.CreatedAt)
	return &f, err
}

func (r *PostgresFriendshipRepository) GetByID(ctx context.Context, requestID string) (*domain.Friendship, error) {
	var f domain.Friendship
	query := `SELECT id, user_id1, user_id2, status, created_at FROM friendships WHERE id = $1`
	err := r.db.QueryRow(ctx, query, requestID).Scan(&f.ID, &f.UserID1, &f.UserID2, &f.Status, &f.CreatedAt)
	return &f, err
}

func (r *PostgresFriendshipRepository) GetPendingRequests(ctx context.Context, userID string) ([]*domain.FriendshipRequest, error) {
	query := `
		SELECT f.id, f.status, f.created_at, u.id, u.username, u.profile_picture_url
		FROM friendships f
		JOIN users u ON f.user_id1 = u.id
		WHERE f.user_id2 = $1 AND f.status = 'pending'
		ORDER BY f.created_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*domain.FriendshipRequest
	for rows.Next() {
		var req domain.FriendshipRequest
		err := rows.Scan(&req.ID, &req.Status, &req.CreatedAt, &req.Sender.ID, &req.Sender.Username, &req.Sender.ProfilePictureURL)
		if err != nil {
			return nil, err
		}
		requests = append(requests, &req)
	}
	return requests, nil
}

func (r *PostgresFriendshipRepository) GetFriends(ctx context.Context, userID string) ([]*domain.User, error) {
	query := `
		SELECT u.id, u.username, u.profile_picture_url
		FROM users u
		WHERE u.id IN (
			SELECT user_id2 FROM friendships WHERE user_id1 = $1 AND status = 'accepted'
			UNION
			SELECT user_id1 FROM friendships WHERE user_id2 = $1 AND status = 'accepted'
		)
		ORDER BY u.username`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friends []*domain.User
	for rows.Next() {
		var friend domain.User
		err := rows.Scan(&friend.ID, &friend.Username, &friend.ProfilePictureURL)
		if err != nil {
			return nil, err
		}
		friends = append(friends, &friend)
	}
	return friends, nil
}

func (r *PostgresFriendshipRepository) Exists(ctx context.Context, userID1, userID2 string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM friendships 
			WHERE (user_id1 = $1 AND user_id2 = $2) OR (user_id1 = $2 AND user_id2 = $1)
		)`
	err := r.db.QueryRow(ctx, query, userID1, userID2).Scan(&exists)
	return exists, err
}
