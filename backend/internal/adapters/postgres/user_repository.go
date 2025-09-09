package postgres

import (
	"context"
	"errors"
	"real-time-chat/internal/domain"
	"real-time-chat/internal/services" // For custom error types

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) domain.UserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id, username, email, password_hash, profile_picture_url, is_verified) 
              VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query, user.ID, user.Username, user.Email, user.PasswordHash, user.ProfilePictureURL, user.IsVerified)
	return err
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.findUserByField(ctx, "email", email)
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	return r.findUserByField(ctx, "id", id)
}

func (r *PostgresUserRepository) FindByName(ctx context.Context, name string) (*domain.User, error) {
	return r.findUserByField(ctx, "username", name)
}

func (r *PostgresUserRepository) UpdateVerificationStatus(ctx context.Context, userID string, isVerified bool) error {
	query := `UPDATE users SET is_verified = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, isVerified, userID)
	return err
}

func (r *PostgresUserRepository) UpdateProfilePicture(ctx context.Context, userID, url string) error {
	query := `UPDATE users SET profile_picture_url = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, url, userID)
	return err
}

func (r *PostgresUserRepository) UpdatePassword(ctx context.Context, userID, newPasswordHash string) error {
	query := `UPDATE users SET password_hash = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, newPasswordHash, userID)
	return err
}

func (r *PostgresUserRepository) findUserByField(ctx context.Context, field string, value interface{}) (*domain.User, error) {
	user := &domain.User{}
	query := fmt.Sprintf(`SELECT id, username, email, password_hash, profile_picture_url, is_verified, created_at FROM users WHERE %s = $1`, field)
	err := r.db.QueryRow(ctx, query, value).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.ProfilePictureURL, &user.IsVerified, &user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, services.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}
