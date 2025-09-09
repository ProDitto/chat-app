package postgres

import (
	"context"
	"errors"
	"real-time-chat/internal/domain"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresEventRepository struct {
	db *pgxpool.Pool
}

func NewPostgresEventRepository(db *pgxpool.Pool) domain.EventRepository {
	return &PostgresEventRepository{db: db}
}

func (r *PostgresEventRepository) Create(ctx context.Context, event *domain.Event) error {
	query := `INSERT INTO events (id, user_id, event_type, payload, server_timestamp) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query, event.ID, event.UserID, event.EventType, event.Payload, event.ServerTimestamp)
	return err
}

func (r *PostgresEventRepository) GetEventsForUser(ctx context.Context, userID string, sinceEventID string, limit int) ([]*domain.Event, error) {
	baseQuery := `SELECT id, user_id, event_type, payload, server_timestamp FROM events WHERE user_id = $1`
	args := []interface{}{userID}
	argCount := 1

	if sinceEventID != "" {
		// First, get the timestamp of the sinceEventID
		var sinceTimestamp string
		timestampQuery := `SELECT server_timestamp FROM events WHERE id = $1`
		err := r.db.QueryRow(ctx, timestampQuery, sinceEventID).Scan(&sinceTimestamp)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, errors.New("since_event_id not found")
			}
			return nil, err
		}
		baseQuery += ` AND server_timestamp > $` + (strconv.Itoa(argCount + 1)) + ` ORDER BY server_timestamp ASC LIMIT $` + (strconv.Itoa(argCount + 2))
		args = append(args, sinceTimestamp, limit)
	} else {
		baseQuery += ` ORDER BY server_timestamp DESC LIMIT $` + (strconv.Itoa(argCount + 1))
		args = append(args, limit)
		// If no sinceEventID, we fetch latest events and then reverse them later.
	}

	rows, err := r.db.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.Event
	for rows.Next() {
		var event domain.Event
		err := rows.Scan(&event.ID, &event.UserID, &event.EventType, &event.Payload, &event.ServerTimestamp)
		if err != nil {
			return nil, err
		}
		events = append(events, &event)
	}

	if sinceEventID == "" {
		// If we fetched the latest events (no sinceEventID), reverse to get chronological order
		for i, j := 0, len(events)-1; i < j; i, j = i+1, j-1 {
			events[i], events[j] = events[j], events[i]
		}
	}

	return events, nil
}

func (r *PostgresEventRepository) GetEventByID(ctx context.Context, eventID string) (*domain.Event, error) {
	query := `SELECT id, user_id, event_type, payload, server_timestamp FROM events WHERE id = $1`
	var event domain.Event
	err := r.db.QueryRow(ctx, query, eventID).Scan(&event.ID, &event.UserID, &event.EventType, &event.Payload, &event.ServerTimestamp)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("event not found")
		}
		return nil, err
	}
	return &event, nil
}
