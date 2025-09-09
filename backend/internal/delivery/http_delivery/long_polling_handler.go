package http_delivery

import (
	"context"
	"net/http"
	"real-time-chat/internal/services"
	"real-time-chat/internal/usecase"
	"strconv"
	"strings"
	"time"
)

type LongPollingHandler struct {
	eventService usecase.EventUseCase
	userService  usecase.UserUseCase
}

func NewLongPollingHandler(es usecase.EventUseCase, us usecase.UserUseCase) *LongPollingHandler {
	return &LongPollingHandler{eventService: es, userService: us}
}

// HandleLongPolling retrieves events for a user after a given event ID.
func (h *LongPollingHandler) HandleLongPolling(w http.ResponseWriter, r *http.Request) {
	// Authenticate the user for long polling, similar to middleware
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		ErrorResponse(w, http.StatusUnauthorized, "Authorization header required")
		return
	}
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
		ErrorResponse(w, http.StatusUnauthorized, "Invalid Authorization header format")
		return
	}
	accessToken := tokenParts[1]

	// Temporarily create a dummy token service or pass it in.
	// For simplicity, let's use a simplified validation that's good enough for this context.
	claims, err := usecase.TokenUseCase.ValidateToken(h.userService.(*services.UserService).TokenService, accessToken) // Access tokenService via userService
	if err != nil {
		ErrorResponse(w, http.StatusUnauthorized, "Invalid token")
		return
	}
	user, err := h.userService.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		ErrorResponse(w, http.StatusUnauthorized, "User not found")
		return
	}

	// Get query parameters
	sinceEventID := r.URL.Query().Get("since")
	limitStr := r.URL.Query().Get("limit")
	limit, _ := strconv.Atoi(limitStr)
	if limit == 0 {
		limit = 50 // Default events per poll
	}

	// Set headers for long polling
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ctx, cancel := context.WithTimeout(r.Context(), 59*time.Second) // Long polling timeout (less than client's 60s)
	defer cancel()

	// Loop to poll for new events
	for {
		events, err := h.eventService.GetEventsForUser(ctx, user.ID, sinceEventID, limit)
		if err != nil {
			ErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve events")
			return
		}

		if len(events) > 0 {
			// If events found, send them immediately
			JSONResponse(w, http.StatusOK, events)
			return
		}

		select {
		case <-ctx.Done():
			// Timeout or client disconnected, send empty array
			JSONResponse(w, http.StatusOK, []interface{}{})
			return
		case <-time.After(1 * time.Second): // Polling interval on server side
			// Continue to next iteration after a short delay
			// This prevents constant database hammering if no events are present.
		}
	}
}
