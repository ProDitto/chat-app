package http_delivery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"real-time-chat/internal/domain"
	"real-time-chat/internal/usecase"

	"github.com/go-chi/chi/v5"
)

type FriendshipHandler struct {
	service usecase.FriendshipUseCase
}

func NewFriendshipHandler(service usecase.FriendshipUseCase) *FriendshipHandler {
	return &FriendshipHandler{service: service}
}

func (h *FriendshipHandler) SendRequest(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*domain.User)
	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.service.SendRequest(r.Context(), user.ID, req.Username)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	JSONResponse(w, http.StatusCreated, map[string]string{"message": "Friend request sent"})
}

func (h *FriendshipHandler) GetRequests(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*domain.User)
	requests, err := h.service.GetPendingRequests(r.Context(), user.ID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSONResponse(w, http.StatusOK, requests)
}

func (h *FriendshipHandler) RespondToRequest(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*domain.User)
	requestID := chi.URLParam(r, "requestID")
	var req struct {
		Status string `json:"status"` // "accepted" or "declined"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var status domain.FriendshipStatus
	switch req.Status {
	case "accepted":
		status = domain.Accepted
	case "declined":
		status = domain.Declined
	default:
		ErrorResponse(w, http.StatusBadRequest, "Invalid status")
		return
	}

	err := h.service.RespondToRequest(r.Context(), user.ID, requestID, status)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	JSONResponse(w, http.StatusOK, map[string]string{"message": fmt.Sprintf("Friend request %s", status)})
}

func (h *FriendshipHandler) GetFriends(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*domain.User)
	friends, err := h.service.GetFriends(r.Context(), user.ID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSONResponse(w, http.StatusOK, friends)
}
