package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"real-time-chat/internal/domain"
	"real-time-chat/internal/usecase"
	"time"

	"github.com/google/uuid"
)

var (
	ErrGameNotFound         = errors.New("game not found")
	ErrNotYourTurn          = errors.New("it's not your turn")
	ErrInvalidMove          = errors.New("invalid move")
	ErrGameAlreadyFinished  = errors.New("game is already finished")
	ErrGameAlreadyStarted   = errors.New("game already started")
	ErrAlreadyInvited       = errors.New("player already invited to a game")
	ErrNotGameParticipant   = errors.New("not a participant of this game")
	ErrGameAlreadyResponded = errors.New("game invitation already responded to")
)

type gameService struct {
	gameRepo     domain.GameRepository
	userRepo     domain.UserRepository
	convoService usecase.ConversationUseCase // To get participants for broadcasting
	eventService usecase.EventUseCase        // For publishing game events
}

func NewGameService(gameRepo domain.GameRepository, userRepo domain.UserRepository, convoService usecase.ConversationUseCase, eventService usecase.EventUseCase) usecase.GameUseCase {
	return &gameService{gameRepo: gameRepo, userRepo: userRepo, convoService: convoService, eventService: eventService}
}

func (s *gameService) InviteToTicTacToe(ctx context.Context, player1ID, player2Username string) (*domain.Game, error) {
	player2, err := s.userRepo.FindByName(ctx, player2Username)
	if err != nil {
		return nil, ErrUserNotFound
	}
	if player1ID == player2.ID {
		return nil, errors.New("cannot invite yourself to a game")
	}

	// For simplicity, check for existing pending game for these two players
	// A more robust solution might query for games where status is pending and participants are player1ID and player2.ID
	// Skipping this complex query for now, assuming unique game invites per two players for MVP.

	// Randomly decide who goes first
	rand.Seed(time.Now().UnixNano())
	firstPlayerID := player1ID
	if rand.Intn(2) == 1 {
		firstPlayerID = player2.ID
	}

	initialState := &domain.TicTacToeState{
		Board:       [3][3]string{},
		CurrentTurn: firstPlayerID,
	}
	stateBytes, err := json.Marshal(initialState)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal initial game state: %w", err)
	}

	game := &domain.Game{
		ID:          uuid.NewString(),
		Player1ID:   player1ID,
		Player2ID:   player2.ID,
		InitiatorID: player1ID,
		GameType:    "tic-tac-toe",
		Status:      domain.GamePending, // Waiting for player2 to accept
		State:       stateBytes,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := s.gameRepo.Create(ctx, game); err != nil {
		return nil, err
	}

	// Create an event for the invited player
	s.eventService.CreateEvent(ctx, player2.ID, domain.EventGameInvite, game)

	return game, nil
}

func (s *gameService) RespondToGameInvite(ctx context.Context, gameID, userID string, accept bool) (*domain.Game, error) {
	game, err := s.gameRepo.FindByID(ctx, gameID)
	if err != nil {
		return nil, ErrGameNotFound
	}
	if game.Status != domain.GamePending {
		return nil, ErrGameAlreadyResponded
	}
	if game.Player2ID != userID {
		return nil, errors.New("not authorized to respond to this invite")
	}

	if accept {
		game.Status = domain.GameActive
	} else {
		game.Status = domain.GameDeclined
	}
	game.UpdatedAt = time.Now().UTC()

	if err := s.gameRepo.Update(ctx, game); err != nil {
		return nil, err
	}

	// Create a game update event for both players
	s.eventService.CreateEvent(ctx, game.Player1ID, domain.EventGameUpdate, game)
	s.eventService.CreateEvent(ctx, game.Player2ID, domain.EventGameUpdate, game)

	return game, nil
}

func (s *gameService) GetGame(ctx context.Context, gameID string) (*domain.Game, error) {
	game, err := s.gameRepo.FindByID(ctx, gameID)
	if err != nil {
		return nil, ErrGameNotFound
	}
	return game, nil
}

func (s *gameService) MakeTicTacToeMove(ctx context.Context, gameID, playerID string, row, col int) (*domain.Game, error) {
	game, err := s.gameRepo.FindByID(ctx, gameID)
	if err != nil {
		return nil, ErrGameNotFound
	}
	if game.Status != domain.GameActive {
		return nil, ErrGameAlreadyFinished // Or pending/declined
	}
	if game.GameType != "tic-tac-toe" {
		return nil, errors.New("not a tic-tac-toe game")
	}

	if game.Player1ID != playerID && game.Player2ID != playerID {
		return nil, ErrNotGameParticipant
	}

	gameState, err := game.GetTicTacToeState()
	if err != nil {
		return nil, err
	}

	if gameState.CurrentTurn != playerID {
		return nil, ErrNotYourTurn
	}

	if row < 0 || row > 2 || col < 0 || col > 2 || gameState.Board[row][col] != "" {
		return nil, ErrInvalidMove
	}

	// Make the move
	playerMark := "X"
	if game.Player2ID == playerID {
		playerMark = "O"
	}
	gameState.Board[row][col] = playerMark
	gameState.LastMove = &struct {
		Row int `json:"row"`
		Col int `json:"col"`
	}{Row: row, Col: col}

	// Check for winner or draw
	winner := checkTicTacToeWinner(gameState.Board)
	if winner != "" {
		gameState.Winner = winner
		game.Status = domain.GameFinished
	} else if isBoardFull(gameState.Board) {
		gameState.Winner = "draw"
		game.Status = domain.GameFinished
	} else {
		// Switch turn
		if game.Player1ID == playerID {
			gameState.CurrentTurn = game.Player2ID
		} else {
			gameState.CurrentTurn = game.Player1ID
		}
	}

	if err := game.SetTicTacToeState(gameState); err != nil {
		return nil, err
	}
	game.UpdatedAt = time.Now().UTC()

	if err := s.gameRepo.Update(ctx, game); err != nil {
		return nil, err
	}

	// Create a game update event for both players
	s.eventService.CreateEvent(ctx, game.Player1ID, domain.EventGameUpdate, game)
	s.eventService.CreateEvent(ctx, game.Player2ID, domain.EventGameUpdate, game)

	return game, nil
}

func checkTicTacToeWinner(board [3][3]string) string {
	// Check rows
	for i := 0; i < 3; i++ {
		if board[i][0] != "" && board[i][0] == board[i][1] && board[i][1] == board[i][2] {
			return board[i][0]
		}
	}
	// Check columns
	for i := 0; i < 3; i++ {
		if board[0][i] != "" && board[0][i] == board[1][i] && board[1][i] == board[2][i] {
			return board[0][i]
		}
	}
	// Check diagonals
	if board[0][0] != "" && board[0][0] == board[1][1] && board[1][1] == board[2][2] {
		return board[0][0]
	}
	if board[0][2] != "" && board[0][2] == board[1][1] && board[1][1] == board[2][0] {
		return board[0][2]
	}
	return ""
}

func isBoardFull(board [3][3]string) bool {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if board[i][j] == "" {
				return false
			}
		}
	}
	return true
}
