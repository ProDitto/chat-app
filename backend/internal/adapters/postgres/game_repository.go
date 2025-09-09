package postgres

import (
	"context"
	"errors"
	"real-time-chat/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresGameRepository struct {
	db *pgxpool.Pool
}

func NewPostgresGameRepository(db *pgxpool.Pool) domain.GameRepository {
	return &PostgresGameRepository{db: db}
}

func (r *PostgresGameRepository) Create(ctx context.Context, game *domain.Game) error {
	query := `
		INSERT INTO games (id, player1_id, player2_id, initiator_id, game_type, status, state, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.Exec(ctx, query,
		game.ID, game.Player1ID, game.Player2ID, game.InitiatorID, game.GameType, game.Status,
		game.State, game.CreatedAt, game.UpdatedAt)
	return err
}

func (r *PostgresGameRepository) FindByID(ctx context.Context, gameID string) (*domain.Game, error) {
	var game domain.Game
	query := `
		SELECT id, player1_id, player2_id, initiator_id, game_type, status, state, created_at, updated_at
		FROM games WHERE id = $1`
	err := r.db.QueryRow(ctx, query, gameID).Scan(
		&game.ID, &game.Player1ID, &game.Player2ID, &game.InitiatorID, &game.GameType,
		&game.Status, &game.State, &game.CreatedAt, &game.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("game not found")
		}
		return nil, err
	}
	return &game, nil
}

func (r *PostgresGameRepository) Update(ctx context.Context, game *domain.Game) error {
	query := `
		UPDATE games
		SET player1_id = $2, player2_id = $3, initiator_id = $4, game_type = $5, status = $6, state = $7, updated_at = $8
		WHERE id = $1`
	_, err := r.db.Exec(ctx, query,
		game.ID, game.Player1ID, game.Player2ID, game.InitiatorID, game.GameType, game.Status,
		game.State, time.Now().UTC()) // Always update 'updated_at'
	return err
}
