package postgres

import (
	"context"
	"errors"
	"real-time-chat/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresGroupRepository struct {
	db *pgxpool.Pool
}

func NewPostgresGroupRepository(db *pgxpool.Pool) domain.GroupRepository {
	return &PostgresGroupRepository{db: db}
}

func (r *PostgresGroupRepository) Create(ctx context.Context, group *domain.Group) error {
	query := `INSERT INTO groups (id, name, slug, owner_id) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, query, group.ID, group.Name, group.Slug, group.OwnerID)
	return err
}

func (r *PostgresGroupRepository) FindByID(ctx context.Context, groupID string) (*domain.Group, error) {
	var group domain.Group
	query := `SELECT id, name, slug, owner_id, created_at FROM groups WHERE id = $1`
	err := r.db.QueryRow(ctx, query, groupID).Scan(&group.ID, &group.Name, &group.Slug, &group.OwnerID, &group.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("group not found")
		}
		return nil, err
	}
	return &group, nil
}

func (r *PostgresGroupRepository) FindBySlug(ctx context.Context, slug string) (*domain.Group, error) {
	var group domain.Group
	query := `SELECT id, name, slug, owner_id, created_at FROM groups WHERE slug = $1`
	err := r.db.QueryRow(ctx, query, slug).Scan(&group.ID, &group.Name, &group.Slug, &group.OwnerID, &group.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("group not found by slug")
		}
		return nil, err
	}
	return &group, nil
}

func (r *PostgresGroupRepository) GetMembers(ctx context.Context, groupID string) ([]*domain.User, error) {
	query := `
		SELECT u.id, u.username, u.profile_picture_url
		FROM users u
		JOIN conversation_participants cp ON u.id = cp.user_id
		WHERE cp.conversation_id = $1`
	rows, err := r.db.Query(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*domain.User
	for rows.Next() {
		var member domain.User
		if err := rows.Scan(&member.ID, &member.Username, &member.ProfilePictureURL); err != nil {
			return nil, err
		}
		members = append(members, &member)
	}
	return members, nil
}

func (r *PostgresGroupRepository) AddMember(ctx context.Context, groupID, userID string) error {
	// This is now handled by ConversationRepository.AddParticipant, as group membership is tied to conversation participants
	return errors.New("use ConversationRepository.AddParticipant for group members")
}

func (r *PostgresGroupRepository) RemoveMember(ctx context.Context, groupID, userID string) error {
	// This is now handled by ConversationRepository.RemoveParticipant
	return errors.New("use ConversationRepository.RemoveParticipant for group members")
}

func (r *PostgresGroupRepository) CountMembers(ctx context.Context, groupID string) (int, error) {
	query := `SELECT COUNT(*) FROM conversation_participants WHERE conversation_id = $1`
	var count int
	err := r.db.QueryRow(ctx, query, groupID).Scan(&count)
	return count, err
}

func (r *PostgresGroupRepository) Delete(ctx context.Context, groupID string) error {
	query := `DELETE FROM groups WHERE id = $1`
	_, err := r.db.Exec(ctx, query, groupID)
	return err
}

func (r *PostgresGroupRepository) UpdateOwner(ctx context.Context, groupID, newOwnerID string) error {
	query := `UPDATE groups SET owner_id = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, newOwnerID, groupID)
	return err
}

func (r *PostgresGroupRepository) GetOldestMember(ctx context.Context, groupID string) (string, error) {
	query := `
		SELECT cp.user_id
		FROM conversation_participants cp
		JOIN users u ON cp.user_id = u.id
		WHERE cp.conversation_id = $1
		ORDER BY u.created_at ASC
		LIMIT 1`
	var userID string
	err := r.db.QueryRow(ctx, query, groupID).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errors.New("no members found in group")
		}
		return "", err
	}
	return userID, nil
}
