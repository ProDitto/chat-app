package postgres

import (
	"context"
	"errors"
	"real-time-chat/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConversationRepository struct {
	db *pgxpool.Pool
}

func NewPostgresConversationRepository(db *pgxpool.Pool) domain.ConversationRepository {
	return &PostgresConversationRepository{db: db}
}

func (r *PostgresConversationRepository) Create(ctx context.Context, conversation *domain.Conversation) (string, error) {
	query := `INSERT INTO conversations (id, type) VALUES ($1, $2) RETURNING id`
	var id string
	err := r.db.QueryRow(ctx, query, conversation.ID, conversation.Type).Scan(&id)
	return id, err
}

func (r *PostgresConversationRepository) AddParticipant(ctx context.Context, conversationID, userID string) error {
	query := `INSERT INTO conversation_participants (conversation_id, user_id, last_read_timestamp) VALUES ($1, $2, NOW()) ON CONFLICT (conversation_id, user_id) DO NOTHING`
	_, err := r.db.Exec(ctx, query, conversationID, userID)
	return err
}

func (r *PostgresConversationRepository) RemoveParticipant(ctx context.Context, conversationID, userID string) error {
	query := `DELETE FROM conversation_participants WHERE conversation_id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, query, conversationID, userID)
	return err
}

func (r *PostgresConversationRepository) IsUserInConversation(ctx context.Context, conversationID, userID string) (bool, error) {
	participants, err := r.GetParticipantIDs(ctx, conversationID)
	if err != nil {
		return false, err
	}
	for _, pid := range participants {
		if pid == userID {
			return true, nil
		}
	}
	return false, nil
}

func (r *PostgresConversationRepository) GetParticipantIDs(ctx context.Context, conversationID string) ([]string, error) {
	query := `SELECT user_id FROM conversation_participants WHERE conversation_id = $1`
	rows, err := r.db.Query(ctx, query, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *PostgresConversationRepository) FindForUser(ctx context.Context, userID string) ([]*domain.Conversation, error) {
	query := `
		WITH UserConversations AS (
			SELECT conversation_id, last_read_timestamp FROM conversation_participants WHERE user_id = $1
		),
		RankedMessages AS (
			SELECT m.*, ROW_NUMBER() OVER(PARTITION BY conversation_id ORDER BY server_timestamp DESC) as rn
			FROM messages m
			WHERE m.conversation_id IN (SELECT conversation_id FROM UserConversations)
		),
		LastMessages AS (
			SELECT rm.*, u.username as sender_username, u.profile_picture_url as sender_profile_picture_url FROM RankedMessages rm
			JOIN users u ON rm.sender_id = u.id
			WHERE rm.rn = 1
		)
		SELECT 
			c.id, c.type, c.created_at,
			lm.id, lm.sender_id, lm.content, lm.server_timestamp,
			lm.sender_username, lm.sender_profile_picture_url,
			uc.last_read_timestamp,
			(SELECT COUNT(m_unread.id) FROM messages m_unread 
			 WHERE m_unread.conversation_id = c.id AND m_unread.server_timestamp > uc.last_read_timestamp) as unread_count,
			g.name as group_name, g.slug as group_slug, g.owner_id as group_owner_id, g.created_at as group_created_at
		FROM conversations c
		JOIN UserConversations uc ON c.id = uc.conversation_id
		LEFT JOIN LastMessages lm ON c.id = lm.conversation_id
		LEFT JOIN groups g ON c.id = g.id AND c.type = 'group'
		ORDER BY lm.server_timestamp DESC NULLS LAST;
	` // Adjusted query to correctly get unread count and group details
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []*domain.Conversation
	for rows.Next() {
		var convo domain.Conversation
		var lastMessage domain.Message
		var sender domain.User
		var group domain.Group
		var lastReadTimestamp time.Time

		var lastMessageID, lastMessageSenderID, lastMessageContent, senderUsername, senderProfilePictureURL pgtype.Text
		var lastMessageTimestamp pgtype.Timestamp
		var unreadCount pgtype.Int4
		var groupName, groupSlug, groupOwnerID pgtype.Text
		var groupCreatedAt pgtype.Timestamp

		err := rows.Scan(
			&convo.ID, &convo.Type, &convo.CreatedAt,
			&lastMessageID, &lastMessageSenderID, &lastMessageContent, &lastMessageTimestamp,
			&senderUsername, &senderProfilePictureURL,
			&lastReadTimestamp,
			&unreadCount,
			&groupName, &groupSlug, &groupOwnerID, &groupCreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if lastMessageID.Valid {
			lastMessage.ID = lastMessageID.String
			lastMessage.SenderID = lastMessageSenderID.String
			lastMessage.Content = lastMessageContent.String
			lastMessage.ServerTimestamp = lastMessageTimestamp.Time
			sender.ID = lastMessageSenderID.String
			sender.Username = senderUsername.String
			sender.ProfilePictureURL = senderProfilePictureURL.String
			lastMessage.Sender = &sender
			convo.LastMessage = &lastMessage
		}
		convo.UnreadCount = int(unreadCount.Int32)
		if groupName.Valid {
			group.ID = convo.ID
			group.Name = groupName.String
			group.Slug = groupSlug.String
			group.OwnerID = groupOwnerID.String
			group.CreatedAt = groupCreatedAt.Time
			convo.Group = &group
			convo.Name = group.Name // Set conversation name to group name
		}

		conversations = append(conversations, &convo)
	}

	return conversations, nil
}

func (r *PostgresConversationRepository) FindByID(ctx context.Context, conversationID string) (*domain.Conversation, error) {
	query := `SELECT id, type, created_at FROM conversations WHERE id = $1`
	var convo domain.Conversation
	err := r.db.QueryRow(ctx, query, conversationID).Scan(&convo.ID, &convo.Type, &convo.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("conversation not found")
		}
		return nil, err
	}
	return &convo, nil
}

func (r *PostgresConversationRepository) FindOneToOne(ctx context.Context, userID1, userID2 string) (string, error) {
	query := `
		SELECT cp1.conversation_id
		FROM conversation_participants cp1
		JOIN conversation_participants cp2 ON cp1.conversation_id = cp2.conversation_id
		JOIN conversations c ON cp1.conversation_id = c.id
		WHERE cp1.user_id = $1 AND cp2.user_id = $2 AND c.type = 'one-on-one'
		LIMIT 1
	`
	var conversationID string
	err := r.db.QueryRow(ctx, query, userID1, userID2).Scan(&conversationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil // Not found
		}
		return "", err
	}
	return conversationID, nil
}

func (r *PostgresConversationRepository) UpdateLastRead(ctx context.Context, conversationID, userID string) error {
	query := `
		UPDATE conversation_participants
		SET last_read_timestamp = NOW()
		WHERE conversation_id = $1 AND user_id = $2
	`
	_, err := r.db.Exec(ctx, query, conversationID, userID)
	return err
}

func (r *PostgresConversationRepository) Delete(ctx context.Context, conversationID string) error {
	query := `DELETE FROM conversations WHERE id = $1`
	_, err := r.db.Exec(ctx, query, conversationID)
	return err
}

func (r *PostgresConversationRepository) GetLastReadTimestamp(ctx context.Context, conversationID, userID string) (time.Time, error) {
	query := `SELECT last_read_timestamp FROM conversation_participants WHERE conversation_id = $1 AND user_id = $2`
	var lastRead time.Time
	err := r.db.QueryRow(ctx, query, conversationID, userID).Scan(&lastRead)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, nil // No entry, effectively never read
		}
		return time.Time{}, err
	}
	return lastRead, nil
}
