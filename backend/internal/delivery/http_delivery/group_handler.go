package http_delivery

import (
	"encoding/json"
	"net/http"
	"real-time-chat/internal/domain"
	"real-time-chat/internal/usecase"

	"github.com/go-chi/chi/v5"
)

type GroupHandler struct {
	groupService usecase.GroupUseCase
	convoService usecase.ConversationUseCase
}

func NewGroupHandler(gs usecase.GroupUseCase, cs usecase.ConversationUseCase) *GroupHandler {
	return &GroupHandler{groupService: gs, convoService: cs}
}

type CreateGroupRequest struct {
	Name           string   `json:"name"`
	Slug           string   `json:"slug"`
	InitialMembers []string `json:"initial_members"` // User IDs
}

func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*domain.User)
	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	group, err := h.groupService.CreateGroup(r.Context(), user.ID, req.Name, req.Slug, req.InitialMembers)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	JSONResponse(w, http.StatusCreated, group)
}

func (h *GroupHandler) GetGroupDetails(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*domain.User)
	groupID := chi.URLParam(r, "groupID")

	// Ensure the requesting user is a member of the group
	isMember, err := h.convoService.IsUserInConversation(r.Context(), groupID, user.ID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !isMember {
		ErrorResponse(w, http.StatusForbidden, "User is not a member of this group")
		return
	}

	group, members, err := h.groupService.GetGroupDetails(r.Context(), groupID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"group":   group,
		"members": members,
	})
}

func (h *GroupHandler) JoinGroup(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*domain.User)
	groupID := chi.URLParam(r, "groupID")

	err := h.groupService.JoinGroup(r.Context(), groupID, user.ID)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	JSONResponse(w, http.StatusOK, map[string]string{"message": "Joined group successfully"})
}

func (h *GroupHandler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*domain.User)
	groupID := chi.URLParam(r, "groupID")

	err := h.groupService.LeaveGroup(r.Context(), groupID, user.ID)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
		// Handle specific errors like ErrMinGroupMembers for better feedback
	}
	JSONResponse(w, http.StatusOK, map[string]string{"message": "Left group successfully"})
}

func (h *GroupHandler) RemoveGroupMember(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*domain.User)
	groupID := chi.URLParam(r, "groupID")
	memberToRemoveID := chi.URLParam(r, "memberID")

	err := h.groupService.RemoveGroupMember(r.Context(), user.ID, groupID, memberToRemoveID)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	JSONResponse(w, http.StatusOK, map[string]string{"message": "Member removed successfully"})
}

func (h *GroupHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*domain.User)
	groupID := chi.URLParam(r, "groupID")

	err := h.groupService.DeleteGroup(r.Context(), user.ID, groupID)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	JSONResponse(w, http.StatusOK, map[string]string{"message": "Group deleted successfully"})
}
