package domain

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

type GameStatus string

const (
	GamePending  GameStatus = "pending"
	GameActive   GameStatus = "active"
	GameFinished GameStatus = "finished"
	GameDeclined GameStatus = "declined"
)

// TicTacToeState represents the state of a Tic-Tac-Toe game
type TicTacToeState struct {
	Board       [3][3]string `json:"board"`            // "X", "O", or ""
	CurrentTurn string       `json:"current_turn"`     // Player ID
	Winner      string       `json:"winner,omitempty"` // Player ID or "draw"
	LastMove    *struct {
		Row int `json:"row"`
		Col int `json:"col"`
	} `json:"last_move,omitempty"`
}

type Game struct {
	ID          string          `json:"id"`
	Player1ID   string          `json:"player1_id"`
	Player2ID   string          `json:"player2_id"`
	InitiatorID string          `json:"initiator_id"` // Who sent the invite
	GameType    string          `json:"game_type"`    // "tic-tac-toe"
	Status      GameStatus      `json:"status"`
	State       json.RawMessage `json:"state"` // JSON representation of TicTacToeState
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func (g *Game) GetTicTacToeState() (*TicTacToeState, error) {
	if g.GameType != "tic-tac-toe" || g.State == nil || len(g.State) == 0 {
		return nil, errors.New("not a tic-tac-toe game or state is nil/empty")
	}
	var state TicTacToeState
	if err := json.Unmarshal(g.State, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func (g *Game) SetTicTacToeState(state *TicTacToeState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	g.State = data
	return nil
}

type GameRepository interface {
	Create(ctx context.Context, game *Game) error
	FindByID(ctx context.Context, gameID string) (*Game, error)
	Update(ctx context.Context, game *Game) error
}
