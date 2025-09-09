package postgres

import (
	"context"
	"errors"
	"real-time-chat/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresMessageRepository struct {
	db *pgxpool.Pool
}

func NewPostgresMessageRepository(db *pgxpool.Pool) domain.MessageRepository {
	return &PostgresMessageRepository{db: db}
}

func (r *PostgresMessageRepository) Create(ctx context.Context, message *domain.Message) error {
	query := `INSERT INTO messages (id, conversation_id, sender_id, content, server_timestamp) 
              VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query, message.ID, message.ConversationID, message.SenderID, message.Content, message.ServerTimestamp)
	return err
}

func (r *PostgresMessageRepository) FindByConversationID(ctx context.Context, conversationID string, before time.Time, limit int) ([]*domain.Message, error) {
	query := `
		SELECT m.id, m.conversation_id, m.sender_id, m.content, m.server_timestamp, u.id, u.username, u.profile_picture_url
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.conversation_id = $1 AND m.server_timestamp < $2
		ORDER BY m.server_timestamp DESC
		LIMIT $3`

	rows, err := r.db.Query(ctx, query, conversationID, before, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		var msg domain.Message
		var sender domain.User
		msg.Sender = &sender
		err := rows.Scan(
			&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Content, &msg.ServerTimestamp,
			&sender.ID, &sender.Username, &sender.ProfilePictureURL,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}

	// Reverse slice to return messages in chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (r *PostgresMessageRepository) GetLastMessage(ctx context.Context, conversationID string) (*domain.Message, error) {
	query := `
		SELECT m.id, m.conversation_id, m.sender_id, m.content, m.server_timestamp, u.id, u.username, u.profile_picture_url
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.conversation_id = $1
		ORDER BY m.server_timestamp DESC
		LIMIT 1`

	var msg domain.Message
	var sender domain.User
	msg.Sender = &sender

	err := r.db.QueryRow(ctx, query, conversationID).Scan(
		&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Content, &msg.ServerTimestamp,
		&sender.ID, &sender.Username, &sender.ProfilePictureURL,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No messages found
		}
		return nil, err
	}
	return &msg, nil
}
