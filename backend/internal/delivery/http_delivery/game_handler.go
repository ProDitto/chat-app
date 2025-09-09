package http_delivery

import (
	"encoding/json"
	"net/http"
	"real-time-chat/internal/delivery/ws_delivery" // For broadcasting game updates
	"real-time-chat/internal/domain"
	"real-time-chat/internal/usecase"

	"github.com/go-chi/chi/v5"
)

type GameHandler struct {
	gameService usecase.GameUseCase
	hub         *ws_delivery.Hub // To broadcast game updates
}

func NewGameHandler(gs usecase.GameUseCase, hub *ws_delivery.Hub) *GameHandler {
	return &GameHandler{gameService: gs, hub: hub}
}

type InviteToGameRequest struct {
	OpponentUsername string `json:"opponent_username"`
	GameType         string `json:"game_type"` // "tic-tac-toe"
}

func (h *GameHandler) InviteToGame(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*domain.User)
	var req InviteToGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.GameType != "tic-tac-toe" {
		ErrorResponse(w, http.StatusBadRequest, "Unsupported game type")
		return
	}

	game, err := h.gameService.InviteToTicTacToe(r.Context(), user.ID, req.OpponentUsername)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// The event service will now handle creating events, which the hub picks up for WS/LP
	// h.hub.BroadcastGameUpdate(game.Player2ID, "game_invite", game) // No longer needed directly here

	JSONResponse(w, http.StatusCreated, game)
}

func (h *GameHandler) RespondToGameInvite(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*domain.User)
	gameID := chi.URLParam(r, "gameID")

	var req struct {
		Accept bool `json:"accept"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	game, err := h.gameService.RespondToGameInvite(r.Context(), gameID, user.ID, req.Accept)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Event service now handles broadcasting game status updates
	// h.hub.BroadcastGameUpdate(game.Player1ID, "game_update", game)
	// h.hub.BroadcastGameUpdate(game.Player2ID, "game_update", game)

	JSONResponse(w, http.StatusOK, game)
}

func (h *GameHandler) GetGameState(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*domain.User)
	gameID := chi.URLParam(r, "gameID")

	game, err := h.gameService.GetGame(r.Context(), gameID)
	if err != nil {
		ErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	// Ensure the requesting user is a participant
	if game.Player1ID != user.ID && game.Player2ID != user.ID {
		ErrorResponse(w, http.StatusForbidden, "Not a participant of this game")
		return
	}

	JSONResponse(w, http.StatusOK, game)
}

type MakeMoveRequest struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

func (h *GameHandler) MakeMove(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*domain.User)
	gameID := chi.URLParam(r, "gameID")
	var req MakeMoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updatedGame, err := h.gameService.MakeTicTacToeMove(r.Context(), gameID, user.ID, req.Row, req.Col)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Event service now handles broadcasting game updates
	// h.hub.BroadcastGameUpdate(updatedGame.Player1ID, "game_update", updatedGame)
	// h.hub.BroadcastGameUpdate(updatedGame.Player2ID, "game_update", updatedGame)

	JSONResponse(w, http.StatusOK, updatedGame)
}
