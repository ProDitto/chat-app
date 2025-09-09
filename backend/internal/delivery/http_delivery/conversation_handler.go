package http_delivery

import (
	"net/http"
	"real-time-chat/internal/domain"
	"real-time-chat/internal/usecase"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type ConversationHandler struct {
	convoService   usecase.ConversationUseCase
	messageService usecase.MessageUseCase
}

func NewConversationHandler(cs usecase.ConversationUseCase, ms usecase.MessageUseCase) *ConversationHandler {
	return &ConversationHandler{convoService: cs, messageService: ms}
}

func (h *ConversationHandler) GetUserConversations(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*domain.User)
	convos, err := h.convoService.GetUserConversations(r.Context(), user.ID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSONResponse(w, http.StatusOK, convos)
}

func (h *ConversationHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*domain.User)
	conversationID := chi.URLParam(r, "conversationID")

	cursorStr := r.URL.Query().Get("before")
	var before time.Time
	if cursorStr != "" {
		ts, err := time.Parse(time.RFC3339Nano, cursorStr)
		if err != nil {
			ErrorResponse(w, http.StatusBadRequest, "Invalid 'before' timestamp format")
			return
		}
		before = ts
	} else {
		before = time.Now().UTC()
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}

	messages, err := h.messageService.GetMessagesForConversation(r.Context(), conversationID, user.ID, before, limit)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSONResponse(w, http.StatusOK, messages)
}

func (h *ConversationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*domain.User)
	conversationID := chi.URLParam(r, "conversationID")

	err := h.convoService.MarkConversationAsRead(r.Context(), conversationID, user.ID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSONResponse(w, http.StatusOK, map[string]string{"message": "success"})
}

func (h *ConversationHandler) DeleteOneToOneConversation(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*domain.User)
	conversationID := chi.URLParam(r, "conversationID")

	err := h.convoService.DeleteOneToOneConversation(r.Context(), conversationID, user.ID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSONResponse(w, http.StatusOK, map[string]string{"message": "Conversation deleted for user"})
}
